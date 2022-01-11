package utils

// import (
// 	"fmt"
// 	"sort"
// 	"strings"

// 	"github.com/rkapps/go_finance/store"
// )

// func WriteTaxFile(invTax store.InvTax, salesFileName string, rewardsFileName string) {

// 	writeTaxSales(invTax.Sales, salesFileName)
// 	writeTaxRewards(invTax.Rewards, rewardsFileName)
// }

// func writeTaxSales(sales []*store.InvLot, fileName string) {

// 	var lines []string
// 	lasm := make(map[string][]store.InvLot)
// 	sasm := make(map[string][]store.InvLot)
// 	var lkeys []string
// 	var skeys []string

// 	for _, sale := range sales {
// 		key := sale.Account + sale.Symbol
// 		if sale.SaleDate.Sub(*sale.Date).Hours()/24 >= 364 {

// 			as := lasm[key]
// 			if as == nil {
// 				as = []store.InvLot{}
// 				lasm[key] = as
// 				lkeys = append(lkeys, key)
// 			}

// 			lasm[key] = append(lasm[key], *sale)

// 		} else {

// 			as := sasm[key]
// 			if as == nil {
// 				as = []store.InvLot{}
// 				sasm[key] = as
// 				skeys = append(skeys, key)
// 			}

// 			sasm[key] = append(sasm[key], *sale)

// 		}
// 		// log.Printf("lenth: %d", len(asm[key]))
// 		// break
// 	}

// 	lines = append(lines, fmt.Sprint(",Account,Symbol,Qty,Date,Cost,Cost Value,Sale Date,Sale Price,Sale Value,GL"))
// 	lines = append(lines, "")
// 	lines1 := writeTaxLines("Long Term", lkeys, lasm)
// 	lines = append(lines, lines1...)
// 	lines = append(lines, "")

// 	lines1 = writeTaxLines("Short Term", skeys, sasm)
// 	lines = append(lines, lines1...)

// 	WriteLinesToFile(fileName, lines)

// }

// func writeTaxLines(term string, keys []string, asm map[string][]store.InvLot) []string {

// 	var lines []string

// 	lines = append(lines, term)

// 	sort.Strings(keys)

// 	var costTotal, saleTotal, costSubTotal, saleSubTotal float64

// 	for _, key := range keys {
// 		sales := asm[key]
// 		// log.Printf("key: %s sales: %d\n", key, len(sales))

// 		sort.SliceStable(sales, func(i int, j int) bool {
// 			return sales[i].SaleDate.After(*sales[j].SaleDate)
// 		})

// 		costSubTotal = 0.0
// 		saleSubTotal = 0.0

// 		for _, sale := range sales {
// 			var entries []string
// 			entries = append(entries, "")
// 			entries = append(entries, sale.Account)
// 			entries = append(entries, sale.Symbol)
// 			entries = append(entries, fmt.Sprintf("%f", sale.Qty))
// 			entries = append(entries, sale.Date.Format("2006-01-02"))
// 			entries = append(entries, fmt.Sprintf("%f", sale.Cost))
// 			entries = append(entries, fmt.Sprintf("%f", sale.CostValue))
// 			entries = append(entries, sale.SaleDate.Format("2006-01-02"))
// 			entries = append(entries, fmt.Sprintf("%f", sale.SalePrice))
// 			entries = append(entries, fmt.Sprintf("%f", sale.SaleValue))
// 			entries = append(entries, fmt.Sprintf("%f", sale.Glamount))

// 			line := strings.Join(entries, ",")
// 			lines = append(lines, line)

// 			costSubTotal += sale.CostValue
// 			saleSubTotal += sale.SaleValue

// 			costTotal += sale.CostValue
// 			saleTotal += sale.SaleValue

// 		}

// 		lines = append(lines, "")
// 		lines = append(lines, fmt.Sprintf(",,,,,,,,%f,%f,%f", costSubTotal, saleSubTotal, saleSubTotal-costSubTotal))
// 		lines = append(lines, "")
// 	}

// 	lines = append(lines, "")
// 	lines = append(lines, fmt.Sprintf("%s %s,,,,,,,,%f,%f,%f", term, "Totals", costTotal, saleTotal, saleTotal-costTotal))
// 	lines = append(lines, "")

// 	return lines
// }

// func writeTaxRewards(rewards []store.InvReward, fileName string) {

// 	var lines []string

// 	lines = append(lines, fmt.Sprint("Account,Symbol,Date,Qty,Cost,Cost Value"))

// 	for _, reward := range rewards {
// 		var entries []string
// 		entries = append(entries, reward.Account)
// 		entries = append(entries, reward.Symbol)
// 		entries = append(entries, reward.Date.Format("2006-01-02"))
// 		entries = append(entries, fmt.Sprintf("%f", reward.Qty))
// 		entries = append(entries, fmt.Sprintf("%f", reward.Cost))
// 		entries = append(entries, fmt.Sprintf("%f", reward.CostValue))

// 		line := strings.Join(entries, ",")
// 		lines = append(lines, line)
// 	}
// 	WriteLinesToFile(fileName, lines)

// }
