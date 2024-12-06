package es

import (
	"app-store-server/internal/gitapp"
	"app-store-server/pkg/models"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/some"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/golang/glog"
)

func existIndex() bool {
	exists, err := esClient.typedClient.Indices.Exists(indexName).IsSuccess(context.Background())
	if exists {
		return true
	}
	if err != nil {
		glog.Warningf("index %s Exists err:%s", indexName, err.Error())
	}

	return false
}

func delIndex() {
	suc, err := esClient.typedClient.Indices.Delete(indexName).IsSuccess(context.Background())
	if err != nil {
		glog.Warningf("index %s Delete err:%s", indexName, err.Error())
	}

	if !suc {
		glog.Warningf("index %s Delete failed", indexName)
	}
}

func createIndex() error {
	props := map[string]types.Property{
		"history": types.ObjectProperty{
			Properties: map[string]types.Property{
				"latest": types.ObjectProperty{
					Properties: map[string]types.Property{
						"name": types.TextProperty{
							Fields: map[string]types.Property{
								"keyword": types.KeywordProperty{},
							},
							Analyzer:       some.String("caseSensitive"),
							SearchAnalyzer: some.String("caseSensitiveSearch"),
						},
						"title": types.TextProperty{
							Fields: map[string]types.Property{
								"keyword": types.KeywordProperty{},
							},
							Analyzer:       some.String("caseSensitive"),
							SearchAnalyzer: some.String("caseSensitiveSearch"),
						},
						"desc": types.TextProperty{
							Fields: map[string]types.Property{
								"keyword": types.KeywordProperty{},
							},
							Analyzer:       some.String("caseSensitive"),
							SearchAnalyzer: some.String("caseSensitiveSearch"),
						},
						"fullDescription": types.TextProperty{
							Fields: map[string]types.Property{
								"keyword": types.KeywordProperty{},
							},
							Analyzer:       some.String("caseSensitive"),
							SearchAnalyzer: some.String("caseSensitiveSearch"),
						},
						"upgradeDescription": types.TextProperty{
							Fields: map[string]types.Property{
								"keyword": types.KeywordProperty{},
							},
							Analyzer:       some.String("caseSensitive"),
							SearchAnalyzer: some.String("caseSensitiveSearch"),
						},
						"submitter": types.TextProperty{
							Fields: map[string]types.Property{
								"keyword": types.KeywordProperty{},
							},
							Analyzer:       some.String("caseSensitive"),
							SearchAnalyzer: some.String("caseSensitiveSearch"),
						},
						"developer": types.TextProperty{
							Fields: map[string]types.Property{
								"keyword": types.KeywordProperty{
									//Normalizer: some.String(""),
								},
							},
							Analyzer:       some.String("caseSensitive"),
							SearchAnalyzer: some.String("caseSensitiveSearch"),
						},
						"categories":     types.TextProperty{},
						"lastCommitHash": types.KeywordProperty{},
						"createTime":     types.DateProperty{},
						"updateTime":     types.DateProperty{},
					},
				},
			},
		},
	}
	err := esClient.CreateIndexWithMapping(indexName, props)
	if err != nil {
		glog.Warningf("createIndex err:%s", err.Error())
	}

	return err
}

func UpsertAppInfoToDb(appInfo *models.ApplicationInfoFullData) error {
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
	lastCommitHash, err := gitapp.GetLastHash()
	if err != nil {
		glog.Warningf("GetLastHash error: %s", err.Error())
		return
	}

	totalCategoriesAgg, err := esClient.typedClient.Search().
		Index(indexName).
		Request(
			&search.Request{
				Size: some.Int(0),
				Query: &types.Query{
					Bool: &types.BoolQuery{
						Filter: []types.Query{
							{
								Term: map[string]types.TermQuery{
									"history.latest.lastCommitHash": {Value: lastCommitHash}},
							},
						},
					},
				},
				Aggregations: map[string]types.Aggregations{
					"history.latest.categories": {
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

func SearchByCategory(from, size int, category string) (infos []*models.ApplicationInfoFullData, err error) {
	var resp *search.Response
	var lastCommitHash string
	lastCommitHash, err = gitapp.GetLastHash()
	if err != nil {
		glog.Warningf("GetLastHash error: %s", err.Error())
		return
	}

	resp, err = esClient.typedClient.Search().
		Index(indexName).
		Request(
			&search.Request{
				Size: some.Int(size),
				From: some.Int(from),
				Query: &types.Query{
					Bool: &types.BoolQuery{
						Filter: []types.Query{
							{
								Term: map[string]types.TermQuery{
									"history.latest.lastCommitHash": {Value: lastCommitHash}},
							},
							{
								Term: map[string]types.TermQuery{
									"history.latest.categories": {Value: category},
								},
							},
						},
					},
				},
				Sort: []types.SortCombinations{
					types.SortOptions{SortOptions: map[string]types.FieldSort{
						"history.latest.updateTime":   {Order: &sortorder.Desc},
						"history.latest.name.keyword": {Order: &sortorder.Asc},
					}},
				},
			}).
		Do(context.TODO())
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return
	}

	for _, hit := range resp.Hits.Hits {
		info := &models.ApplicationInfoFullData{}
		err = json.Unmarshal(hit.Source_, info)
		if err != nil {
			continue
		}
		glog.Infof("name:%s, update:%d\n", info.Name, info.History["latest"].UpdateTime)

		infos = append(infos, info)
	}

	return
}

func SearchByNameAccurate(name string) (*models.ApplicationInfoFullData, error) {
	var resp *search.Response
	lastCommitHash, err := gitapp.GetLastHash()
	if err != nil {
		glog.Warningf("GetLastHash error: %s", err.Error())
		return nil, err
	}

	resp, err = esClient.typedClient.Search().
		Index(indexName).
		Request(
			&search.Request{
				Query: &types.Query{
					Bool: &types.BoolQuery{
						Filter: []types.Query{
							{
								Term: map[string]types.TermQuery{
									"history.latest.lastCommitHash": {Value: lastCommitHash}},
							},
							{
								Term: map[string]types.TermQuery{
									"history.latest.name": {Value: name},
								},
							},
						},
					},
				},
				Sort: []types.SortCombinations{
					types.SortOptions{SortOptions: map[string]types.FieldSort{
						"history.latest.updateTime":   {Order: &sortorder.Desc},
						"history.latest.name.keyword": {Order: &sortorder.Asc},
					}},
				},
			}).
		Do(context.TODO())
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return nil, err
	}

	for _, hit := range resp.Hits.Hits {
		info := &models.ApplicationInfoFullData{}
		err = json.Unmarshal(hit.Source_, info)
		if err != nil {
			return nil, err
		}
		glog.Infof("name:%s, update:%d\n", info.Name, info.History["latest"].UpdateTime)

		return info, nil
	}

	return nil, fmt.Errorf("get info failed")
}

func getWildcardName(word string) string {
	if word == "" {
		return "*"
	}

	if word[0] != '*' {
		word = "*" + word
	}

	if word[len(word)-1] != '*' {
		word += "*"
	}

	return strings.ToLower(word)
}

func SearchByNameWildcard(from, size int, name string) (infos []*models.ApplicationInfoFullData, count int64, err error) {
	var resp *search.Response
	var lastCommitHash string
	lastCommitHash, err = gitapp.GetLastHash()
	if err != nil {
		glog.Warningf("GetLastHash error: %s", err.Error())
		return
	}

	wildcardName := getWildcardName(name)
	resp, err = esClient.typedClient.Search().
		Index(indexName).
		Request(
			&search.Request{
				Size: some.Int(size),
				From: some.Int(from),
				Query: &types.Query{
					Bool: &types.BoolQuery{
						Filter: []types.Query{
							{
								Term: map[string]types.TermQuery{
									"history.latest.lastCommitHash": {Value: lastCommitHash}},
							},
							{
								Bool: &types.BoolQuery{
									Should: []types.Query{
										{
											Wildcard: map[string]types.WildcardQuery{
												"history.latest.name": {
													Value:           &wildcardName,
													CaseInsensitive: some.Bool(true),
												},
											},
										},
										{
											Wildcard: map[string]types.WildcardQuery{
												"history.latest.title": {
													Value:           &wildcardName,
													CaseInsensitive: some.Bool(true),
												},
											},
										},
										{
											Wildcard: map[string]types.WildcardQuery{
												"history.latest.desc": {
													Value:           &wildcardName,
													CaseInsensitive: some.Bool(true),
												},
											},
										},
										{
											Wildcard: map[string]types.WildcardQuery{
												"history.latest.fullDescription": {
													Value:           &wildcardName,
													CaseInsensitive: some.Bool(true),
												},
											},
										},
										{
											Wildcard: map[string]types.WildcardQuery{
												"history.latest.upgradeDescription": {
													Value:           &wildcardName,
													CaseInsensitive: some.Bool(true),
												},
											},
										},
										{
											Wildcard: map[string]types.WildcardQuery{
												"history.latest.submitter": {
													Value:           &wildcardName,
													CaseInsensitive: some.Bool(true),
												},
											},
										},
										{
											Wildcard: map[string]types.WildcardQuery{
												"history.latest.developer": {
													Value:           &wildcardName,
													CaseInsensitive: some.Bool(true),
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Sort: []types.SortCombinations{
					types.SortOptions{SortOptions: map[string]types.FieldSort{
						"history.latest.updateTime":   {Order: &sortorder.Desc},
						"history.latest.name.keyword": {Order: &sortorder.Asc},
					}},
				},
			}).
		Do(context.TODO())
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return
	}

	count = resp.Hits.Total.Value

	for _, hit := range resp.Hits.Hits {
		info := &models.ApplicationInfoFullData{}
		err = json.Unmarshal(hit.Source_, info)
		if err != nil {
			continue
		}
		glog.Infof("name:%s, update:%d\n", info.Name, info.History["latest"].UpdateTime)

		infos = append(infos, info)
	}

	return
}
