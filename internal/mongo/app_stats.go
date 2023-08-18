package mongo

import (
	"context"
	"errors"
	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Counter struct {
	ID    string `bson:"_id"`
	Count int64  `bson:"count"`
}

func SetAppInstallCount(name string) error {
	counter := &Counter{}
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"count": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	err := mgoClient.findOneAndUpdate(AppStoreDb, AppStatsCollection, filter, update, opts).Decode(counter)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	if err != nil {
		glog.Warningf("err:%s", err.Error())
	}

	return err
}

func GetAppInstallCount(name string) (int64, error) {
	result := &Counter{}
	filter := bson.M{"_id": name}
	err := mgoClient.queryOne(AppStoreDb, AppStatsCollection, filter).Decode(&result)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return 0, err
	}

	return result.Count, nil
}

func GetAppsInstallCounts(names []string) (list []*Counter, err error) {
	filter := make(bson.M)
	if len(names) > 0 {
		filter["_id"] = bson.M{"$in": names}
	}

	sort := bson.D{
		bson.E{Key: "count", Value: -1},
	}

	findOpts := options.Find().SetSort(sort)

	var cur *mongo.Cursor
	cur, err = mgoClient.queryMany(AppStoreDb, AppStatsCollection, filter, findOpts)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		// To decode into a struct, use cursor.Decode()
		result := &Counter{}
		err := cur.Decode(result)
		if err != nil {
			glog.Warningf("err:%s", err.Error())
			continue
		}
		list = append(list, result)
	}

	return
}
