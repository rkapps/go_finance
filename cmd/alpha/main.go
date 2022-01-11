package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rkapps/go_finance/providers"
	"github.com/rkapps/go_finance/store"
	"github.com/rkapps/go_finance/utils"
)

const (
	sep = ";"
)

func main() {

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Invalid arguements")
		return
	}

	inpFile := args[1]
	outFile := args[2]

	lines := utils.LoadFromFile(inpFile, "\t")

	_, nls, els := processLines(lines)
	// log.Println(len(uts))
	errFile := strings.ReplaceAll(outFile, ".csv", "-error.csv")
	utils.WriteToFile(outFile, nls, sep)
	utils.WriteToFile(errFile, els, sep)

}

func processLines(lines [][]string) (store.Tickers, [][]string, [][]string) {

	var nls [][]string
	var els [][]string

	var t *store.Ticker
	var uts store.Tickers

	for _, line := range lines {

		if len(line) == 0 {
			continue
		}
		log.Println(line)
		t = &store.Ticker{}
		t.Symbol = line[1]
		t.Name = line[2]

		estr := ""
		_, err := providers.UpdateTickerDetails(t)

		var nl []string
		mcap := strconv.Itoa(t.MarketCap)

		if err == nil {
			if t.MarketCap < 1000000000 {

				estr = "Market Cap < 1B"
				nl = append(nl, t.Exchange, t.Symbol, t.Name, t.Sector, t.Industry, mcap, estr)
				els = append(els, nl)

			} else {
				nl = append(nl, t.Exchange, t.Symbol, t.Name, t.Sector, t.Industry, mcap, estr)
				nls = append(nls, nl)
			}
		} else {
			estr = err.Error()
			nl = append(nl, t.Exchange, t.Symbol, t.Name, t.Sector, t.Industry, mcap, estr)
			els = append(els, nl)
		}

		uts = append(uts, t)

		time.Sleep(time.Second)

		// if len(uts) > 100 {
		// 	break
		// }
	}

	return uts, nls, els
}
