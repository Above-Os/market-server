package es

import (
	"app-store-server/pkg/models"
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/golang/glog"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/some"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

func existIndex() bool {
	exists, err := esClient.typedClient.Indices.Exists(indexName).IsSuccess(context.Background())
	if !exists && err == nil {
		return false
	}
	if err != nil {
		glog.Warningf("index %s Exists err:%s", indexName, err.Error())
	}

	return true
}

func createIndex() error {
	props := map[string]types.Property{
		"name":           types.KeywordProperty{},
		"categories":     types.KeywordProperty{},
		"lastCommitHash": types.KeywordProperty{},
		"createTime":     types.DateProperty{},
		"updateTime":     types.DateProperty{},
	}
	err := esClient.CreateIndexWithMapping(indexName, props)
	if err != nil {
		glog.Warningf("createIndex err:%s", err.Error())
	}

	return err
}

func UpsertAppInfoToDb(appInfo *models.ApplicationInfo) error {
	resp, err := esClient.typedClient.Index(indexName).Id(appInfo.Id).Request(appInfo).Do(context.TODO())
	if err != nil {
		glog.Warningf("resp:%+v, err:%s", resp, err.Error())
		return err
	}

	return nil
}

//UpsertAppInfosToDb todo bulk
//func UpsertAppInfosToDb(appInfos []*models.ApplicationInfo) error {
//
//}

func GetCategories() (categories []string) {
	totalCategoriesAgg, err := esClient.typedClient.Search().
		Index(indexName).
		Request(
			&search.Request{
				Size: some.Int(0),
				Aggregations: map[string]types.Aggregations{
					"categories": {
						Terms: &types.TermsAggregation{
							Field: some.String("categories"),
						},
					},
				},
			},
		).Do(context.Background())

	if err != nil {
		glog.Warningf("GetCategories error:%s", err.Error())
		return
	}
	glog.Infof("totalCategoriesAgg:%#v, err:%#v\n", totalCategoriesAgg, err)

	aggs, exist := totalCategoriesAgg.Aggregations["categories"]
	if !exist {
		return
	}

	for _, bucket := range aggs.(*types.StringTermsAggregate).Buckets.([]types.StringTermsBucket) {
		categories = append(categories, bucket.Key.(string))
	}

	return
}

func SearchByCategory(from, size int, category string) (infos []*models.ApplicationInfo, err error) {
	var resp *search.Response
	resp, err = esClient.typedClient.Search().
		Index(indexName).
		Request(
			&search.Request{
				Size: some.Int(size),
				From: some.Int(from),
				Query: &types.Query{
					Match: map[string]types.MatchQuery{
						"categories": {Query: category},
					},
				},
				Sort: []types.SortCombinations{
					types.SortOptions{SortOptions: map[string]types.FieldSort{
						"updateTime": {Order: &sortorder.Desc},
						"name":       {Order: &sortorder.Asc},
					}},
				},
			}).
		Do(context.TODO())

	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return
	}

	for _, hit := range resp.Hits.Hits {
		info := &models.ApplicationInfo{}
		err = json.Unmarshal(hit.Source_, info)
		if err != nil {
			continue
		}
		glog.Infof("name:%s, update:%d\n", info.Name, info.UpdateTime)

		infos = append(infos, info)
	}

	return
}

func SearchByName(from, size int, name string) (infos []*models.ApplicationInfo, count int64, err error) {
	var resp *search.Response
	resp, err = esClient.typedClient.Search().
		Index(indexName).
		Request(
			&search.Request{
				Size: some.Int(size),
				From: some.Int(from),
				Query: &types.Query{
					Match: map[string]types.MatchQuery{
						"name": {Query: name},
					},
				},
				Sort: []types.SortCombinations{
					types.SortOptions{SortOptions: map[string]types.FieldSort{
						"updateTime": {Order: &sortorder.Desc},
						"name":       {Order: &sortorder.Asc},
					}},
				},
			}).
		Do(context.TODO())

	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return
	}

	for _, hit := range resp.Hits.Hits {
		info := &models.ApplicationInfo{}
		err = json.Unmarshal(hit.Source_, info)
		if err != nil {
			continue
		}
		glog.Infof("name:%s, update:%d\n", info.Name, info.UpdateTime)

		infos = append(infos, info)
	}

	return
}
