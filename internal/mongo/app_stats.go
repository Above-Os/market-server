package mongo

import (
	"app-store-server/internal/constants"
	"app-store-server/pkg/models"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Counter struct {
	Name  string `bson:"name"`
	Count int64  `bson:"count"`
}

func InitCounterByApp(name string) error {
	filter := bson.M{"name": name}

	var counter Counter
	err := mgoClient.queryOne(AppStoreDb, AppStatsCollection, filter).Decode(&counter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			newCounter := Counter{
				Name:  name,
				Count: 0,
			}
			_, err = mgoClient.insertOne(AppStoreDb, AppStatsCollection, newCounter)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	return nil
}

func SetAppInstallCount(name string) error {
	counter := &Counter{}
	filter := bson.M{"name": name}
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

func GetTopApps(count int64) ([]string, error) {
	if count <= 0 {
		count = constants.DefaultTopCount
	}
	opts := options.Find().SetSort(bson.M{"count": -1}).SetLimit(count)

	cur, err := mgoClient.queryMany(AppStoreDb, AppStatsCollection, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	defer cur.Close(ctx)

	var names []string
	for cur.Next(ctx) {
		var counter Counter
		err := cur.Decode(&counter)
		if err != nil {
			return nil, err
		}
		names = append(names, counter.Name)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return names, nil
}

func GetTopApplicationInfos(category, ty string, excludedLabels []string, count int) ([]models.ApplicationInfoFullData, error) {
	lastCommitHash, err := GetLastCommitHashFromDB()
	if err != nil {
		return nil, err
	}

	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         AppStatsCollection,
				"localField":   "name",
				"foreignField": "name",
				"as":           "counter",
			},
		},
		{
			"$unwind": "$counter",
		},
		{
			"$addFields": bson.M{
				"count": "$counter.count",
			},
		},
	}
	filter := make(bson.M)
	if lastCommitHash != "" {
		filter["history.latest.lastCommitHash"] = lastCommitHash
	}
	if ty != "" {
		tys := strings.Split(ty, ",")
		if len(tys) > 1 {
			filter["history.latest.cfgType"] = bson.M{"$in": tys}
		} else {
			filter["history.latest.cfgType"] = ty
		}
	}

	if len(excludedLabels) > 0 {
		filter["appLabels"] = bson.M{
			"$nin": excludedLabels,
		}
	}

	if category != "" {
		categoriesRegex := bson.M{
			"$regex": primitive.Regex{Pattern: fmt.Sprintf("^%s$", category), Options: "i"},
		}
		filter["history.latest.categories"] = bson.M{
			"$elemMatch": categoriesRegex,
		}
	}

	pipeline = append(pipeline,
		bson.M{
			"$match": filter,
		},
	)

	if count <= 0 {
		count = constants.DefaultTopCount
	}
	pipeline = append(pipeline,
		bson.M{
			"$sort": bson.D{
				bson.E{Key: "history.latest.counter.count", Value: -1},
				bson.E{Key: "history.latest.updateTime", Value: -1},
				bson.E{Key: "name", Value: 1},
			},
		},
		bson.M{
			"$limit": count,
		},
	)
	appInfoCollection := mgoClient.mgo.Database(AppStoreDb).Collection(AppInfosCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cursor, err := appInfoCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var applicationInfos []models.ApplicationInfoFullData
	if err := cursor.All(ctx, &applicationInfos); err != nil {
		return nil, err
	}

	return applicationInfos, nil
}

func GetAppInstallCount(name string) (int64, error) {
	result := &Counter{}
	filter := bson.M{"name": name}
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
		filter["name"] = bson.M{"$in": names}
	}

	sort := bson.D{
		bson.E{Key: "history.latest.count", Value: -1},
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
