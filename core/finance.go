package core

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/rkapps/go_finance/store"
	"github.com/shopspring/decimal"
)

//Finance defines the main struct
type Finance struct {
	MDB *store.MongoDB
}

//NewFinance creates new Finance
func NewFinance(mongoConnStr string) (*Finance, error) {

	mdb, err := store.NewMongoDB(mongoConnStr)
	if err != nil {
		return nil, err
	}
	return &Finance{MDB: mdb}, nil
}

func (fn *Finance) ActivitiesImport(ctx context.Context, actyType string, group string, category string, fromDate *time.Time, toDate *time.Time, actvs store.Activities) error {

	// fn.MDB.DropActivitiesCollection(ctx)

	fn.MDB.DeleteActivities(ctx, actyType, group, category, fromDate, toDate)
	if strings.Compare("Investment", actyType) == 0 {
		fn.MDB.DeleteInvlots(ctx, group, category, fromDate, toDate)
	}

	// return nil

	var uptxns store.Activities

	//Sort by date ascending
	sort.SliceStable(actvs, func(i, j int) bool {
		return actvs[i].Date.Before(*actvs[j].Date)
	})

	for _, actv := range actvs {

		var upactvs store.Activities
		var ulots store.InvLots
		var qty = actv.Qty
		var hifo bool = false
		var asc = true

		// //LIFO logic
		if actv.Date.Year() > 2020 {
			// asc = false
			hifo = true
		}

		// var update = true
		// fmt.Printf("date: %v\n", actv.ActyType)
		if strings.Compare("Investment", actv.ActyType) == 0 {

			upactvs = append(upactvs, actv)
			fn.MDB.ActivitiesUpdate(ctx, upactvs)

			if strings.Compare("Buy", actv.TxnType) == 0 ||
				strings.Compare("Rewards", actv.TxnType) == 0 {
				// actv.OrigQty = actv.Qty

				lot := &store.InvLot{}
				lot.ActvID = actv.ID
				lot.Group = actv.Group
				lot.Category = actv.Category
				lot.Account = actv.Account
				lot.Symbol = actv.Symbol
				lot.Date = actv.Date
				lot.TxnType = actv.TxnType
				lot.TxnDate = actv.Date
				lot.Status = "O"
				lot.OrigQty = actv.Qty
				lot.Qty = actv.Qty
				lot.Cost = actv.Price
				// lot.CostValue = lot.Qty * lot.Cost
				lot.Fee = actv.Fee

				ulots = append(ulots, lot)

			} else if strings.Compare("Send", actv.TxnType) == 0 ||
				strings.Compare("Sale", actv.TxnType) == 0 {

				//Get the open lots
				lots := fn.getLots(ctx, actv.Group, actv.Category, actv.Account, actv.Symbol, true, asc, hifo)
				if len(lots) == 0 {
					continue
				}

				// log.Printf("Send To: %s\n", actv.ToMerchant)
				// log.Printf("Account: %s TxnType: %s  Qty: %f \n", actv.Account, actv.TxnType, qty)

				for _, lot := range lots {

					if lot.Status == "C" || lot.Qty == 0 {
						continue
					}

					// log.Printf("    Orig--- Date: %v Qty: %v Price: %f \n", lot.Date, lot.Qty, lot.Cost)
					lqty := qty

					if qty > lot.Qty {
						lqty = lot.Qty
					}

					dqty := decimal.NewFromFloat(qty)
					sqty := decimal.NewFromFloat(lqty)

					qty, _ = dqty.Sub(sqty).Float64()
					// log.Printf("    Lot Qty: %f Remaining Qty: %f", lqty, qty)

					lot.Qty = lot.Qty - lqty

					slot := &store.InvLot{}
					slot.ActvID = lot.ActvID
					slot.Group = lot.Group
					slot.Category = lot.Category
					slot.Account = lot.Account
					slot.Symbol = lot.Symbol
					slot.Date = lot.Date
					slot.TxnType = lot.TxnType
					slot.TxnDate = actv.Date
					slot.OrigQty = lot.OrigQty
					slot.Qty = lqty
					slot.Cost = lot.Cost
					// slot.CostValue = slot.Qty * slot.Cost
					slot.Status = "O"
					slot.Fee = actv.Fee

					if strings.Compare("Sale", actv.TxnType) == 0 {

						if lot.Qty == 0 {

							lot.Status = "C"
							lot.TxnDate = actv.Date
							lot.SaleDate = actv.Date
							lot.SaleQty = lqty
							lot.SalePrice = actv.Price

						} else {
							slot.Status = "C"
							slot.Qty = 0
							slot.SaleDate = actv.Date
							slot.SaleQty = lqty
							slot.SalePrice = actv.Price
							ulots = append(ulots, slot)
						}

					} else {

						lot.SendQty = lot.SendQty + lqty
						lot.SendDate = actv.Date
						lot.TxnDate = actv.Date

						if lot.Qty == 0 {

							lot.Status = "C"
							// lot.Account = actv.ToAccount
						}

						slot.Account = actv.ToAccount
						slot.OrigQty = slot.Qty
						// slot.SendDate = actv.Date
						// slot.SendQty = lqty
						slot.OrigAccount = lot.OrigAccount + ":" + lot.Account
						ulots = append(ulots, slot)
					}

					ulots = append(ulots, lot)

					if qty <= 0 {
						break
					}

				}

			}

			// log.Printf("Activities county: %d\n", len(upactvs))
			fn.MDB.InvLotsUpdate(ctx, ulots)

		} else if strings.Compare("Transaction", actv.ActyType) == 0 {
			uptxns = append(uptxns, actv)
		}

	}

	if len(uptxns) > 0 {
		fn.MDB.ActivitiesUpdate(ctx, uptxns)
	}

	return nil

}
