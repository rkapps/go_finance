package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/rkapps/go_finance/store"
	"github.com/rkapps/go_finance/utils"
)

const (
	alphaURL    = "https://www.alphavantage.co/query?"
	alphaAPIKey = "GG216NYO5GM8L88X" //
)

type hResponse struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
}

type aResponse struct {
	Res map[string]hResponse `json:"Time Series (Daily)"`
}

type TickerOResponse struct {
	Exchange  string `json:"Exchange"`
	Symbol    string `json:"Symbol"`
	Name      string `json:"Name"`
	Overview  string `json:"Description"`
	Sector    string `json:"Sector"`
	Industry  string `json:"Industry"`
	MktCap    string `json:"MarketCapitalization"`
	PERatio   string `json:"PERatio"`
	PBRatio   string `json:"PriceToBookRatio"`
	PSRatio   string `json:"PriceToSalesRatioTTM"`
	PEGRatio  string `json:"PEGRatio"`
	EPS       string `json:"EPS"`
	DivAmt    string `json:"DividendPerShare"`
	Yield     string `json:"DividendYield"`
	ExDivDate string `json:"ExDividendDate"`
	PayDate   string `json:"DividendDate"`
	PayRatio  string `json:"PayoutRatio"`
}

//GetTickerQuotes returns LastPrice and PrevPrice for the ticker
func GetTickerQuotes(tickers store.Tickers) (map[string]*store.TickerHistory, map[string]*store.TickerHistory) {

	thm := make(map[string]*store.TickerHistory)
	pthm := make(map[string]*store.TickerHistory)

	for _, ticker := range tickers {
		if ticker.IsIndex() {
			thm1 := getTickerQuoteFromAlpha(*ticker, "compact")
			if len(thm1) > 0 {
				thm[ticker.GetTickerID()] = thm1[len(thm1)-1]
				if len(thm1) > 1 {
					pthm[ticker.GetTickerID()] = thm1[len(thm1)-2]
				}
			}
		}
	}

	return thm, pthm
}

func getTickerQuoteFromAlpha(ticker store.Ticker, output string) []*store.TickerHistory {

	var ths []*store.TickerHistory

	symbol := fmt.Sprintf("&symbol=%s", ticker.Symbol)
	size := fmt.Sprintf("&outputsize=%s", output)
	api := fmt.Sprintf("&apikey=%s", alphaAPIKey)
	url := strings.Join([]string{alphaURL, "function=TIME_SERIES_DAILY", symbol, size, api}, "")
	// log.Printf("url: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Print(err)
		return ths
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var s = new(aResponse)
	s.Res = make(map[string]hResponse)
	err1 := json.Unmarshal(body, &s)

	if err1 != nil {
		//if the url does not return any data, just return
		log.Print(err1)
		// panic(err1)
		return ths
	}

	//sort of date keys
	dates := make([]string, 0, len(s.Res))
	for k := range s.Res {
		dates = append(dates, k)
	}
	sort.Strings(dates)

	for _, k := range dates {
		var h hResponse = s.Res[k]
		var th store.TickerHistory
		// th.Exchange = ticker.Exchange
		th.Symbol = ticker.Symbol
		th.Open, _ = strconv.ParseFloat(h.Open, 32)
		th.High, _ = strconv.ParseFloat(h.High, 32)
		th.Low, _ = strconv.ParseFloat(h.Low, 32)
		th.Close, _ = strconv.ParseFloat(h.Close, 32)
		th.Date = utils.DateFromString(k)
		ths = append(ths, &th)
		// th.Volume, _ = strconv.Atoi(h.Volume)
	}
	return ths
}

func UpdateTickerDetails(ticker *store.Ticker) (string, error) {

	var or TickerOResponse
	url := fmt.Sprintf("%sfunction=%s&symbol=%s&apikey=%s", alphaURL, "OVERVIEW", ticker.Symbol, alphaAPIKey)
	// log.Println(url)
	err := runHTTPGet(url, &or)
	if err != nil {
		return url, err
	}
	if len(or.Symbol) == 0 || len(or.Name) == 0 {
		return url, errors.New("Symbol is blank")
	}
	if len(or.Exchange) == 0 {
		return url, errors.New("Exchange is blank")
	}

	/*
		if len(or.Sector) == 0 {
			if len(ticker.Sector) == 0 {
				return errors.New("Sector is blank")
			}
		}
		if strings.Compare(or.Sector, "Other") == 0 {
			if len(ticker.Sector) == 0 {
				return errors.New("Sector is unknown")
			}
		}

		if len(or.MktCap) == 0 {
			return errors.New("Market Cap is zero")
		}
		var imcap, _ = strconv.Atoi(or.MktCap)
		if imcap == 0 {
			return errors.New("Market cap is zero")
		}

		if strings.Compare(or.EPS, "None") == 0 && strings.Compare(or.PBRatio, "None") == 0 && strings.Compare(or.PSRatio, "None") == 0 {
			return errors.New("EPS or PB/PE/PR is blank")
		}

		ticker.Exchange = or.Exchange
		ticker.Name = or.Name

		if len(ticker.Sector) == 0 && len(or.Sector) > 0 {

			if strings.Compare(or.Sector, "Other") != 0 {
				ticker.Sector = or.Sector
			}

			if strings.Compare(ticker.Sector, "Consumer Cyclical") == 0 {
				ticker.Sector = "Consumer Discretionary"
			}
			if strings.Compare(ticker.Sector, "Consumer Defensive") == 0 {
				ticker.Sector = "Consumer Staples"
			}
			if strings.Compare(ticker.Sector, "Financial Services") == 0 {
				ticker.Sector = "Financials"
			}

		}

		if len(ticker.Industry) == 0 && len(or.Industry) > 0 {
			if strings.Compare(or.Industry, "Other") != 0 {
				ticker.Industry = or.Industry
			}
		}
	*/

	var imcap, _ = strconv.Atoi(or.MktCap)

	ticker.Overview = or.Overview
	ticker.MarketCap = imcap
	ticker.EPS, _ = strconv.ParseFloat(or.EPS, 64)
	ticker.PBRatio, _ = strconv.ParseFloat(or.PBRatio, 64)
	ticker.PEGRatio, _ = strconv.ParseFloat(or.PEGRatio, 64)
	ticker.PERatio, _ = strconv.ParseFloat(or.PERatio, 64)
	ticker.PSRatio, _ = strconv.ParseFloat(or.PSRatio, 64)
	ticker.DivAmt, _ = strconv.ParseFloat(or.DivAmt, 64)
	ticker.Yield, _ = strconv.ParseFloat(or.Yield, 64)
	ticker.Yield = ticker.Yield * 100

	divDate := utils.DateFromString(or.PayDate)
	if !divDate.IsZero() {
		ticker.PayDate = &divDate
	}

	exDivDate := utils.DateFromString(or.ExDivDate)
	if !exDivDate.IsZero() {
		ticker.ExDivDate = &exDivDate
	}

	//Fix for bad paydate
	if ticker.ExDivDate == nil {
		ticker.PayDate = nil
		ticker.PayRatio = 0
	}
	if ticker.PayDate == nil {
		ticker.PayRatio = 0
		ticker.ExDivDate = nil
		ticker.Yield = 0
		ticker.DivAmt = 0
	}
	if ticker.DivAmt == 0 {
		ticker.Yield = 0
		ticker.ExDivDate = nil
		ticker.PayDate = nil
		ticker.PayRatio = 0
	}

	ticker.PayRatio, _ = strconv.ParseFloat(or.PayRatio, 64)
	ticker.PayRatio = ticker.PayRatio * 100
	// ticker.MarketCap = or.MktCap

	// log.Printf("Fundas: %v", or)
	// log.Println(ticker.FormatTickerData(";"))

	return url, nil
}
