package mongo

import (
	"app-store-server/pkg/models"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/bson"
)

func SetTopics(infos []models.CmsTopic) error {
	delResult, err := mgoClient.deleteMany(AppStoreAdminDb, AppTopicsCollection, bson.D{})
	if err != nil {
		glog.Warningf("err:%s", err.Error())
	}
	glog.Infof("delResult:%+v", delResult)

	bsonArray := make([]interface{}, len(infos))
	for i, v := range infos {
		bsonArray[i] = v
	}

	InsertResult, err := mgoClient.insertMany(AppStoreAdminDb, AppTopicsCollection, bsonArray)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return err
	}
	glog.Infof("InsertResult:%+v", InsertResult)

	return nil
}

func GetTopics() (lists []*models.CmsTopic, err error) {
	var cur *mongo.Cursor
	cur, err = mgoClient.queryMany(AppStoreAdminDb, AppTopicsCollection, bson.D{})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		// To decode into a struct, use cursor.Decode()
		result := &models.CmsTopic{}
		err := cur.Decode(result)
		if err != nil {
			glog.Warningf("err:%s", err.Error())
			continue
		}
		lists = append(lists, result)
	}

	return
}

func SetRecommends(infos []models.CmsRecommend) error {
	delResult, err := mgoClient.deleteMany(AppStoreAdminDb, AppRecommendsCollection, bson.D{})
	if err != nil {
		glog.Warningf("err:%s", err.Error())
	}
	glog.Infof("delResult:%+v", delResult)

	bsonArray := make([]interface{}, len(infos))
	for i, v := range infos {
		bsonArray[i] = v
	}

	InsertResult, err := mgoClient.insertMany(AppStoreAdminDb, AppRecommendsCollection, bsonArray)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return err
	}
	glog.Infof("InsertResult:%+v", InsertResult)

	return nil
}

func GetRecommends() (lists []*models.CmsRecommend, err error) {
	var cur *mongo.Cursor
	cur, err = mgoClient.queryMany(AppStoreAdminDb, AppRecommendsCollection, bson.D{})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		// To decode into a struct, use cursor.Decode()
		result := &models.CmsRecommend{}
		err := cur.Decode(result)
		if err != nil {
			glog.Warningf("err:%s", err.Error())
			continue
		}
		lists = append(lists, result)
	}

	return
}

func SetCategoryRecommends(infos []models.CmsCategoryRecommend) error {
	delResult, err := mgoClient.deleteMany(AppStoreAdminDb, AppCategoryRecommendsCollection, bson.D{})
	if err != nil {
		glog.Warningf("err:%s", err.Error())
	}
	glog.Infof("delResult:%+v", delResult)

	bsonArray := make([]interface{}, len(infos))
	for i, v := range infos {
		bsonArray[i] = v
	}

	InsertResult, err := mgoClient.insertMany(AppStoreAdminDb, AppCategoryRecommendsCollection, bsonArray)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return err
	}
	glog.Infof("InsertResult:%+v", InsertResult)

	return nil
}

func GetCategoryRecommends() (lists []*models.CmsCategoryRecommend, err error) {
	var cur *mongo.Cursor
	cur, err = mgoClient.queryMany(AppStoreAdminDb, AppCategoryRecommendsCollection, bson.D{})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		// To decode into a struct, use cursor.Decode()
		result := &models.CmsCategoryRecommend{}
		err := cur.Decode(result)
		if err != nil {
			glog.Warningf("err:%s", err.Error())
			continue
		}
		lists = append(lists, result)
	}

	return
}
