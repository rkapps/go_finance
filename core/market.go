package core

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/rkapps/go_finance/providers"
	"github.com/rkapps/go_finance/store"
	"github.com/rkapps/go_finance/utils"
)

func (fn *Finance) AddTicker(ctx context.Context, t *store.Ticker) error {

	keys := []string{strings.ToLower(t.Symbol), strings.ToLower(t.Exchange), strings.ToLower(t.Name), strings.ToLower(t.Sector), strings.ToLower(t.Industry)}
	keywords := utils.CreateTickerKeywords(keys)

	return fn.MDB.AddTicker(ctx, t, keywords)
}

//LoadTickers add tickers to the database
func (fn *Finance) LoadTickers(ctx context.Context, tickers []store.Ticker) {

	var uts store.Tickers
	var t *store.Ticker

	log.Println(len(tickers))

	for _, ticker := range tickers {

		log.Println(ticker.Symbol)
		if len(ticker.Symbol) == 0 {
			continue
		}
		t = fn.MDB.GetTicker(ctx, ticker.Symbol)
		if t == nil {
			t = &store.Ticker{}
		}

		t.Exchange = ticker.Exchange
		t.Symbol = ticker.Symbol
		t.Name = ticker.Name
		t.Sector = ticker.Sector
		t.Industry = ticker.Industry
		t.DivAmt = ticker.DivAmt
		t.Yield = ticker.Yield
		t.Overview = ticker.Overview

		// log.Printf("Ticker: %v", ticker)

		uts = append(uts, t)
	}

	log.Printf("Loading tickers %d", len(uts))
	if len(uts) > 0 {
		fn.UpdateTickers(ctx, uts, true)
	}
}

//DeleteTicker deletes ticker data
func (fn *Finance) DeleteTicker(ctx context.Context, symbol string) error {
	return fn.MDB.DeleteTicker(ctx, symbol)
}

//GetTickers returns tickers
func (fn *Finance) GetTickers(ctx context.Context, symbols []string) []*store.Ticker {
	return fn.MDB.GetTickers(ctx, symbols)
}

//GetTickerHistory returns history for the ticker
func (fn *Finance) GetTickerHistory(ctx context.Context, symbol string) []*store.TickerHistory {
	return fn.MDB.GetTickerHistory(ctx, symbol)
}

//GetTickerNews returns news for the ticker
func (fn *Finance) GetTickerNews(ctx context.Context, symbol string) []*store.TickerNews {
	return fn.MDB.GetTickerNews(ctx, symbol)
}

//SearchTickers searches tickers
func (fn *Finance) SearchTickers(ctx context.Context, ts store.TickerSearch) store.Tickers {
	return fn.MDB.SearchTickers(ctx, ts)
}

func (fn *Finance) UpdateTickers(ctx context.Context, tickers store.Tickers, updateHistory bool) {

	if len(tickers) == 0 {
		tickers = fn.MDB.GetTickers(ctx, nil)
	}

	// log.Println("Updating tickers started...")
	for _, ticker := range tickers {

		if ticker.IsStock() {
			url, err := providers.UpdateTickerDetails(ticker)
			time.Sleep(1000 * time.Millisecond)

			if err != nil {
				log.Println(url)
				log.Println(err)
				// log.Println(ticker.FormatTickerData(";"))
				// continue
			}
		}

		err := fn.AddTicker(ctx, ticker)
		if err != nil {
			log.Printf("UpdateTickers: Ticker: %s  Error: %s", ticker.Symbol, err)
		} else {

			// if i+1%100 == 0 {
			// 	log.Printf("Updating %d of %d", i+1, len(tickers))
			// }
			var uts store.Tickers
			uts = append(uts, ticker)
			fn.updateEOD(ctx, uts, updateHistory, true)
		}

	}

	// log.Println("Updating tickers done.")

}

//UpdateEODStocks updates all tickers EOD
func (fn *Finance) UpdateStocksEOD(ctx context.Context) {
	// print("Reached")
	tickers := fn.MDB.GetTickersByExchange(ctx, []string{store.ExNasdaq, store.ExNyse, store.ExNyseArca, store.ExIndex, store.ExOtc})
	if tickers != nil {
		log.Printf("Updating %d tickers.", len(tickers))
		fn.updateEOD(ctx, tickers, true, false)
	} else {
		log.Printf("No tickers to update.")
	}

}

//UpdateEODStocks updates all tickers EOD
func (fn *Finance) UpdateCryptosEOD(ctx context.Context) {
	// print("Reached")
	tickers := fn.MDB.GetTickersByExchange(ctx, []string{store.ExCurrency})
	if tickers != nil {
		log.Printf("Updating %d crypto tickers.", len(tickers))
		fn.updateEOD(ctx, tickers, true, false)
	} else {
		log.Printf("No crypto tickers to update.")
	}

}

func (fn *Finance) updateEOD(ctx context.Context, tickers store.Tickers, updateHistory bool, all bool) {

	var uts store.Tickers
	var uths []*store.TickerHistory
	var th *store.TickerHistory
	var pth *store.TickerHistory

	//update ticker data from barchat
	// m.bc.UpdateTickerData(&tickers)

	//update ticker fundamentals from tiingo
	// m.tg.UpdateTickerFundamentals(&tickers)

	// log.Println("Before update ticker history")
	// thm := m.updateTickersHistory(ctx, tickers)

	for _, ticker := range tickers {

		// tha := thm[ticker.GetTickerID()]
		tha := fn.updateTickerHistory(ctx, ticker)

		// log.Printf("Ticker: %s hist: %d\n", ticker.Symbol, len(tha))
		if len(tha) == 0 {
			log.Printf("Ticker: %s history not available.", ticker.GetTickerID())
			// continue
			prLast, err := providers.GetTickerPriceFromBinance(ticker.Symbol)
			if err != nil {
				continue
			}
			// log.Printf("Ticker price from binance %f", prLast)
			ticker.PrLast = prLast
			ticker.PrClose = prLast
			if ticker.PrPrev != 0.0 {
				ticker.SetPriceDiff()
			}
			ticker.PrPrev = ticker.PrLast

		} else {

			pth = nil
			th = tha[len(tha)-1]
			if len(tha) > 1 {
				pth = tha[len(tha)-2]
			}

			ticker.Technicals = make(map[string]map[string]float64)
			ticker.Technicals[store.SMA] = th.SMA
			ticker.Technicals[store.EMA] = th.EMA
			ticker.Technicals[store.RSI] = th.RSI
			ticker.PrDate = &th.Date
			ticker.PrLast = th.Close
			ticker.PrOpen = th.Open
			ticker.PrHigh = th.High
			ticker.PrLow = th.Low
			ticker.PrClose = th.Close
			if pth != nil {
				ticker.PrPrev = pth.Close
			}
			ticker.SetPriceDiff()
			fn.updatePerformance(ctx, ticker, tha)

		}

		uts = append(uts, ticker)

		if all {
			for _, th = range tha {
				uths = append(uths, th)
			}
		} else {
			uths = append(uths, th)
		}

	}

	// log.Println("Before update tickers history")
	if len(uts) > 0 {
		// log.Println("Updating tickers...")
		fn.MDB.UpdateTickersEOD(ctx, uts)
		if updateHistory {
			// log.Println("Updating tickers history...")
			fn.MDB.UpdateTickersHistory(ctx, uths)

		}
	}
	// log.Printf("Total tickers: %d --- Updated tickers: %d", len(tickers), len(uts))
}

//UpdateEODStocks updates all tickers EOD
func (fn *Finance) UpdateStocksRealtime(ctx context.Context) {
	// print("Reached")
	tickers := fn.MDB.GetTickersByExchange(ctx, []string{store.ExNasdaq, store.ExNyse, store.ExNyseArca, store.ExIndex, store.ExOtc})
	if tickers != nil {
		log.Printf("Updating %d tickers.", len(tickers))
		fn.updateRealtime(ctx, tickers)
	} else {
		log.Printf("No tickers to update.")
	}

}

//UpdateEODStocks updates all tickers EOD
func (fn *Finance) UpdateCryptosRealtime(ctx context.Context) {
	// print("Reached")
	tickers := fn.MDB.GetTickersByExchange(ctx, []string{store.ExCurrency})
	if tickers != nil {
		log.Printf("Updating %d crypto tickers.", len(tickers))
		fn.updateRealtime(ctx, tickers)
	} else {
		log.Printf("No crypto tickers to update.")
	}

}

func (fn *Finance) updateRealtime(ctx context.Context, tickers store.Tickers) {

	var uts store.Tickers
	today := time.Now()

	tom := time.Now().Add(time.Hour * 48)
	ctm := providers.GetCryptoHistory(tickers, today, tom)
	itm, ptm := providers.GetTickerQuotes(tickers)

	// var debugSymbols string
	// debugSymbols = "AAPL,ZG"
	// debugSymbols = m.db.GetDebugTickers(ctx)

	// uts := m.tg.UpdateTickerRealTimeQuotes(&tickers)
	for _, ticker := range tickers {

		// if strings.Contains(debugSymbols, ticker.Symbol) {
		// 	log.Printf("Before --- %v", ticker.PrLast)
		// } else {
		// 	continue
		// }

		if ticker.IsStock() {

			lp, date := providers.GetTickerRealTimeQuote(*ticker)
			if date != nil && utils.DateEqual(today, *date) {

				if ticker.PrDate == nil || !utils.DateEqual(*date, *ticker.PrDate) {
					ticker.PrDate = date
					ticker.PrPrev = ticker.PrLast
				}
				// log.Printf("lastprice: %f", lp)
				ticker.PrLast = formatDec(lp)
				ticker.SetPriceDiff()
				// if strings.Contains(debugSymbols, ticker.Symbol) {
				// 	log.Printf("Stock --- %v", ticker.PrLast)
				// }
				uts = append(uts, ticker)
			}

		} else if ticker.IsCrypto() {

			tha := ctm[ticker.GetTickerID()]
			if tha == nil || len(tha) == 0 {

				prLast, err := providers.GetTickerPriceFromBinance(ticker.Symbol)
				if err != nil {
					continue
				}
				// log.Printf("Binance price - %s: %f", ticker.Symbol, prLast)
				ticker.PrLast = prLast
				if ticker.PrPrev != 0.0 {
					ticker.SetPriceDiff()
				}

			} else {
				// if tha != nil && len(tha) > 0 {

				th := tha[len(tha)-1]
				// log.Println(th)
				ticker.PrLast = formatDec(th.Close)
				ticker.PrDate = &th.Date
				ticker.SetPriceDiff()
				// if strings.Contains(debugSymbols, ticker.Symbol) {
				// log.Printf("Crypto --- %s", ticker.FormatPriceDiff())
				// }
				uts = append(uts, ticker)
			}
		} else if ticker.IsIndex() {

			ih := itm[ticker.GetTickerID()]
			if ih != nil {
				// log.Printf("Date: %v close: %f", ih.Date, ih.Close)
				// if ticker.PrDate == nil || !utils.DateEqual(*date, *ticker.PrDate) {
				// 	ticker.PrDate = date
				// 	ticker.PrPrev = ticker.PrLast
				// }
				ph := ptm[ticker.GetTickerID()]
				if ph != nil {
					ticker.PrDate = &ph.Date
					ticker.PrPrev = ph.Close
				}

				ticker.PrLast = formatDec(ih.Close)
				ticker.PrDate = &ih.Date
				ticker.SetPriceDiff()
				// if strings.Contains(debugSymbols, ticker.Symbol) {
				// log.Printf("Index --- %s", ticker.FormatPriceDiff())
				// }
				uts = append(uts, ticker)

			}
		}

	}

	if len(uts) > 0 {
		fn.MDB.UpdateTickersRealtime(ctx, uts)
		// log.Printf("Total tickers updated: %d", len(uts))
	}

}

func (fn *Finance) updatePerformance(ctx context.Context, ticker *store.Ticker, tha []*store.TickerHistory) {

	//var thm map[string]*core.TickerHistory
	thm := make(map[string]*store.TickerHistory)
	var th *store.TickerHistory
	for _, th := range tha {
		id := ticker.Exchange + ":" + ticker.Symbol + ":" + utils.DateFormat1(th.Date)
		thm[id] = th
	}
	// println(ticker.Symbol)
	ticker.Performance = make(map[string]map[string]float64)
	for _, period := range store.PerfPeriods {
		date := utils.DateForPeriod(period)

		for x := 0; x < 5; x++ {
			ndate := date.Add(-time.Hour * time.Duration(24*x))
			id := ticker.Exchange + ":" + ticker.Symbol + ":" + utils.DateFormat1(ndate)
			th = thm[id]
			if th != nil {
				break
			}
		}

		ticker.Performance[period] = make(map[string]float64)
		if th == nil {
			ticker.Performance[period]["price"] = 0
			ticker.Performance[period]["diff"] = 0
		} else {
			ticker.Performance[period]["price"] = th.Close
			_, ticker.Performance[period]["diff"] = utils.PriceDiff(ticker.PrLast, th.Close)
		}
		// utils.PriceDiff(ticker.PrLast, th.Close)
	}
}

func (fn *Finance) updateTickerHistory(ctx context.Context, ticker *store.Ticker) []*store.TickerHistory {

	var thm = make(map[string][]*store.TickerHistory)
	var tha []*store.TickerHistory

	et := time.Now()
	st := time.Now().Add(-time.Hour * 24 * 365 * 6)

	if ticker.IsCrypto() {
		thm = providers.GetCryptoHistory(store.Tickers{ticker}, st, et)
	} else {
		thm = providers.GetTickersHistory(store.Tickers{ticker}, st, et)
	}
	tha = thm[ticker.GetTickerID()]

	// log.Println(len(tha))
	for _, th := range tha {
		th.Close = formatDec(th.Close)
		th.Open = formatDec(th.Open)
		th.High = formatDec(th.High)
		th.Low = formatDec(th.Low)
	}

	updateRSI(tha)
	updateMAs(tha)
	return tha
}

func (fn *Finance) UpdateTickersNews(ctx context.Context) {

	tickers := fn.MDB.GetTickers(ctx, []string{})

	tnm := providers.GetTickerNews(tickers)
	fn.MDB.UpdateTickersNews(ctx, tnm)
}
