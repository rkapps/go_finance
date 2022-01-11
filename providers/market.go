package providers

import (
	"github.com/rkapps/go_finance/store"
)

//GetTickersDetails returns the ticker details
func GetTickersDetails(baseURL string, symbols string) (store.Tickers, error) {

	var tickers store.Tickers
	var url = baseURL + "tickers?symbols=" + symbols
	// log.Printf("url: %s", url)
	err := runHTTPGet(url, &tickers)
	return tickers, err
}

//GetTickerDetails returns the ticker details
func GetTickerDetails(baseURL string, symbol string) (*store.Ticker, error) {

	// log.Printf("url: %s", url)
	var ticker *store.Ticker
	var url = baseURL + "tickers/" + symbol
	err := runHTTPGet(url, &ticker)
	return ticker, err
}

// //TickersLoad loads tickers
// func TickersLoad(baseURL string, ts []store.TickerImport) error {
// 	var tks store.Tickers

// 	var url = baseURL + "tickers/load"
// 	err := runHTTPPost(url, ts, &tks)
// 	return err
// }

// //TickerSearch searches for tickers based on the parameters
// func TickerSearch(baseURL string, ts store.TickerSearch) (store.Tickers, error) {
// 	var tks store.Tickers

// 	var url = baseURL + "tickers/search"
// 	err := runHTTPPost(url, ts, &tks)
// 	return tks, err
// }
