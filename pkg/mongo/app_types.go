package mongo

import (
	"context"
	"time"

	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetAppTypesFromDb() (appTypes []string, err error) {
	var cur *mongo.Cursor
	cur, err = mgoClient.queryMany(AppStoreDb, AppTypesCollection, bson.D{})
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		// To decode into a struct, use cursor.Decode()
		result := struct {
			AppType string
		}{}
		err := cur.Decode(&result)
		if err != nil {
			glog.Warningf("%s", err.Error())
			continue
		}
		appTypes = append(appTypes, result.AppType)
	}

	if err = cur.Err(); err != nil {
		glog.Warningf("err:%s", err.Error())
		return
	}

	return
}
