package mongo

import (
	"app-store-server/pkg/models"
	"app-store-server/pkg/utils"
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

func GetAppLists(offset, size int64, category, ty string) (list []*models.ApplicationInfoFullData, count int64, err error) {
	filter := make(bson.M)
	if category != "" {
		//filter["categories"] = category
		//regex := primitive.Regex{Pattern: category, Options: "i"}
		//filter["categories"] = bson.M{"$in": bson.A{regex}}
		//filter["categories"] = bson.M{
		//	"$elemMatch": bson.M{
		//		"$eq": category,
		//	},
		//}

		categoriesRegex := bson.M{
			"$regex": primitive.Regex{Pattern: fmt.Sprintf("^%s$", category), Options: "i"},
		}

		filter["history.latest.categories"] = bson.M{
			"$elemMatch": categoriesRegex,
		}
	}
	if ty != "" {
		tys := strings.Split(ty, ",")
		if len(tys) > 1 {
			filter["history.latest.cfgType"] = bson.M{"$in": tys}
		} else {
			filter["history.latest.cfgType"] = ty
		}
	}

	var lastCommitHash string
	lastCommitHash, err = GetLastCommitHashFromDB()
	if err != nil {
		return
	}
	if lastCommitHash != "" {
		filter["history.latest.lastCommitHash"] = lastCommitHash
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
		result := &models.ApplicationInfoFullData{}
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

func GetAppInfos(names []string) (mapInfo map[string]*models.ApplicationInfoFullData, err error) {
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
		filter["history.latest.lastCommitHash"] = lastCommitHash
	}

	cur, err := mgoClient.queryMany(AppStoreDb, AppInfosCollection, filter)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	defer cur.Close(ctx)

	mapInfo = make(map[string]*models.ApplicationInfoFullData)
	for cur.Next(ctx) {
		// To decode into a struct, use cursor.Decode()
		result := &models.ApplicationInfoFullData{}
		err := cur.Decode(result)
		if err != nil {
			glog.Warningf("err:%s", err.Error())
			continue
		}
		mapInfo[result.Name] = result
	}

	return
}

func GetAppInfoByName(name string) (*models.ApplicationInfoFullData, error) {
	filter := bson.M{"name": name}
	info := &models.ApplicationInfoFullData{}
	err := mgoClient.queryOne(AppStoreDb, AppInfosCollection, filter).Decode(&info)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return nil, err
	}

	return info, nil
}

func DisableAppInfoToDb(appInfo *models.ApplicationInfoFullData) error {
	filter := bson.M{"name": appInfo.Name}

	_, err := mgoClient.deleteOne(AppStoreDb, AppInfosCollection, filter)

	return err
}

func UpsertAppInfoToDb(appInfo *models.ApplicationInfoFullData) error {
	filter := bson.M{"name": appInfo.Name}
	updatedDocument := &models.ApplicationInfoFullData{}
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

func getUpdates(appInfoNew *models.ApplicationInfoFullData) *bson.M {

	update := bson.M{}
	latest := bson.M{}
	version := bson.M{}

	latest["lastCommitHash"] = appInfoNew.History["latest"].LastCommitHash
	latest["updateTime"] = appInfoNew.History["latest"].UpdateTime
	latest["createTime"] = appInfoNew.History["latest"].CreateTime

	latest["chartName"] = appInfoNew.History["latest"].ChartName
	latest["cfgType"] = appInfoNew.History["latest"].CfgType
	latest["icon"] = appInfoNew.History["latest"].Icon
	latest["desc"] = appInfoNew.History["latest"].Description
	nameMd58 := utils.Md5String(appInfoNew.Name)[:8]
	latest["appid"] = nameMd58
	latest["id"] = nameMd58
	latest["title"] = appInfoNew.History["latest"].Title
	latest["version"] = appInfoNew.History["latest"].Version
	latest["categories"] = appInfoNew.History["latest"].Categories
	latest["versionName"] = appInfoNew.History["latest"].VersionName
	latest["fullDescription"] = appInfoNew.History["latest"].FullDescription
	latest["upgradeDescription"] = appInfoNew.History["latest"].UpgradeDescription
	latest["promoteImage"] = appInfoNew.History["latest"].PromoteImage
	latest["promoteVideo"] = appInfoNew.History["latest"].PromoteVideo
	latest["subCategory"] = appInfoNew.History["latest"].SubCategory
	latest["developer"] = appInfoNew.History["latest"].Developer
	latest["requiredMemory"] = appInfoNew.History["latest"].RequiredMemory
	latest["requiredDisk"] = appInfoNew.History["latest"].RequiredDisk
	latest["supportClient"] = appInfoNew.History["latest"].SupportClient
	latest["supportArch"] = appInfoNew.History["latest"].SupportArch
	latest["requiredGpu"] = appInfoNew.History["latest"].RequiredGPU
	latest["requiredCpu"] = appInfoNew.History["latest"].RequiredCPU
	latest["rating"] = appInfoNew.History["latest"].Rating
	latest["target"] = appInfoNew.History["latest"].Target
	latest["permission"] = appInfoNew.History["latest"].Permission
	latest["entrances"] = appInfoNew.History["latest"].Entrances
	latest["middleware"] = appInfoNew.History["latest"].Middleware
	latest["options"] = appInfoNew.History["latest"].Options
	latest["locale"] = appInfoNew.History["latest"].Locale
	latest["i18n"] = appInfoNew.History["latest"].I18n

	latest["submitter"] = appInfoNew.History["latest"].Submitter
	latest["doc"] = appInfoNew.History["latest"].Doc
	latest["website"] = appInfoNew.History["latest"].Website
	latest["featuredImage"] = appInfoNew.History["latest"].FeaturedImage
	latest["sourceCode"] = appInfoNew.History["latest"].SourceCode
	latest["license"] = appInfoNew.History["latest"].License
	latest["legal"] = appInfoNew.History["latest"].Legal
	//update["status"] = appInfoNew.Status
	latest["appLabels"] = appInfoNew.History["latest"].AppLabels
	latest["modelSize"] = appInfoNew.History["latest"].ModelSize
	latest["namespace"] = appInfoNew.History["latest"].Namespace
	latest["onlyAdmin"] = appInfoNew.History["latest"].OnlyAdmin

	// version
	version["lastCommitHash"] = appInfoNew.History["latest"].LastCommitHash
	version["updateTime"] = appInfoNew.History["latest"].UpdateTime
	version["createTime"] = appInfoNew.History["latest"].CreateTime

	version["chartName"] = appInfoNew.History["latest"].ChartName
	version["cfgType"] = appInfoNew.History["latest"].CfgType
	version["icon"] = appInfoNew.History["latest"].Icon
	version["desc"] = appInfoNew.History["latest"].Description
	version["appid"] = nameMd58
	version["id"] = nameMd58
	version["title"] = appInfoNew.History["latest"].Title
	version["version"] = appInfoNew.History["latest"].Version
	version["categories"] = appInfoNew.History["latest"].Categories
	version["versionName"] = appInfoNew.History["latest"].VersionName
	version["fullDescription"] = appInfoNew.History["latest"].FullDescription
	version["upgradeDescription"] = appInfoNew.History["latest"].UpgradeDescription
	version["promoteImage"] = appInfoNew.History["latest"].PromoteImage
	version["promoteVideo"] = appInfoNew.History["latest"].PromoteVideo
	version["subCategory"] = appInfoNew.History["latest"].SubCategory
	version["developer"] = appInfoNew.History["latest"].Developer
	version["requiredMemory"] = appInfoNew.History["latest"].RequiredMemory
	version["requiredDisk"] = appInfoNew.History["latest"].RequiredDisk
	version["supportClient"] = appInfoNew.History["latest"].SupportClient
	version["supportArch"] = appInfoNew.History["latest"].SupportArch
	version["requiredGpu"] = appInfoNew.History["latest"].RequiredGPU
	version["requiredCpu"] = appInfoNew.History["latest"].RequiredCPU
	version["rating"] = appInfoNew.History["latest"].Rating
	version["target"] = appInfoNew.History["latest"].Target
	version["permission"] = appInfoNew.History["latest"].Permission
	version["entrances"] = appInfoNew.History["latest"].Entrances
	version["middleware"] = appInfoNew.History["latest"].Middleware
	version["options"] = appInfoNew.History["latest"].Options
	version["locale"] = appInfoNew.History["latest"].Locale
	version["i18n"] = appInfoNew.History["latest"].I18n

	version["submitter"] = appInfoNew.History["latest"].Submitter
	version["doc"] = appInfoNew.History["latest"].Doc
	version["website"] = appInfoNew.History["latest"].Website
	version["featuredImage"] = appInfoNew.History["latest"].FeaturedImage
	version["sourceCode"] = appInfoNew.History["latest"].SourceCode
	version["license"] = appInfoNew.History["latest"].License
	version["legal"] = appInfoNew.History["latest"].Legal
	//update["status"] = appInfoNew.Status
	version["appLabels"] = appInfoNew.History["latest"].AppLabels
	version["modelSize"] = appInfoNew.History["latest"].ModelSize
	version["namespace"] = appInfoNew.History["latest"].Namespace
	version["onlyAdmin"] = appInfoNew.History["latest"].OnlyAdmin

	update["history"] = bson.M{
		"latest":                             latest,
		appInfoNew.History["latest"].Version: version,
	}

	update["id"] = nameMd58
	update["appLabels"] = appInfoNew.History["latest"].AppLabels

	return &update
}
