package store

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/rkapps/go_finance/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

const (
	//ExNasdaq defines the string NASDAQ
	ExNasdaq string = "NASDAQ"
	//ExNyse defines the string NYSE
	ExNyse string = "NYSE"
	//ExNyseArca defines the string NYSEARCA
	ExNyseArca string = "NYSEARCA"
	//ExCurrency defines the string CURRENCY
	ExCurrency string = "CURRENCY"
	//ExMutf defines the string MUTF
	ExMutf string = "MUTF"
	//ExIndex defines the string INDEX
	ExIndex string = "INDEX"
	//ExOtc defines the string OTC
	ExOtc string = "OTC"

	//SMA defines the string SMA
	SMA string = "SMA"
	//EMA defines the string EMA
	EMA string = "EMA"
	//RSI defines the string RSI
	RSI string = "RSI"
)

var (
	//PerfPeriods defines the performance periods
	PerfPeriods []string = []string{"1W", "1M", "3M", "6M", "YTD", "1Y", "2Y", "3Y", "5Y"}
	//RSIPeriods defines the RSI periods
	RSIPeriods []int = []int{5, 10, 14, 20, 26}
	//SMAPeriods defines the SMA periods
	SMAPeriods []int = []int{5, 20, 50, 100, 200}
	//EMAPeriods defines the EMA periods
	EMAPeriods []int = []int{5, 12, 26, 50, 100, 200}
)

// Ticker holds metadata about a ticker.
type Ticker struct {
	Symbol      string                        `json:"symbol" bson:"symbol"`
	Exchange    string                        `json:"exchange" bson:"exchange"`
	Name        string                        `json:"name" bson:"name"`
	Sector      string                        `json:"sector" bson:"sector"`
	Industry    string                        `json:"industry" bson:"industry"`
	Overview    string                        `json:"overview" bson:"overview"`
	MarketCap   int                           `json:"marketCap" bson:"marketCap"`
	Volume      int                           `json:"volume" bson:"volume"`
	AvgVolume   int                           `json:"avgVolume" bson:"avgVolume"`
	EPS         float64                       `json:"eps" bson:"eps"`
	PERatio     float64                       `json:"peRatio" bson:"peRatio"`
	PEGRatio    float64                       `json:"pegRatio" bson:"pegRatio"`
	PBRatio     float64                       `json:"pbRatio" bson:"pbRatio"`
	PSRatio     float64                       `json:"psRatio" bson:"psRatio"`
	Tpeg1Y      float64                       `json:"tpeg1Y" bson:"tpeg1Y"`
	DivAmt      float64                       `json:"divAmt" bson:"divAmt"`
	Yield       float64                       `json:"yield" bson:"yield"`
	ExDivDate   *time.Time                    `json:"exDivDate" bson:"exDivDate"`
	PayDate     *time.Time                    `json:"payDate" bson:"payDate"`
	PayRatio    float64                       `json:"payRatio" bson:"payRatio"`
	PrDate      *time.Time                    `json:"prDate" bson:"prDate"`
	PrOpen      float64                       `json:"prOpen" bson:"prOpen"`
	PrHigh      float64                       `json:"prHigh" bson:"prHigh"`
	PrLow       float64                       `json:"prLow" bson:"prLow"`
	PrClose     float64                       `json:"prClose" bson:"prClose"`
	PrLast      float64                       `json:"prLast" bson:"prLast"`
	PrPrev      float64                       `json:"prPrev" bson:"prPrev"`
	PrDiffAmt   float64                       `json:"prDiffAmt" bson:"prDiffAmt"`
	PrDiffPerc  float64                       `json:"prDiffPerc" bson:"prDiffPerc"`
	Pr52WkHigh  float64                       `json:"pr52WkHigh" bson:"pr52WkHigh"`
	Pr52WkLow   float64                       `json:"pr52WkLow" bsone:"pr52WkLow"`
	Performance map[string]map[string]float64 `json:"performance" bson:"performance"`
	Technicals  map[string]map[string]float64 `json:"technicals" bson:"technicals"`
}

//Tickers holds a list of tickers
type Tickers []*Ticker

// TickerGroup holds the sector and industry combination
type TickerGroup struct {
	Sector   string `json:"sector" bson:"sector"`
	Industry string `json:"industry" bson:"industry"`
}

//TickerGroups holds a list of ticker groups.
type TickerGroups []*TickerGroup

//TickerHistory holds metadata about a ticker history
type TickerHistory struct {
	Symbol   string             `json:"symbol" firestore:"symbol,omitempty"`
	Date     time.Time          `json:"date" firestore:"date,omitempty"`
	Open     float64            `json:"open" firestore:"open,omitempty"`
	High     float64            `json:"high" firestore:"high,omitempty"`
	Low      float64            `json:"low" firestore:"low,omitempty"`
	Close    float64            `json:"close" firestore:"close,omitempty"`
	AdjOpen  float64            `json:"adjOpen" firestore:"open,omitempty"`
	AdjHigh  float64            `json:"adjHigh" firestore:"high,omitempty"`
	AdjLow   float64            `json:"adjLow" firestore:"low,omitempty"`
	AdjClose float64            `json:"adjClose" firestore:"close,omitempty"`
	DivCash  float64            `json:"divCash" firestore:"divCash,omitempty"`
	SMA      map[string]float64 `json:"sma" firestore:"sma,omitempty"`
	EMA      map[string]float64 `json:"ema" firestore:"sma,omitempty"`
	RSI      map[string]float64 `json:"rsi" firestore:"rsi,omitempty"`
}

//TickerSearch holds metadata for search
type TickerSearch struct {
	Function     string   `json:"function"`
	Sectors      []string `json:"sectors"`
	Industries   []string `json:"industries"`
	SearchText   string   `json:"searchText"`
	PerfPeriod   string   `json:"perfPeriod"`
	FromPerfPerc int      `json:"fromPerfPerc"`
	ToPerfPerc   int      `json:"toPerfPerc"`
	FromDiv      int      `json:"fromDiv"`
	ToDiv        int      `json:"toDiv"`
	RsiPeriod    string   `json:"rsiPeriod"`
	FromRsi      int      `json:"fromRsi"`
	ToRsi        int      `json:"toRsi"`
	// PrAboveMA     string   `json:"prAboveMa"`
	// PrAbovePeriod string   `json:"prAbovePeriod"`
	PrAbove  bool   `json:"prAbove"`
	PrMA     string `json:"prMa"`
	PrPeriod string `json:"prPeriod"`
}

//TickerNews holds metadata for news feed
type TickerNews struct {
	Symbol      string    `json:"symbol" firestore:"symbol,omitempty"`
	Date        time.Time `json:"date" firestore:"date,omitempty"`
	URL         string    `json:"url" firestore:"url,omitempty"`
	Title       string    `json:"title" firestore:"title,omitempty"`
	Description string    `json:"description" firestore:"description,omitempty"`
	Source      string    `json:"source" firestore:"source,omitempty"`
}

func createTickersIndices(ctx context.Context, col *mongo.Collection) {

	keys := bsonx.Doc{{Key: "exchange", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_exchange", keys, false)

	keys = bsonx.Doc{{Key: "symbol", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_symbol", keys, true)

	keys = bsonx.Doc{{Key: "zkeywords", Value: bsonx.String("text")}}
	createIndex(ctx, col, "idx_text", keys, false)

	keys = bsonx.Doc{{Key: "sector", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_sector", keys, false)

	keys = bsonx.Doc{{Key: "yield", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_yield", keys, false)

	keys = bsonx.Doc{{Key: "prDiffPerc", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_prDiffPerc", keys, false)

	for _, pp := range PerfPeriods {
		var field = "performance." + pp + ".diff"
		keys = bsonx.Doc{{Key: field, Value: bsonx.Int32(1)}}
		createIndex(ctx, col, "idx_"+field, keys, false)
	}

	for _, rp := range RSIPeriods {
		var field = "technicals.RSI." + strconv.Itoa(rp)
		keys = bsonx.Doc{{Key: field, Value: bsonx.Int32(1)}}
		createIndex(ctx, col, "idx_"+field, keys, false)
	}

	for _, sp := range SMAPeriods {
		var field = "technicals.SMA." + strconv.Itoa(sp)
		keys = bsonx.Doc{{Key: field, Value: bsonx.Int32(1)}}
		createIndex(ctx, col, "idx_"+field, keys, false)
	}

	for _, ep := range EMAPeriods {
		var field = "technicals.SMA." + strconv.Itoa(ep)
		keys = bsonx.Doc{{Key: field, Value: bsonx.Int32(1)}}
		createIndex(ctx, col, "idx_"+field, keys, false)
	}

}

func createTickerHistoryIndices(ctx context.Context, col *mongo.Collection) {
	keys := bsonx.Doc{
		{Key: "symbol", Value: bsonx.Int32(1)},
		{Key: "date", Value: bsonx.Int32(1)},
	}
	createIndex(ctx, col, "idx_thistory", keys, true)
}

func createTickerNewsIndices(ctx context.Context, col *mongo.Collection) {
	keys := bsonx.Doc{
		{Key: "symbol", Value: bsonx.Int32(1)},
		{Key: "seq", Value: bsonx.Int32(1)},
	}
	createIndex(ctx, col, "idx_tnews", keys, true)
}

//GetTickerID returns the unique ticker ID
func (t *Ticker) GetTickerID() string {
	return t.Exchange + ":" + t.Symbol
}

//IsCrypto returns true if the ticker is a cryptocurrency
func (t Ticker) IsCrypto() bool {
	return strings.Compare(t.Exchange, "CURRENCY") == 0
}

//IsIndex returns true if the ticker is a index
func (t Ticker) IsIndex() bool {
	return strings.Contains(t.Exchange, "INDEX")
}

//IsMutf returns true if the ticker is a Mutual Fund
func (t Ticker) IsMutf() bool {
	return strings.Compare(t.Exchange, "MUTF") == 0
}

//IsStock returns true if the ticker is a NAsdaq, nyse, nysearca
func (t Ticker) IsStock() bool {
	return strings.Compare(t.Exchange, ExNasdaq) == 0 ||
		strings.Compare(t.Exchange, ExNyse) == 0 ||
		strings.Compare(t.Exchange, ExNyseArca) == 0 ||
		strings.Compare(t.Exchange, ExOtc) == 0
}

//SetPriceDiff sets the price difference between the last price and previous price
func (t *Ticker) SetPriceDiff() {

	t.PrDiffAmt, t.PrDiffPerc = utils.PriceDiff(t.PrLast, t.PrPrev)
	// if t.PrPrev == 0 {
	// 	log.Printf("Ticker: %v", t)
	// }
	// log.Println(t.PrDiffPerc)
}

func (mdb *MongoDB) AddTicker(ctx context.Context, t *Ticker, keywords []string) error {

	tickersCol := mdb.db.Collection(TICKERScol)
	_, err := tickersCol.InsertOne(ctx, t)
	filter := bson.M{"symbol": t.Symbol}
	if err != nil {
		update := bson.M{"$set": t}
		_, err = tickersCol.UpdateOne(ctx, filter, update)
	}

	// keys := []string{strings.ToLower(t.Symbol), strings.ToLower(t.Exchange), strings.ToLower(t.Name), strings.ToLower(t.Sector), strings.ToLower(t.Industry)}
	// log.Println(keys)
	// keywords := utils.CreateTickerKeywords(keys)
	update := bson.M{"$set": bson.M{"zkeywords": keywords}}
	_, err = tickersCol.UpdateOne(ctx, filter, update)

	return err
}

func (mdb *MongoDB) DeleteTicker(ctx context.Context, symbol string) error {
	tickersCol := mdb.db.Collection(TICKERScol)
	_, err := tickersCol.DeleteOne(ctx, bson.M{"symbol": strings.ToUpper(symbol)})
	thistoryCol := mdb.db.Collection(THISTORYcol)
	_, err = thistoryCol.DeleteMany(ctx, bson.M{"symbol": strings.ToUpper(symbol)})
	return err
}

func (mdb *MongoDB) GetTicker(ctx context.Context, symbol string) *Ticker {
	query := bson.M{"symbol": strings.ToUpper(symbol)}
	tickers := mdb.getTickers(query, nil)
	if len(tickers) > 0 {
		return tickers[0]
	}
	return nil
}

func (mdb *MongoDB) GetTickers(ctx context.Context, symbols []string) Tickers {

	var query primitive.M
	if len(symbols) == 0 {
		query = bson.M{}
	} else {
		var qs []string
		for _, symbol := range symbols {
			qs = append(qs, strings.ToUpper(symbol))
		}
		query = bson.M{
			"symbol": bson.M{"$in": qs},
		}
	}
	return mdb.getTickers(query, nil)
}

func (mdb *MongoDB) GetTickersByExchange(ctx context.Context, exchanges []string) Tickers {
	query := bson.M{
		"exchange": bson.M{"$in": exchanges},
	}
	return mdb.getTickers(query, nil)
}

func (mdb *MongoDB) GetTickerHistory(ctx context.Context, symbol string) []*TickerHistory {

	var th []*TickerHistory
	query := bson.M{"symbol": strings.ToUpper(symbol)}
	tHistoryCol := mdb.db.Collection(THISTORYcol)
	cur, err := tHistoryCol.Find(context.TODO(), query)
	if err != nil {
		log.Println(err)
	} else {
		cur.All(context.TODO(), &th)
	}
	if th == nil {
		th = []*TickerHistory{}
	}
	// print(len(th))
	return th
}

func (mdb *MongoDB) GetTickerNews(ctx context.Context, symbol string) []*TickerNews {

	var tn []*TickerNews
	query := bson.M{"symbol": strings.ToUpper(symbol)}
	tNewsCol := mdb.db.Collection(TNEWScol)
	cur, err := tNewsCol.Find(context.TODO(), query)
	if err != nil {
		log.Println(err)
	} else {
		cur.All(context.TODO(), &tn)
	}
	if tn == nil {
		tn = []*TickerNews{}
	}
	// print(len(th))
	return tn
}

func (mdb *MongoDB) GetTickerGroups(ctx context.Context) (TickerGroups, error) {

	var tgs TickerGroups

	var pipeline []interface{}
	var match map[string]interface{}
	match = make(map[string]interface{})

	query := bson.M{
		"_id": bson.M{
			"sector":   "$sector",
			"industry": "$industry",
		},
	}

	queryStage := bson.M{
		"$group": query,
	}

	matchStage := bson.M{
		"$match": match,
	}

	pipeline = append(pipeline, matchStage, queryStage)
	col := mdb.db.Collection(TICKERScol)

	cur, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Aggregate error: %v", err)
		return nil, err
	} else {
		var results []map[string]interface{}
		cur.All(context.TODO(), &results)
		for _, result := range results {
			for _, v := range result {
				tg := &TickerGroup{}
				switch entry := v.(type) {
				case map[string]interface{}:

					tg.Sector = fmt.Sprintf("%s", entry["sector"])
					tg.Industry = fmt.Sprintf("%s", entry["industry"])
				}

				tgs = append(tgs, tg)
			}

		}
	}
	return tgs, nil
}

func (mdb *MongoDB) getTickers(query primitive.M, ops *options.FindOptions) Tickers {

	if ops == nil {
		ops = options.Find()
	}
	var result Tickers
	tickersCol := mdb.db.Collection(TICKERScol)
	cur, err := tickersCol.Find(context.TODO(), query, ops)

	if err != nil {
		log.Println(err)
	} else {
		cur.All(context.TODO(), &result)
	}
	return result
}

func (mdb *MongoDB) SearchTickers(ctx context.Context, ts TickerSearch) Tickers {
	// query = bson.M{
	// 	"$expr": bson.M{
	// 		"$lt": bson.A{"$technicals.SMA.50", "$technicals.SMA.200"},
	// 	},
	// }
	var options = options.Find()
	var query map[string]interface{}
	query = make(map[string]interface{})
	var field string

	//expr = make(map[string]inteface)
	if len(ts.SearchText) > 0 {
		query["$text"] = bson.M{
			"$search": ts.SearchText,
		}
	}

	options.SetSort(bson.D{{"symbol", 1}})

	if len(ts.Sectors) > 0 {
		field = "sector"
		query[field] = bson.M{
			"$in": ts.Sectors,
		}
		options.SetSort(bson.D{{"marketCap", -1}})
	}

	if len(ts.Industries) > 0 {
		field = "industry"
		query[field] = bson.M{
			"$in": ts.Industries,
		}
		options.SetSort(bson.D{{"marketCap", -1}})
	}

	if ts.FromDiv > 0 {
		query["yield"] = bson.M{
			"$gt": ts.FromDiv,
		}
	}

	// log.Print(ts.Function)
	if strings.Compare(ts.Function, "Top Gainers") == 0 {

		field = "prDiffPerc"
		query[field] = bson.M{
			"$gt": 0,
		}
		options.SetSort(bson.D{{field, -1}})
		options.SetLimit(50)

	} else if strings.Compare(ts.Function, "Top Losers") == 0 {
		field = "prDiffPerc"
		query[field] = bson.M{
			"$lt": 0,
		}
		options.SetLimit(50)

	} else if strings.Compare(ts.Function, "Top Gainers (Ytd)") == 0 {
		field = "performance.YTD.diff"
		query[field] = bson.M{
			"$gt": 0,
		}
		options.SetSort(bson.D{{field, -1}})
		options.SetLimit(50)

	} else if strings.Compare(ts.Function, "Top Losers (Ytd)") == 0 {
		field = "performance.YTD.diff"
		query[field] = bson.M{
			"$lt": 0,
		}
		options.SetLimit(50)
	}

	if strings.Compare(ts.PerfPeriod, "N") != 0 {

		if strings.Compare(ts.PerfPeriod, "1D") == 0 {
			field = "prDiffPerc"
		} else {
			field = "performance." + ts.PerfPeriod + ".diff"
		}
		if ts.FromPerfPerc > 0 {
			query[field] = bson.M{
				// "$and": bson.M{
				"$gt": ts.FromPerfPerc,
				"$lt": ts.ToPerfPerc,
				// },
			}
			options.SetSort(bson.D{{field, -1}})

		} else {
			query[field] = bson.M{
				"$gt": ts.FromPerfPerc,
			}
		}
	}

	if ts.FromRsi > 0 || (ts.ToRsi > 0 && ts.ToRsi < 100) {
		field = "technicals.RSI." + ts.RsiPeriod
		query[field] = bson.M{
			"$gt": ts.FromRsi,
			"$lt": ts.ToRsi,
		}
		options.SetSort(bson.D{{field, -1}})
	}

	// if len(ts.PrAboveMA) > 0 {
	// 	field = "$technicals." + ts.PrAboveMA + "." + ts.PrAbovePeriod
	// 	query["$expr"] = bson.M{
	// 		"$gt": bson.A{"$prLast", field},
	// 	}
	// }

	if len(ts.PrMA) > 0 {
		field = "$technicals." + ts.PrMA + "." + ts.PrPeriod
		if ts.PrAbove {
			query["$expr"] = bson.M{
				"$gt": bson.A{"$prLast", field},
			}
		} else {
			query["$expr"] = bson.M{
				"$lt": bson.A{"$prLast", field},
			}
		}
	}

	tickers := mdb.getTickers(query, options)
	if tickers == nil {
		tickers = []*Ticker{}
	}
	return tickers
}

func (mdb *MongoDB) UpdateTickersEOD(ctx context.Context, tickers Tickers) {

	var operations []mongo.WriteModel

	for _, ticker := range tickers {
		operation := mongo.NewUpdateManyModel()
		operation.SetFilter(bson.M{"symbol": ticker.Symbol})
		update := bson.M{"$set": ticker}
		operation.SetUpdate(update)
		operation.SetUpsert(true)
		operations = append(operations, operation)
	}

	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(true)

	tickersCol := mdb.db.Collection(TICKERScol)

	_, err := tickersCol.BulkWrite(context.TODO(), operations, &bulkOption)
	if err != nil {
		log.Fatal(err)
	}
}

func (mdb *MongoDB) UpdateTickersRealtime(ctx context.Context, tickers Tickers) {

	var operations []mongo.WriteModel

	for _, ticker := range tickers {
		operation := mongo.NewUpdateManyModel()
		operation.SetFilter(bson.M{"symbol": ticker.Symbol})
		update := bson.M{"$set": bson.M{
			"prDate":     ticker.PrDate,
			"prLast":     ticker.PrLast,
			"prPrev":     ticker.PrPrev,
			"prDiffAmt":  ticker.PrDiffAmt,
			"prDiffPerc": ticker.PrDiffPerc,
		},
		}

		operation.SetUpdate(update)
		operation.SetUpsert(true)
		operations = append(operations, operation)
	}

	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(true)

	tickersCol := mdb.db.Collection(TICKERScol)
	_, err := tickersCol.BulkWrite(context.TODO(), operations, &bulkOption)
	if err != nil {
		log.Fatal(err)
	}

}

func (mdb *MongoDB) UpdateTickersHistory(ctx context.Context, tha []*TickerHistory) {

	if len(tha) == 0 {
		return
	}
	var operations []mongo.WriteModel

	for _, th := range tha {
		// log.Println(th.FormatTickerClose())
		operation := mongo.NewUpdateManyModel()
		operation.SetFilter(bson.M{"symbol": th.Symbol, "date": th.Date})
		update := bson.M{"$set": th}
		operation.SetUpdate(update)
		operation.SetUpsert(true)
		operations = append(operations, operation)
	}

	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(true)

	tHistoryCol := mdb.db.Collection(THISTORYcol)
	_, err := tHistoryCol.BulkWrite(context.TODO(), operations, &bulkOption)
	if err != nil {
		log.Fatal(err)
	}

}

func (mdb *MongoDB) UpdateTickersNews(ctx context.Context, tnm map[string][]TickerNews) {

	var operations []mongo.WriteModel
	for _, v := range tnm {
		for x := 0; x < 100 && x < len(v); x++ {
			tn := v[x]
			// if x == 2 {
			// 	log.Println(tn.Title)
			// }
			operation := mongo.NewUpdateManyModel()
			operation.SetFilter(bson.M{"symbol": tn.Symbol, "seq": x})
			update := bson.M{"$set": tn}
			operation.SetUpdate(update)
			operation.SetUpsert(true)
			operations = append(operations, operation)

		}
	}
	bulkOption := options.BulkWriteOptions{}
	tNewsCol := mdb.db.Collection(TNEWScol)
	_, err := tNewsCol.BulkWrite(context.TODO(), operations, &bulkOption)
	if err != nil {
		log.Fatal(err)
	}
	// log.Printf("%v", result)

}
