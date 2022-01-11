package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rkapps/go_finance/core"
	"github.com/rkapps/go_finance/store"
	"github.com/rkapps/go_finance/utils"
)

var fn *core.Finance
var fbApp *firebase.App
var fbAuthClient *auth.Client

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	mongoConnStr := os.Getenv("MONGO_ATLAS_CONN_STR")
	// mongoConnStr = "mongodb://localhost:27017"

	if mongoConnStr == "" {
		log.Fatal("Mongo DB connection string not available")
	}

	var err error

	fbApp, err = firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	fbAuthClient, err = fbApp.Auth(context.Background())
	if err != nil {
		log.Fatalf("Firebase authorization error: %v\n", err)
	}

	fn, err = core.NewFinance(mongoConnStr)
	if err != nil {
		log.Fatalf("Finance initialization error: %v\n", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", indexHandler)
	//test handler

	router.HandleFunc("/activities/import", authHandler(activitiesImportHandler))

	router.HandleFunc("/investments/accounts", authHandler(investmentsAccountsHandler))
	router.HandleFunc("/investments/holdings", authHandler(investmentsHoldingsHandler))
	router.HandleFunc("/investments/lots", authHandler(investmentsLotsHandler))
	router.HandleFunc("/investments/gainloss", authHandler(investmentsGainLossHandler))
	router.HandleFunc("/investments/income", authHandler(investmentsIncomeHandler))

	router.HandleFunc("/tickers", tickersHandler)
	router.HandleFunc("/tickers/import", tickersImportHandler)
	router.HandleFunc("/tickers/groups", tickersGroupsHandler)
	router.HandleFunc("/tickers/search", tickersSearchHandler)

	router.HandleFunc("/tickers/updateCryptosEOD", updateCryptosEODHandler)
	router.HandleFunc("/tickers/updateStocksEOD", updateStocksEODHandler)
	router.HandleFunc("/tickers/updateNews", updateNewsHandler)
	router.HandleFunc("/tickers/updateTickers", updateTickersHandler)

	router.HandleFunc("/tickers/updateStocksRealtime", updateStocksRealtimeHandler)
	router.HandleFunc("/tickers/updateCryptosRealtime", updateCryptosRealtimeHandler)

	router.HandleFunc("/tickers/{symbol}/delete", tickerDeleteHandler)
	router.HandleFunc("/tickers/{symbol}/history", tickerHistoryHandler)
	router.HandleFunc("/tickers/{symbol}/news", tickerNewsHandler)

	router.HandleFunc("/transactions/aggregate", authHandler(transactionsAggregateHandler))

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port),
		handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
			handlers.AllowedOrigins([]string{"*"}))(router)))

}

// indexHandler responds to requests with our greeting.git
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Rkapps - go_finance !!!")
}

func authHandler(f func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		// fmt.Printf("%v", r.Header)
		var header = r.Header.Get("Authorization")
		var token string

		if len(header) > 7 && strings.ToLower(header[0:6]) == "bearer" {
			token = header[7:]
		} else {
			fmt.Println("Authorization token not available")
			http.Error(w, "Authorization token not available", http.StatusUnauthorized)
			return
		}
		authToken, err := fbAuthClient.VerifyIDToken(r.Context(), token)
		if err != nil {
			fmt.Printf("VerifyIdToken: %v\n", err)
			http.Error(w, "Authorization token not verified", http.StatusUnauthorized)
			// fmt.Fprintf(w, err.Error())
			return
		}

		// fmt.Printf("Token: %v\n", authToken.UID)
		ctx := context.WithValue(r.Context(), store.UserContextUID, store.User{UID: authToken.UID})

		f(w, r.WithContext(ctx))
	}

}

// activitiesImportHandler handles imported activities
func activitiesImportHandler(w http.ResponseWriter, r *http.Request) {

	values := r.URL.Query()
	log.Printf("Activities Import - Values: %v\n", values)

	actyType := values["actyType"][0]
	group := ""
	category := ""
	if len(values.Get("group")) > 0 {
		group = values["group"][0]
	}
	if len(values.Get("category")) > 0 {
		category = values["category"][0]
	}

	var fromDate, toDate time.Time
	fromDate = utils.DateFromString(values["fromDate"][0])
	toDate = utils.DateFromString(values["toDate"][0])

	var actvs store.Activities
	err := json.NewDecoder(r.Body).Decode(&actvs)
	if err != nil {
		fmt.Printf("activitiesImportHandler: %v\n", err)
		fmt.Fprintf(w, err.Error())
		return
	}

	log.Printf("Group: %s Category: %s FromDate: %v ToDate: %v", group, category, fromDate, toDate)
	// if err == nil {
	err = fn.ActivitiesImport(r.Context(), actyType, group, category, &fromDate, &toDate, actvs)
	// // }

	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, err.Error())
	}

	log.Printf("Import activities - %s count: %d\n", actyType, len(actvs))

}

func investmentsAccountsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Investment accounts")
	fn.MDB.InvestmentsAccounts(r.Context())
}

func investmentsHoldingsHandler(w http.ResponseWriter, r *http.Request) {

	values := r.URL.Query()
	group := ""
	category := ""
	if len(values.Get("group")) > 0 {
		group = values["group"][0]
	}
	if len(values.Get("category")) > 0 {
		category = values["category"][0]
	}

	byAccount, _ := strconv.ParseBool(values["byAccount"][0])
	log.Printf("Values: %v\n", values)

	// fn.UpdateStocksRealtime(r.Context())
	// fn.UpdateCryptosRealtime(r.Context())

	holds, err := fn.InvestmentsHoldings(r.Context(), group, category, byAccount)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, err.Error())
		return
	}
	fmt.Printf("Holdings: %d\n", len(holds))

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(holds); err != nil {
		panic(err)
	}
}

func investmentsLotsHandler(w http.ResponseWriter, r *http.Request) {

	values := r.URL.Query()
	group := ""
	category := ""
	symbol := ""
	status := ""

	if len(values.Get("group")) > 0 {
		group = values["group"][0]
	}
	if len(values.Get("category")) > 0 {
		category = values["category"][0]
	}
	if len(values.Get("symbol")) > 0 {
		symbol = values["symbol"][0]
	}
	if len(values.Get("status")) > 0 {
		status = values["status"][0]
	}

	lots := fn.InvestmentsLots(r.Context(), group, category, symbol, status)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(lots); err != nil {
		panic(err)
	}
	fmt.Printf("InvestmentsLotsHandler - Lots: %d\n", len(lots))
}

func investmentsGainLossHandler(w http.ResponseWriter, r *http.Request) {

	values := r.URL.Query()
	year, _ := strconv.Atoi(values["year"][0])
	group := ""
	category := ""
	if len(values.Get("group")) > 0 {
		group = values["group"][0]
	}
	if len(values.Get("category")) > 0 {
		category = values["category"][0]
	}

	ft := time.Date(year, time.Month(1), 1, 0, 0, 0, 0, time.Now().Location())
	et := time.Date(year, time.Month(12), 31, 24, 0, 0, 0, time.Now().Location())

	log.Printf("investmentsGainLoss Query Values: %v", values)

	lots := fn.InvestmentsGainLoss(r.Context(), group, category, ft, et)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(lots); err != nil {
		panic(err)
	}
	fmt.Printf("InvestmentsGainLossHandler - Lots: %d\n", len(lots))
}

func investmentsIncomeHandler(w http.ResponseWriter, r *http.Request) {

	values := r.URL.Query()
	year, _ := strconv.Atoi(values["year"][0])
	group := ""
	category := ""
	open := false
	if len(values.Get("group")) > 0 {
		group = values["group"][0]
	}
	if len(values.Get("category")) > 0 {
		category = values["category"][0]
	}
	if len(values.Get("open")) > 0 {
		open, _ = strconv.ParseBool(values["open"][0])
	}

	ft := time.Date(year, time.Month(1), 1, 0, 0, 0, 0, time.Now().Location())
	et := time.Date(year, time.Month(12), 31, 24, 0, 0, 0, time.Now().Location())

	log.Printf("investmentsIncome Query Values: %v", values)

	lots := fn.InvestmentsRewards(r.Context(), group, category, open, ft, et)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(lots); err != nil {
		panic(err)
	}
	fmt.Printf("InvestmentsRewardsHandler - Lots: %d\n", len(lots))
}

//tickersImportHandler loads tickers from a csv file
func tickersImportHandler(w http.ResponseWriter, r *http.Request) {

	var tickers []store.Ticker
	err := json.NewDecoder(r.Body).Decode(&tickers)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, err.Error())
		return
	}

	fn.LoadTickers(r.Context(), tickers)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tickers); err != nil {
		// panic(err)
	}

}

func tickerDeleteHandler(w http.ResponseWriter, r *http.Request) {
	//tickers := m.db.GetAllTickers(r.Context())
	// enableCors(&w)
	vars := mux.Vars(r)
	id := vars["symbol"]
	log.Printf("Delete Ticker: %s", id)

	err := fn.DeleteTicker(r.Context(), id)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func tickerHistoryHandler(w http.ResponseWriter, r *http.Request) {
	//tickers := m.db.GetAllTickers(r.Context())
	// enableCors(&w)
	vars := mux.Vars(r)
	id := vars["symbol"]
	log.Printf("Get Ticker History: %s", id)

	th := fn.GetTickerHistory(r.Context(), id)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(th); err != nil {
		panic(err)
	}
}

func tickersHandler(w http.ResponseWriter, r *http.Request) {

	keys, ok := r.URL.Query()["symbols"]
	log.Printf("Get Tickers - symbols: %s", keys)

	if !ok || len(keys[0]) < 1 {
		log.Println("Url param symbols is missing")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	symbols := strings.Split(keys[0], ",")
	// log.Printf("Get Tickers: %s", symbols)
	tickers := fn.GetTickers(r.Context(), symbols)
	var tsa store.Tickers
	if tickers == nil {
		tsa = store.Tickers{}
	} else {
		var tsm map[string]*store.Ticker
		tsm = make(map[string]*store.Ticker)
		for _, ts := range tickers {
			tsm[ts.Symbol] = ts
		}
		// var sm map[string]int
		for _, entry := range symbols {
			ts := tsm[strings.ToUpper(entry)]
			if ts != nil {
				tsa = append(tsa, ts)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tsa); err != nil {
		panic(err)
	}
}

func tickerNewsHandler(w http.ResponseWriter, r *http.Request) {
	//tickers := m.db.GetAllTickers(r.Context())
	// enableCors(&w)
	vars := mux.Vars(r)
	id := vars["symbol"]
	log.Printf("Get Ticker News: %s", id)

	tn := fn.GetTickerNews(r.Context(), id)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tn); err != nil {
		panic(err)
	}
}

//tickersGroupsHandler loads ticker groups
func tickersGroupsHandler(w http.ResponseWriter, r *http.Request) {

	// enableCors(&w)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	tickers, _ := fn.MDB.GetTickerGroups(r.Context())
	// t := tickers[0]
	// log.Printf("%v", t)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tickers); err != nil {
		panic(err)
	}

}

//loadTickersHandler loads tickers from a csv file
func tickersSearchHandler(w http.ResponseWriter, r *http.Request) {

	// enableCors(&w)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var ts store.TickerSearch
	json.NewDecoder(r.Body).Decode(&ts)
	log.Printf("Search for text: %v", ts)
	tickers := fn.SearchTickers(r.Context(), ts)
	// t := tickers[0]
	// log.Printf("%v", t)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tickers); err != nil {
		panic(err)
	}
	log.Printf("Search tickers: %d\n", len(tickers))

}

// indexHandler responds to requests with our greeting.git
func transactionsAggregateHandler(w http.ResponseWriter, r *http.Request) {

	// vars := mux.Vars(r)
	values := r.URL.Query()
	year, _ := strconv.Atoi(values["year"][0])

	// var fromDate, toDate time.Time
	// fromDate = utils.DateFromString(values["fromDate"][0])
	// toDate = utils.DateFromString(values["toDate"][0])
	// et := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 24, 0, 0, 0, time.Now().Location())

	ft := time.Date(year, time.Month(1), 1, 0, 0, 0, 0, time.Now().Location())
	et := time.Date(year, time.Month(12), 31, 24, 0, 0, 0, time.Now().Location())

	log.Printf("Values: %v ", values)
	taggs, err := fn.AggregateTransactions(r.Context(), &ft, &et)
	if err != nil {
		fmt.Printf("Error: %v", err)
		fmt.Fprintf(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(taggs); err != nil {
		panic(err)
	}

}

func updateStocksEODHandler(w http.ResponseWriter, r *http.Request) {
	fn.UpdateStocksEOD(r.Context())
}

func updateCryptosEODHandler(w http.ResponseWriter, r *http.Request) {
	fn.UpdateCryptosEOD(r.Context())
}

func updateStocksRealtimeHandler(w http.ResponseWriter, r *http.Request) {
	fn.UpdateStocksRealtime(r.Context())
}

func updateCryptosRealtimeHandler(w http.ResponseWriter, r *http.Request) {
	fn.UpdateCryptosRealtime(r.Context())
}

func updateNewsHandler(w http.ResponseWriter, r *http.Request) {
	fn.UpdateTickersNews(r.Context())
}

func updateTickersHandler(w http.ResponseWriter, r *http.Request) {
	var ts store.Tickers
	fn.UpdateTickers(r.Context(), ts, true)
}
