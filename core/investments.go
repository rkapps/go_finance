package core

import (
	"context"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/rkapps/go_finance/store"
)

//InvestmentsHoldings returns the current holdings
func (fn *Finance) InvestmentsHoldings(ctx context.Context, group string, category string, byAcct bool) (store.InvHoldings, error) {

	lot := store.InvLot{}
	lot.Group = group
	lot.Category = category

	lots := fn.getLots(ctx, group, category, "", "", true, true, false)
	log.Printf("InvestmentsLots: %v", len(lots))

	hs := store.InvHoldings{}
	hm := make(map[string]*store.InvHolding)
	var key string
	for _, lot := range lots {

		if byAcct {
			key = lot.Group + lot.Category + lot.Account + lot.Symbol
		} else {
			key = lot.Group + lot.Category + lot.Symbol
		}
		h := hm[key]

		if h == nil {
			h = &store.InvHolding{}
			h.Group = lot.Group
			h.Category = lot.Category
			h.Account = lot.Account
			h.Symbol = lot.Symbol
			h.Qty = 0
			h.Cost = 0
			h.CostValue = 0
			h.MktValue = 0
			hm[key] = h
			hs = append(hs, h)

		} else {

		}

		h.PrLast = lot.PrLast
		h.PrDiffAmt = lot.PrDiffAmt
		h.PrDiffPerc = lot.PrDiffPerc
		h.Qty += lot.Qty
		h.CostValue += lot.CostValue
		h.Cost = h.CostValue / h.Qty
		h.MktValue += lot.MktValue
		h.Dglamount += lot.Dglamount
		h.Glamount += lot.Glamount
		h.Glperc = h.Glamount * 100 / h.CostValue

	}

	return hs, nil
}

//InvestmentsGainLoss returns all sales lots for the date range
func (fn *Finance) InvestmentsGainLoss(ctx context.Context, group string, category string, ft time.Time, et time.Time) store.InvLots {

	// var rlots store.InvLots
	log.Printf("InvestmentsGainLoss - Group: %s Category: %s Dates = %v to: %v", group, category, ft, et)

	lot := &store.InvLot{}
	lot.Group = group
	lot.Category = category

	lots := fn.MDB.InvestmentsSaleLots(ctx, lot, ft, et)
	// log.Println(len(lots))

	fn.setLots(ctx, lots)
	return lots
}

//InvestmentsGainLoss returns all sales lots for the date range
func (fn *Finance) InvestmentsRewards(ctx context.Context, group string, category string, open bool, ft time.Time, et time.Time) store.InvLots {

	// var rlots store.InvLots
	lot := &store.InvLot{}
	lot.Group = group
	lot.Category = category

	lots := fn.MDB.InvestmentsRewardsLots(ctx, lot, open, ft, et)
	// fn.setLots(ctx, lots)
	for _, lot = range lots {
		if lot.Status == "O" {
			lot.CostValue = lot.Qty * lot.Cost
		} else {
			lot.CostValue = lot.OrigQty * lot.Cost
		}
	}
	return lots
}

//InvestmentsLots returns lots
func (fn *Finance) InvestmentsLots(ctx context.Context, group string, category string, symbol string, status string) store.InvLots {

	var lots store.InvLots
	lot := &store.InvLot{}
	lot.Group = group
	lot.Category = category
	lot.Symbol = symbol

	ft := time.Time{}
	et := time.Time{}
	if strings.Compare(status, "sales") == 0 {
		lots = fn.MDB.InvestmentsSaleLots(ctx, lot, ft, et)
	} else {
		open := false
		if strings.Compare(status, "open") == 0 {
			open = true
		}
		lots = fn.MDB.InvestmentsLots(ctx, lot, open, true)
	}

	fn.setLots(ctx, lots)
	return lots
}

//InvestmentsGainLoss returns all sales lots for the date range
func (fn *Finance) getLots(ctx context.Context, group string, category string, account string, symbol string, open bool, asc bool, hifo bool) store.InvLots {

	// log.Printf("InvestmentsLots - group: %s category: %s symbol:%s open: %v", group, category, symbol, open)

	lot := &store.InvLot{}
	lot.Group = group
	lot.Category = category
	lot.Account = account
	lot.Symbol = symbol

	lots := fn.MDB.InvestmentsLots(ctx, lot, open, asc)
	fn.setLots(ctx, lots)

	if hifo {
		//Sort by date ascending
		sort.SliceStable(lots, func(i, j int) bool {
			return lots[i].Cost > (*&lots[j].Cost)
		})
	}

	// log.Printf("InvestmentsLots - lots: %d", len(ls))
	return lots
}

//InvestmentsGainLoss returns all sales lots for the date range
func (fn *Finance) setLots(ctx context.Context, lots store.InvLots) {

	//Get the tickers price and add them to the map
	var symbols []string
	var lot *store.InvLot

	for _, lot = range lots {
		symbols = append(symbols, lot.Symbol)
	}

	tm := make(map[string]*store.Ticker)
	// log.Printf("Holdings: %d", len(hs))
	ts := fn.MDB.GetTickers(ctx, symbols)
	for _, t := range ts {
		tm[t.Symbol] = t
	}

	var ticker *store.Ticker
	for _, lot = range lots {

		// if lot.Qty <= 0.001 {
		// 	continue
		// }

		// if strings.Compare(lot.Symbol, "BTC-USD") == 0 {
		// 	log.Printf("Lots - Account: %s Symbol: %s Status: %v Qty: %f Cost: %f CostValue: %f", lot.Account, lot.Symbol, lot.Status, lot.Qty, lot.Cost, lot.CostValue)
		// }

		// if strings.Compare(lot.Symbol, "BTC-USD") == 0 &&
		// 	strings.Compare(lot.Account, "Atomic") == 0 {
		// 	log.Printf("Lots - Account: %s Symbol: %s Status: %v Qty: %f Cost: %f CostValue: %f", lot.Account, lot.Symbol, lot.Status, lot.Qty, lot.Cost, lot.CostValue)
		// }

		if strings.Compare(lot.Symbol, "ETH2-USD") == 0 ||
			strings.Compare(lot.Symbol, "WETH-USD") == 0 {
			ticker = tm["ETH-USD"]
		} else {
			ticker = tm[lot.Symbol]
		}

		if ticker == nil {
			log.Printf("Ticker: %s not found", lot.Symbol)
			// continue
		} else {

			lot.PrLast = ticker.PrLast
			lot.PrDiffAmt = ticker.PrDiffAmt
			lot.PrDiffPerc = ticker.PrDiffPerc

		}

		if strings.Compare(lot.Status, "O") == 0 {
			lot.CostValue = lot.Qty * lot.Cost
			lot.MktValue = lot.Qty * lot.PrLast
			lot.Dglamount = lot.Qty * lot.PrDiffAmt
		} else if lot.SaleQty > 0 {
			lot.CostValue = lot.SaleQty * lot.Cost
			lot.MktValue = lot.SaleQty * lot.SalePrice
		}
		lot.Glamount = lot.MktValue - lot.CostValue
		if lot.CostValue != 0 {
			lot.Glperc = lot.Glamount * 100 / lot.CostValue
		}

	}

}
