package providers

import (
	"fmt"
	"strings"
	"time"

	"github.com/rkapps/go_finance/store"
	"github.com/rkapps/go_finance/utils"
)

const (
	// baseURL   = "https://api.tiingo.com/tiingo/"
	apiToken  = "74b37cc2445eaba7caa254121538ab506a5ca593"
	eodURL    = "https://api.tiingo.com/tiingo/daily/"
	rTimeURL  = "https://api.tiingo.com/iex/"
	cryptoURL = "https://api.tiingo.com/tiingo/crypto/prices"
	fundasURL = "https://api.tiingo.com/tiingo/fundamentals/"
	newsURL   = "https://api.tiingo.com/tiingo/news"
)

type cResponse struct {
	Symbol  string                 `json:"ticker"`
	CTicker []*store.TickerHistory `json:"priceData"`
}

type rResponse struct {
	Symbol   string     `json:"ticker"`
	TngoLast float64    `json:"tngoLast"`
	Date     *time.Time `json:"timestamp"`
}

type nResponse struct {
	Symbol      string    `json:"symbol" firestore:"symbol,omitempty"`
	Date        time.Time `json:"publishedDate" firestore:"date,omitempty"`
	URL         string    `json:"url" firestore:"url,omitempty"`
	Title       string    `json:"title" firestore:"title,omitempty"`
	Description string    `json:"description" firestore:"description,omitempty"`
	Source      string    `json:"source" firestore:"source,omitempty"`
}

//GetTickerRealTimeQuote returns the real time last price and date
func GetTickerRealTimeQuote(ticker store.Ticker) (float64, *time.Time) {

	url := strings.Join([]string{rTimeURL, ticker.Symbol, "?token=", apiToken}, "")
	// log.Printf("url: %s", url)
	var s = new([]rResponse)
	runHTTPGet(url, &s)
	for _, res := range *s {
		// log.Println(res)
		return res.TngoLast, res.Date
	}
	return 0.0, nil
}

//GetCryptoHistory returns the EOD quotes for the date range
func GetCryptoHistory(tickers store.Tickers, st time.Time, et time.Time) map[string][]*store.TickerHistory {

	var thm = make(map[string][]*store.TickerHistory)
	var sBuilder strings.Builder
	tm := make(map[string]*store.Ticker)

	for _, ticker := range tickers {
		if ticker.IsCrypto() {
			ts := strings.ToLower(strings.Replace(ticker.Symbol, "-USD", "USD", 1))
			sBuilder.WriteString(ts)
			sBuilder.WriteString(",")
			tm[ts] = ticker
		}
	}

	if len(sBuilder.String()) == 0 {
		return thm
	}

	dates := fmt.Sprintf("&startDate=%s&endDate=%s", utils.DateFormat1(st), utils.DateFormat1(et))
	url := strings.Join([]string{cryptoURL, "?tickers=", sBuilder.String(), "&resampleFreq=1day", dates, "&token=", apiToken}, "")
	// log.Printf("url: %s", url)
	var s = new([]cResponse)

	runHTTPGet(url, s)

	for _, res := range *s {
		t := tm[res.Symbol]
		if t != nil {
			for _, th := range res.CTicker {
				// th.Exchange = t.Exchange
				th.Symbol = t.Symbol
				th.Date = utils.DateAdjustForUtc(th.Date)
				// log.Printf("Date: %v price: %f", th.Date, th.Close)
			}
			thm[t.GetTickerID()] = res.CTicker
		}
	}

	return thm
}

//GetTickerHistoryQuotes returns the historical EOD quotes between the time range.
func GetTickersHistory(tickers store.Tickers, st time.Time, et time.Time) map[string][]*store.TickerHistory {
	var thm = make(map[string][]*store.TickerHistory)
	for _, ticker := range tickers {
		if ticker.IsStock() || ticker.IsMutf() {
			ths := getTickerQuoteFromTingo(*ticker, st, et)
			thm[ticker.GetTickerID()] = ths
		}
	}
	return thm
}

//GetTickerNews returns the news for the ticker
func GetTickerNews(tickers store.Tickers) map[string][]store.TickerNews {
	var tnm = make(map[string][]store.TickerNews)
	for _, ticker := range tickers {
		tnm[ticker.GetTickerID()] = getTickerNews(ticker)
	}
	return tnm
}

func getTickerNews(ticker *store.Ticker) []store.TickerNews {

	var tnr []nResponse
	// var tns []core.TickerNews
	var rtns []store.TickerNews
	var symbol = ticker.Symbol
	if ticker.IsCrypto() {
		symbol = strings.ReplaceAll(symbol, "-USD", "")
	}
	url := strings.Join([]string{newsURL, "?tickers=", symbol, "&token=", apiToken}, "")
	// tg.runTickerAPI(url, &tnr)

	runHTTPGet(url, &tnr)

	nmap := make(map[string]store.TickerNews)

	for _, nr := range tnr {

		// log.Println(tn.Title)
		if _, ok := nmap[nr.Title]; !ok {
			var tn store.TickerNews
			tn = store.TickerNews{}
			tn.Symbol = ticker.Symbol
			tn.Date = nr.Date
			tn.URL = nr.URL
			tn.Title = nr.Title
			tn.Description = nr.Description
			tn.Source = nr.Source
			// log.Printf("Date: %v Title: %s", tn.Date, tn.Title)
			nmap[tn.Title] = tn
			rtns = append(rtns, tn)
		}
	}
	// log.Println(tfs)
	return rtns

}

//GetEODTickerQuote gets EOD quote for a ticker
func getTickerQuoteFromTingo(ticker store.Ticker, st time.Time, et time.Time) []*store.TickerHistory {

	var ths []*store.TickerHistory
	dates := fmt.Sprintf("&startDate=%s&endDate=%s", utils.DateFormat1(st), utils.DateFormat1(et))
	url := strings.Join([]string{eodURL, ticker.Symbol, "/prices?token=", apiToken, dates}, "")
	// log.Printf("url: %s", url)
	runHTTPGet(url, &ths)
	//	ticker.LastPrice = ((ths)[0].Close)
	for _, th := range ths {
		// th.Exchange = ticker.Exchange
		th.Symbol = ticker.Symbol
		th.Date = utils.DateAdjustForUtc(th.Date)
		th.Close = th.AdjClose
		th.High = th.AdjHigh
		th.Low = th.AdjLow
		th.Open = th.AdjOpen
	}

	return ths
}
