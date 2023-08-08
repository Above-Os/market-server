package mongo

import (
	"app-store-server/internal/constants"
	"context"
	"os"
	"time"

	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Client struct {
	mgo *mongo.Client
}

const (
	AppStoreDb         = "AppStore"
	AppTypesCollection = "AppTypes"
	AppInfosCollection = "AppInfos"
	AppGitCollection   = "AppGit"
)

var mgoClient *Client

func Init() {
	mgoClient = NewMongoClient()
}

func NewMongoClient() *Client {
	uri := os.Getenv(constants.MongoDBUri)
	if uri == "" {
		glog.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		glog.Exitln(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		glog.Exitln(err)
	}

	return &Client{client}
}

func (mc *Client) insertOne(db, collection string, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	coll := mc.mgo.Database(db).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return coll.InsertOne(ctx, document, opts...)
}

func (mc *Client) insertMany(db, collection string, docs []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	coll := mc.mgo.Database(db).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return coll.InsertMany(ctx, docs, opts...)
}

func (mc *Client) deleteOne(db, collection string, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	coll := mc.mgo.Database(db).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return coll.DeleteOne(ctx, filter, opts...)
}

func (mc *Client) deleteMany(db, collection string, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	coll := mc.mgo.Database(db).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return coll.DeleteMany(ctx, filter, opts...)
}

func (mc *Client) updateOne(db, collection string, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	coll := mc.mgo.Database(db).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return coll.UpdateOne(ctx, filter, update, opts...)
}

func (mc *Client) updateMany(db, collection string, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	coll := mc.mgo.Database(db).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return coll.UpdateMany(ctx, filter, update, opts...)
}

func (mc *Client) queryOne(db, collection string, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	coll := mc.mgo.Database(db).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return coll.FindOne(ctx, filter, opts...)
}

func (mc *Client) queryMany(db, collection string, filter interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error) {
	coll := mc.mgo.Database(db).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return coll.Find(ctx, filter, opts...)
}

func (mc *Client) count(db, collection string, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	coll := mc.mgo.Database(db).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return coll.CountDocuments(ctx, filter, opts...)
}

func (mc *Client) findOneAndUpdate(db, collection string, filter, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	coll := mc.mgo.Database(db).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return coll.FindOneAndUpdate(ctx, filter, update, opts...)
}
