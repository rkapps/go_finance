package store

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// //RTransaction holds metadata for imported transaction
// type RTransaction struct {
// 	// Date time.Time `json:"date"`
// 	Date        MyDate  `json:"date"`
// 	Group       string  `json:"group"`
// 	Category    string  `json:"category"`
// 	Account     string  `json:"account"`
// 	Amount      float64 `json:"amount"`
// 	TxnType     string  `json:"txnType"`
// 	Description string  `json:"description"`
// }

// //MyDate holds the time
// type MyDate struct {
// 	time.Time
// }

// //Transaction holds metadata for a transaction
// type Transaction struct {
// 	ID          primitive.ObjectID `bson:"_id"`
// 	Date        time.Time          `json:"date"`
// 	Group       string             `json:"group"`
// 	Category    string             `json:"category"`
// 	Account     string             `json:"account"`
// 	Amount      float64            `json:"amount"`
// 	Type        string             `json:"type"`
// 	Description string             `json:"description"`
// }

//TransactionAgg holds metadata for aggregate data of transactions
type TransactionAgg struct {
	Year     int32   `json:"year"`
	Month    int32   `json:"month"`
	Group    string  `json:"group"`
	Category string  `json:"category"`
	Account  string  `json:"account"`
	Amount   float64 `json:"amount"`
}

// //Transactions holds an array of transactions
// type Transactions []*Transaction

// func (d *MyDate) UnmarshalJSON(b []byte) error {
// 	s := strings.Trim(string(b), "\"")
// 	if strings.Compare("null", s) == 0 {
// 		return nil
// 	}
// 	t, err := time.Parse("2006-01-02T15:04:05.999999999", s)
// 	if err != nil {
// 		return err
// 	}
// 	d.Time = t
// 	return nil
// }

// //ConvertToTransaction converts imported transaction
// func ConvertToTransaction(rtxn RTransaction) *Transaction {
// 	txn := &Transaction{}
// 	txn.Date = rtxn.Date.Time
// 	txn.Group = rtxn.Group
// 	txn.Category = rtxn.Category
// 	txn.Account = rtxn.Account
// 	txn.Amount = rtxn.Amount
// 	txn.Type = rtxn.TxnType
// 	txn.Description = rtxn.Description
// 	// fmt.Printf("Merchant: %s Amount: %f\n", txn.Merchant, txn.Amount)
// 	return txn
// }

//AggregateTransactions groups and aggreates the amount
func (mdb *MongoDB) AggregateTransactions(ctx context.Context, fromDate *time.Time, toDate *time.Time) ([]TransactionAgg, error) {
	var pipeline []interface{}
	// pipeline = make(map[string]interface{})
	log.Printf("Transactions from: %v to: %v", fromDate, toDate)
	var match map[string]interface{}
	match = make(map[string]interface{})

	user := UserFromCtx(ctx)
	match["UID"] = bson.M{"$eq": user.UID}
	match["actyType"] = "Transaction"
	match["dbcr"] = "debit"

	if fromDate != nil || toDate != nil {
		if fromDate != nil && toDate != nil {
			match["date"] = bson.M{"$gte": fromDate, "$lte": toDate}
		} else if fromDate != nil {
			// fmt.Println(fromDate)
			match["date"] = bson.M{"$gte": fromDate}
		} else {
			match["date"] = bson.M{"$lte": toDate}
		}
	}

	query := bson.M{
		"_id": bson.M{
			"year":     bson.M{"$year": "$date"},
			"month":    bson.M{"$month": "$date"},
			"group":    "$group",
			"category": "$category",
			"account":  "$account",
		},
		"amount": bson.M{"$sum": "$amount"},
	}

	queryStage := bson.M{
		"$group": query,
	}

	matchStage := bson.M{
		"$match": match,
	}

	log.Println(match)

	pipeline = append(pipeline, matchStage, queryStage)
	col := mdb.db.Collection(ACTVScol)
	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Aggregate error: %v", err)
		return nil, err
	}

	var results []map[string]interface{}
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}

	taggs := convertToTransactionAgg(results)
	return taggs, nil
}

func convertToTransactionAgg(results []map[string]interface{}) []TransactionAgg {

	var taggs []TransactionAgg
	var tagg TransactionAgg
	fmt.Printf("results: %v\n", len(results))
	for _, result := range results {

		// fmt.Printf("result: %v\n", result)
		tagg = TransactionAgg{}
		for _, v := range result {
			switch entry := v.(type) {
			case map[string]interface{}:
				for k, v2 := range entry {
					switch v3 := v2.(type) {
					case string:
					case int32:
						if strings.Compare("month", k) == 0 {
							tagg.Month = v3
						} else if strings.Compare("year", k) == 0 {
							tagg.Year = v3
						}

					default:
						// fmt.Printf("--%T\n", v3)
						// break
					}

				}
				tagg.Group = fmt.Sprintf("%s", entry["group"])
				tagg.Category = fmt.Sprintf("%s", entry["category"])
				tagg.Account = fmt.Sprintf("%s", entry["account"])
			case float64:
				tagg.Amount = entry
			default:
				fmt.Printf("----%T", entry)
			}

		}
		taggs = append(taggs, tagg)
	}

	return taggs
}
