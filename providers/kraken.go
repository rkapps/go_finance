package providers

import (
	krakenapi "github.com/beldur/kraken-go-api-client"
)

var kapi *krakenapi.KrakenAPI

func init() {
	// kapi = krakenapi.New("KEY", "SECRET")
	// result, err := kapi.Query("Ticker", map[string]string{"pair": "XXBTZUSD"})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Result: %+v\n", result)
	// ticker, err := kapi.Ticker(krakenapi.XXBTZUSD, krakenapi.XETHZUSD)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Printf("%v: \n", ticker.XXBTZUSD.Close)

}
