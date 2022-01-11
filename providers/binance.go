package providers

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	baseURL   = "https://api.binance.com/api/v3/"
	tickerURL = "ticker/price?symbol="
)

//TickerPrice holds metadata for ticker price data returned by binance
type TickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

//GetTickerPrice returns the latest price for the symbol
func GetTickerPriceFromBinance(symbol string) (float64, error) {

	var tp TickerPrice
	symbol = strings.ReplaceAll(symbol, "-", "")
	symbol = fmt.Sprintf("%sT", symbol)

	url := fmt.Sprintf("%s%s%s", baseURL, tickerURL, strings.ToUpper(symbol))

	err := runHTTPGet(url, &tp)
	if err != nil {
		return 0.0, err
	} else {
		// log.Printf("Binance Ticker: %v", tp)
	}

	price, _ := strconv.ParseFloat(tp.Price, 64)
	return price, nil
}
