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

//InvLot represents a security lot
type InvLot struct {
	UID         string             `json:"-"`
	ID          primitive.ObjectID `bson:"_id"`
	ActvID      primitive.ObjectID `bson:"actvid"`
	Group       string             `json:"group" bson:"group"`
	Category    string             `json:"category" bson:"category"`
	Account     string             `json:"account" bson:"account"`
	Symbol      string             `json:"symbol" bson:"symbol"`
	Date        *time.Time         `json:"date" bson:"date"`
	TxnType     string             `json:"txnType" bson:"txnType"`
	TxnDate     *time.Time         `json:"txnDate" bson:"txnDate"`
	Status      string             `json:"status" bson:"status"`
	OrigQty     float64            `json:"origQty" bson:"origQty"`
	Qty         float64            `json:"qty" bson:"qty"`
	Cost        float64            `json:"cost" bson:"cost"`
	CostValue   float64            `json:"costValue"`
	SendQty     float64            `json:"sendQty" bson:"sendQty"`
	SendDate    *time.Time         `json:"sendDate" bson:"sendDate"`
	OrigAccount string             `json:"origAccount" bson:"origAccount"`
	SaleQty     float64            `json:"saleQty" bson:"saleQty"`
	SaleDate    *time.Time         `json:"saleDate" bson:"saleDate"`
	SalePrice   float64            `json:"salePrice" bson:"salePrice"`
	SaleValue   float64            `json:"saleValue"`
	Fee         float64            `bson:"fee"`
	PrLast      float64            `json:"prLast"`
	PrDiffAmt   float64            `json:"prDiffAmt"`
	PrDiffPerc  float64            `json:"prDiffPerc"`
	MktValue    float64            `json:"mktValue"`
	Dglamount   float64            `json:"dglAmount"`
	Glamount    float64            `json:"glAmount"`
	Glperc      float64            `json:"glPerc"`
}

//InvLots holds an array of invsale.
type InvLots []*InvLot

//InvHolding represents a security holding
type InvHolding struct {
	Group      string        `json:"group"`
	Category   string        `json:"category"`
	Account    string        `json:"account"`
	Symbol     string        `json:"symbol"`
	Date       *time.Time    `json:"date"`
	Qty        float64       `json:"qty"`
	Cost       float64       `json:"cost"`
	CostValue  float64       `json:"costValue"`
	PrLast     float64       `json:"prLast"`
	PrDiffAmt  float64       `json:"prDiffAmt"`
	PrDiffPerc float64       `json:"prDiffPerc"`
	MktValue   float64       `json:"mktValue"`
	Dglamount  float64       `json:"dglAmount"`
	Glamount   float64       `json:"glAmount"`
	Glperc     float64       `json:"glPerc"`
	Holdings   []*InvHolding `json:"holdings"`
}

//InvHoldings holds an array of holding.
type InvHoldings []*InvHolding

//InvAccount holds accounts by group and category
type InvAccount struct {
	Group    string `json:"group"`
	Category string `json:"category"`
	Account  string `json:"account"`
}

//InvAccounts holds an array of accounts.
type InvAccounts []*InvAccount

func createInvLotIndices(ctx context.Context, col *mongo.Collection) {

	keys := bsonx.Doc{{Key: "UID", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_UID", keys, false)

	keys = bsonx.Doc{{Key: "group", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_group", keys, false)

	keys = bsonx.Doc{{Key: "category", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_category", keys, false)

	keys = bsonx.Doc{{Key: "account", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_account", keys, false)

	keys = bsonx.Doc{{Key: "status", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_status", keys, false)

	keys = bsonx.Doc{{Key: "txnType", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_txnType", keys, false)

	keys = bsonx.Doc{{Key: "txnDate", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_txnDate", keys, false)

	keys = bsonx.Doc{{Key: "date", Value: bsonx.Int32(1)}}
	createIndex(ctx, col, "idx_date", keys, false)

}

func (mdb *MongoDB) DropInvLotCollection(ctx context.Context) {
	actvsCol := mdb.db.Collection(INVLOTScol)
	actvsCol.Drop(ctx)
}

//DeleteActivities deletes activities by activity Type
func (mdb *MongoDB) DeleteInvlots(ctx context.Context, group string, category string, fromDate *time.Time, toDate *time.Time) {

	user := UserFromCtx(ctx)

	invlotsCol := mdb.db.Collection(INVLOTScol)
	var query map[string]interface{}
	query = make(map[string]interface{})

	query["UID"] = bson.M{"$eq": user.UID}

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

	log.Printf("Delete InvLot query: %v", query)
	result, err := invlotsCol.DeleteMany(ctx, query)
	if err != nil {
		log.Printf("Delete lots error: %v", err)
		return
	}
	log.Printf("Deleted count: %d", result.DeletedCount)
}

//InvlotsUpdate updates invlots
func (mdb *MongoDB) InvLotsUpdate(ctx context.Context, lots InvLots) error {

	user := UserFromCtx(ctx)
	if len(lots) == 0 {
		return nil
	}

	var operations []mongo.WriteModel

	for _, lot := range lots {
		// log.Println(th.FormatTickerClose())
		if lot.ID.IsZero() {
			lot.ID = primitive.NewObjectID()
		}
		lot.UID = user.UID
		operation := mongo.NewUpdateManyModel()
		operation.SetFilter(bson.M{"UID": lot.UID, "_id": lot.ID})
		update := bson.M{"$set": lot}
		operation.SetUpdate(update)
		operation.SetUpsert(true)
		operations = append(operations, operation)
	}

	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(true)
	col := mdb.db.Collection(INVLOTScol)
	_, err := col.BulkWrite(context.TODO(), operations, &bulkOption)
	return err

}

//InvestmentsSaleLots returns all sale lots
func (mdb *MongoDB) InvestmentsSaleLots(ctx context.Context, lot *InvLot, fromDate time.Time, toDate time.Time) InvLots {

	// var posns InvPositions
	var options = options.Find()
	var query map[string]interface{}
	query = make(map[string]interface{})

	user := UserFromCtx(ctx)

	log.Printf("Transactions from: %v to: %v", fromDate, toDate)

	query["UID"] = bson.M{"$eq": user.UID}
	/*
		query["actyType"] = bson.M{"$eq": "Investment"}
	*/
	query["status"] = bson.M{"$eq": "C"}
	query["saleQty"] = bson.M{"$gt": 0}

	// log.Printf("User: %v Lot: %v", user.UID, lot)

	if len(lot.Group) > 0 {
		query["group"] = bson.M{"$eq": lot.Group}
	}
	if len(lot.Category) > 0 {
		query["category"] = bson.M{"$eq": lot.Category}
	}

	if len(lot.Account) > 0 {
		query["account"] = bson.M{"$eq": lot.Account}
	}
	if len(lot.Symbol) > 0 {
		query["symbol"] = bson.M{"$eq": lot.Symbol}
	}

	if !fromDate.IsZero() && !toDate.IsZero() {
		query["txnDate"] = bson.M{"$gte": fromDate, "$lte": toDate}
	} else if !fromDate.IsZero() {
		query["txnDate"] = bson.M{"$gte": fromDate}
	} else if !toDate.IsZero() {
		query["txnDate"] = bson.M{"$lte": toDate}
	}

	// log.Println(query)
	options.SetSort(bson.D{{"account", 1}, {"symbol", 1}, {"txnDate", 1}})
	return mdb.getInvLots(query, options)
}

//InvestmentsSaleLots returns all sale lots
func (mdb *MongoDB) InvestmentsRewardsLots(ctx context.Context, lot *InvLot, open bool, fromDate time.Time, toDate time.Time) InvLots {

	// var posns InvPositions
	var options = options.Find()
	var query map[string]interface{}
	query = make(map[string]interface{})

	user := UserFromCtx(ctx)
	query["UID"] = bson.M{"$eq": user.UID}
	query["txnType"] = bson.M{"$eq": "Rewards"}

	// log.Printf("User: %v Lot: %v", user.UID, lot)
	if open {
		query["status"] = bson.M{"$eq": "O"}
	} else {
		query["origAccount"] = bson.M{"$eq": ""}
	}

	if len(lot.Group) > 0 {
		query["group"] = bson.M{"$eq": lot.Group}
	}
	if len(lot.Category) > 0 {
		query["category"] = bson.M{"$eq": lot.Category}
	}
	if len(lot.Account) > 0 {
		query["account"] = bson.M{"$eq": lot.Account}
	}
	if len(lot.Symbol) > 0 {
		query["symbol"] = bson.M{"$eq": lot.Symbol}
	}

	if !fromDate.IsZero() && !toDate.IsZero() {
		query["date"] = bson.M{"$gte": fromDate, "$lte": toDate}
	} else if !fromDate.IsZero() {
		query["date"] = bson.M{"$gte": fromDate}
	} else {
		query["date"] = bson.M{"$lte": toDate}
	}

	// options.SetSort(bson.D{{"date", sv}, {"sendDate", 1}, {"saleDate", 1}})
	options.SetSort(bson.D{{"account", 1}, {"symbol", 1}, {"date", 1}})
	return mdb.getInvLots(query, options)
}

//InvestmentsOpenActivities returns all investment activities
func (mdb *MongoDB) InvestmentsLots(ctx context.Context, lot *InvLot, open bool, asc bool) InvLots {

	// var posns InvPositions
	var options = options.Find()
	var query map[string]interface{}
	query = make(map[string]interface{})

	user := UserFromCtx(ctx)
	query["UID"] = bson.M{"$eq": user.UID}

	// log.Printf("User: %v Lot: %v", user.UID, lot)

	if lot != nil {
		if len(lot.Group) > 0 {
			query["group"] = bson.M{"$eq": lot.Group}
		}
		if len(lot.Category) > 0 {
			query["category"] = bson.M{"$eq": lot.Category}
		}
		if len(lot.Account) > 0 {
			query["account"] = bson.M{"$eq": lot.Account}
		}
		if len(lot.Symbol) > 0 {
			query["symbol"] = bson.M{"$eq": lot.Symbol}
		}

	}

	if open {
		query["status"] = bson.M{"$eq": "O"}
	}

	var sv = 1
	if !asc {
		sv = -1
	}
	// options.SetSort(bson.D{{"date", sv}, {"sendDate", 1}, {"saleDate", 1}})
	options.SetSort(bson.D{{"date", sv}, {"txnDate", sv}})
	return mdb.getInvLots(query, options)
}

func (mdb *MongoDB) getInvLots(query interface{}, ops *options.FindOptions) InvLots {

	if ops == nil {
		ops = options.Find()
	}

	var result InvLots

	invlotsCol := mdb.db.Collection(INVLOTScol)
	cur, err := invlotsCol.Find(context.TODO(), query, ops)

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
