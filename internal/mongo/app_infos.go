package mongo

import (
	"app-store-server/pkg/models"
	"context"
	"errors"
	"time"

	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAppLists(offset, size int64, category string) (list []*models.ApplicationInfo, count int64, err error) {
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
		result := &models.ApplicationInfo{}
		err := cur.Decode(result)
		if err != nil {
			glog.Warningf("err:%s", err.Error())
			continue
		}
		list = append(list, result)
	}

	count, err = mgoClient.count(AppStoreDb, AppInfosCollection, filter)
	if err = cur.Err(); err != nil {
		glog.Warningf("err:%s", err.Error())
		return
	}

	return
}

func GetAppInfos(names []string) (mapInfo map[string]*models.ApplicationInfo, err error) {
	filter := make(bson.M)
	if len(names) > 0 {
		filter["name"] = bson.M{"$in": names}
	}

	var lastCommitHash string
	lastCommitHash, err = GetLastCommitHashFromDB()
	if err != nil {
		return
	}
	if lastCommitHash != "" {
		filter["lastCommitHash"] = lastCommitHash
	}

	cur, err := mgoClient.queryMany(AppStoreDb, AppInfosCollection, filter)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	defer cur.Close(ctx)

	mapInfo = make(map[string]*models.ApplicationInfo)
	for cur.Next(ctx) {
		// To decode into a struct, use cursor.Decode()
		result := &models.ApplicationInfo{}
		err := cur.Decode(result)
		if err != nil {
			glog.Warningf("err:%s", err.Error())
			continue
		}
		mapInfo[result.Name] = result
	}

	return
}

func GetAppInfoByName(name string) (*models.ApplicationInfo, error) {
	filter := bson.M{"name": name}
	info := &models.ApplicationInfo{}
	err := mgoClient.queryOne(AppStoreDb, AppInfosCollection, filter).Decode(&info)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return nil, err
	}

	return info, nil
}

func UpsertAppInfoToDb(appInfo *models.ApplicationInfo) error {
	filter := bson.M{"name": appInfo.Name}
	updatedDocument := &models.ApplicationInfo{}
	update := getUpdates(appInfo)
	u := bson.M{"$set": update}
	opts := options.FindOneAndUpdate().SetUpsert(true)

	err := mgoClient.findOneAndUpdate(AppStoreDb, AppInfosCollection, filter, u, opts).Decode(updatedDocument)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	if err != nil {
		glog.Warningf("err:%s", err.Error())
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
	update["middleware"] = appInfoNew.Middleware
	update["options"] = appInfoNew.Options
	update["language"] = appInfoNew.Language

	update["submitter"] = appInfoNew.Submitter
	update["doc"] = appInfoNew.Doc
	update["website"] = appInfoNew.Website
	update["license"] = appInfoNew.License
	update["legal"] = appInfoNew.Legal
	update["status"] = appInfoNew.Status

	return &update
}
