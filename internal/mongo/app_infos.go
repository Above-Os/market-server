package mongo

import (
	"app-store-server/pkg/models"
	"context"
	"time"

	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAppListsFromDb(offset, size int64, category string) (list []*models.ApplicationInfo, count int64, err error) {
	filter := make(bson.M)
	if category != "" {
		filter["categories"] = category
	}

	var lastCommitHash string
	lastCommitHash, err = GetLastCommitHashFromDB()
	if err != nil {
		return
	}
	if lastCommitHash != "" {
		filter["lastCommitHash"] = lastCommitHash
	}

	sort := bson.D{
		bson.E{Key: "updateTime", Value: -1},
		bson.E{Key: "name", Value: 1},
	}

	findOpts := options.Find()
	findOpts.SetSort(sort).SetSkip(offset).SetLimit(size)

	var cur *mongo.Cursor
	cur, err = mgoClient.queryMany(AppStoreDb, AppInfosCollection, filter, findOpts)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		// To decode into a struct, use cursor.Decode()
		result := models.ApplicationInfo{}
		err := cur.Decode(&result)
		if err != nil {
			glog.Warningf("err:%s", err.Error())
			continue
		}
		list = append(list, &result)
	}

	count, err = mgoClient.count(AppStoreDb, AppInfosCollection, filter)
	if err = cur.Err(); err != nil {
		glog.Warningf("err:%s", err.Error())
		return
	}

	return
}

func UpsertAppInfoToDb(appInfo *models.ApplicationInfo) error {
	filter := bson.M{"name": appInfo.Name}
	updatedDocument := &models.ApplicationInfo{}
	update := getUpdates(appInfo)
	u := bson.M{"$set": update}
	opts := options.FindOneAndUpdate().SetUpsert(true)

	err := mgoClient.findOneAndUpdate(AppStoreDb, AppInfosCollection, filter, u, opts).Decode(updatedDocument)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
	}
	if err == mongo.ErrNoDocuments {
		return nil
	}
	return err
}

func getUpdates(appInfoNew *models.ApplicationInfo) *bson.M {
	update := bson.M{}
	update["lastCommitHash"] = appInfoNew.LastCommitHash
	update["updateTime"] = appInfoNew.UpdateTime
	update["createTime"] = appInfoNew.CreateTime

	update["chartName"] = appInfoNew.ChartName
	update["icon"] = appInfoNew.Icon
	update["desc"] = appInfoNew.Description
	update["appid"] = appInfoNew.AppID
	update["title"] = appInfoNew.Title
	update["version"] = appInfoNew.Version
	update["categories"] = appInfoNew.Categories
	update["versionName"] = appInfoNew.VersionName
	update["fullDescription"] = appInfoNew.FullDescription
	update["upgradeDescription"] = appInfoNew.UpgradeDescription
	update["promoteImage"] = appInfoNew.PromoteImage
	update["promoteVideo"] = appInfoNew.PromoteVideo
	update["subCategory"] = appInfoNew.SubCategory
	update["developer"] = appInfoNew.Developer
	update["requiredMemory"] = appInfoNew.RequiredMemory
	update["requiredDisk"] = appInfoNew.RequiredDisk
	update["supportClient"] = appInfoNew.SupportClient
	update["requiredGpu"] = appInfoNew.RequiredGPU
	update["requiredCpu"] = appInfoNew.RequiredCPU
	update["rating"] = appInfoNew.Rating
	update["target"] = appInfoNew.Target
	update["permission"] = appInfoNew.Permission
	update["entrance"] = appInfoNew.Entrance

	return &update
}
