package store

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

const (
	//DBname is the database name
	DBname = "finance"

	//ACTVScol is the collection of activities
	ACTVScol = "activity"

	//INVLOTScol is the collection of inventory lots
	INVLOTScol = "invlot"

	//TICKERScol is the collection tickets
	TICKERScol = "ticker"

	//TICKERScol is the collection tickets
	THISTORYcol = "thistory"

	//TNEWScol is the collection tickets
	TNEWScol = "tnews"
)

//MongoDB defines the structure for the database
type MongoDB struct {
	client *mongo.Client
	ctx    context.Context
	db     *mongo.Database
}

var client *mongo.Client

//NewMongoDB returns mongo database
func NewMongoDB(connStr string) (*MongoDB, error) {
	log.Printf("Mongo Conn: %s", connStr)
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))

	if err != nil {
		log.Printf("Mongo Connection error: %v", err)
		return nil, err
	}

	db := client.Database(DBname)

	createTickersIndices(ctx, db.Collection(TICKERScol))
	createTickerHistoryIndices(ctx, db.Collection(THISTORYcol))
	createTickerNewsIndices(ctx, db.Collection(TNEWScol))
	createActivitiesIndices(ctx, db.Collection(ACTVScol))
	createInvLotIndices(ctx, db.Collection(INVLOTScol))

	mdb := &MongoDB{client: client, ctx: ctx, db: db}

	return mdb, nil
}

func createIndex(ctx context.Context, col *mongo.Collection, name string, keys bsonx.Doc, unique bool) error {

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	index := mongo.IndexModel{}
	index.Keys = keys
	index.Options = options.Index()
	index.Options.SetName(name)
	index.Options.SetUnique(unique)

	_, err := col.Indexes().CreateOne(ctx, index, opts)
	if err != nil {
		log.Printf("CreateIndex error: %v", err)
	}
	return err
}
