package mongo

import (
	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SetLastCommitHashToDB(hash string) error {
	updatedDocument := &struct {
		LastCommitHash string
	}{}
	update := bson.M{}
	update["lastCommitHash"] = hash
	u := bson.M{"$set": update}
	opts := options.FindOneAndUpdate().SetUpsert(true)

	err := mgoClient.findOneAndUpdate(AppStoreDb, AppGitCollection, bson.D{}, u, opts).Decode(updatedDocument)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
	}
	if err == mongo.ErrNoDocuments {
		return nil
	}
	return err

}

func GetLastCommitHashFromDB() (string, error) {
	result := struct {
		LastCommitHash string
	}{}
	err := mgoClient.queryOne(AppStoreDb, AppGitCollection, bson.D{}).Decode(&result)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return "", err
	}

	return result.LastCommitHash, nil
}
