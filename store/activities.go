package store

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

//Activity holds meta data for all activities
type Activity struct {
	UID         string             `json:"-"`
	ID          primitive.ObjectID `bson:"_id"`
	Date        *time.Time         `json:"date" bson:"date"`
	ActyType    string             `json:"actyType" bson:"actyType"`
	Group       string             `json:"group" bson:"group"`
	Category    string             `json:"category" bson:"category"`
	Account     string             `json:"account" bson:"account"`
	Description string             `json:"description" bson:"description"`
	TxnType     string             `json:"txnType" bson:"txnType"`
	Dbcr        string             `json:"dbcr" bson:"dbcr"`
	Amount      float64            `json:"amount" bson:"amount"`
	Symbol      string             `json:"symbol" bson:"symbol"`
	Qty         float64            `json:"qty" bson:"qty"`
	Price       float64            `json:"price" bson:"price"`
	ToAccount   string             `json:"toAccount" bson:"toAccount"`
	Fee         float64            `json:"fee" bson:"fee"`
}

//Activities holds an array of activity.
type Activities []*Activity

func createActivitiesIndices(ctx context.Context, col *mongo.Collection) {

	keys := bsonx.Doc{{Key: "UID", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_UID", keys, false)

	keys = bsonx.Doc{{Key: "actyType", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_actyType", keys, false)

	keys = bsonx.Doc{{Key: "group", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_group", keys, false)

	keys = bsonx.Doc{{Key: "category", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_category", keys, false)

	keys = bsonx.Doc{{Key: "account", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_account", keys, false)

	keys = bsonx.Doc{{Key: "txnType", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_txnType", keys, false)

	keys = bsonx.Doc{{Key: "date", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_date", keys, false)

	keys = bsonx.Doc{{Key: "dbcr", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_dbcr", keys, false)

}

func (mdb *MongoDB) DropActivitiesCollection(ctx context.Context) {
	actvsCol := mdb.db.Collection(ACTVScol)
	actvsCol.Drop(ctx)
}

//DeleteActivities deletes activities by activity Type
func (mdb *MongoDB) DeleteActivities(ctx context.Context, actyType string, group string, category string, fromDate *time.Time, toDate *time.Time) {

	actvsCol := mdb.db.Collection(ACTVScol)
	var query map[string]interface{}
	query = make(map[string]interface{})
	query["actyType"] = bson.M{"$eq": actyType}
	if len(group) > 0 {
		query["group"] = bson.M{"$eq": group}
	}
	if len(category) > 0 {
		query["category"] = bson.M{"$eq": category}
	}

	if !fromDate.IsZero() && !toDate.IsZero() {
		query["date"] = bson.M{"$gte": fromDate, "$lte": toDate}
	} else if !fromDate.IsZero() {
		query["date"] = bson.M{"$gte": fromDate}
	} else if !toDate.IsZero() {
		query["date"] = bson.M{"$le": toDate}
	}

	log.Printf("Delete Activity query: %v", query)
	result, err := actvsCol.DeleteMany(ctx, query)
	if err != nil {
		log.Printf("Delete activities error: %v", err)
		return
	}
	log.Printf("Deleted count: %d", result.DeletedCount)
}

//ActivitiesUpdate updates activities
func (mdb *MongoDB) ActivitiesUpdate(ctx context.Context, actvs Activities) error {

	user := UserFromCtx(ctx)
	if len(actvs) == 0 {
		return nil
	}

	var operations []mongo.WriteModel

	for _, actv := range actvs {
		// log.Println(th.FormatTickerClose())
		if actv.ID.IsZero() {
			actv.ID = primitive.NewObjectID()
		}
		// log.Printf("Date: %v type: %s Qty: %f Price: %f \n", actv.Date, actv.Type, actv.Qty, actv.Price)
		// if actv.Sales == nil {
		// 	actv.Sales = []SaleLot{}
		// }

		actv.UID = user.UID
		operation := mongo.NewUpdateManyModel()
		operation.SetFilter(bson.M{"UID": actv.UID, "_id": actv.ID})
		update := bson.M{"$set": actv}
		operation.SetUpdate(update)
		operation.SetUpsert(true)
		operations = append(operations, operation)
	}

	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(true)
	col := mdb.db.Collection(ACTVScol)
	_, err := col.BulkWrite(context.TODO(), operations, &bulkOption)
	return err

}

//InvAccounts returns investment accounts
func (mdb *MongoDB) InvestmentsAccounts(ctx context.Context) (InvAccounts, error) {

	user := UserFromCtx(ctx)

	var pipeline []interface{}
	var match map[string]interface{}
	match = make(map[string]interface{})
	match["UID"] = bson.M{"$eq": user.UID}
	match["actyType"] = "Investment"

	query := bson.M{
		"_id": bson.M{
			"group":    "$group",
			"category": "$category",
			"account":  "$account",
		},
	}

	queryStage := bson.M{
		"$group": query,
	}

	matchStage := bson.M{
		"$match": match,
	}

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

	// log.Println(results)
	return nil, nil
}

//InvestmentsOpenActivities returns all investment activities
func (mdb *MongoDB) InvestmentsActivities(ctx context.Context, actv *Activity, open bool, asc bool) Activities {

	// var posns InvPositions
	var options = options.Find()
	var query map[string]interface{}
	query = make(map[string]interface{})

	user := UserFromCtx(ctx)
	query["UID"] = bson.M{"$eq": user.UID}

	if actv != nil {
		// if strings.Compare(actv.ActyType, "Investment") != 0 {
		// 	return nil
		// }
		query["actyType"] = bson.M{"$eq": actv.ActyType}
		if len(actv.Group) > 0 {
			query["group"] = bson.M{"$eq": actv.Group}
		}
		if len(actv.Category) > 0 {
			query["category"] = bson.M{"$eq": actv.Category}
		}
		if len(actv.Account) > 0 {
			query["account"] = bson.M{"$eq": actv.Account}
		}
		if len(actv.Symbol) > 0 {
			query["symbol"] = bson.M{"$eq": actv.Symbol}
		}

	} else {
		query["actyType"] = bson.M{"$eq": "Investment"}
	}

	var ins []string

	if open {
		ins = append(ins, "Buy")
		ins = append(ins, "Receive")
		ins = append(ins, "Rewards")
		query["txnType"] = bson.M{"$in": ins}

		query["qty"] = bson.M{"$ne": 0}
	}

	var sv = 1
	if !asc {
		sv = -1
	}
	options.SetSort(bson.D{{"date", sv}})

	return mdb.getActivities(query, options)
}

func (mdb *MongoDB) getActivities(query interface{}, ops *options.FindOptions) Activities {

	if ops == nil {
		ops = options.Find()
	}

	var result Activities

	actvsCol := mdb.db.Collection(ACTVScol)
	cur, err := actvsCol.Find(context.TODO(), query, ops)

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		err = cur.All(context.TODO(), &result)
		if err != nil {
			log.Printf("Cursor error: %v\n", err)
		}
	}
	// if result != nil {
	// log.Printf("%v\n", result)
	// }

	return result
}
