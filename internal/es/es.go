package es

import (
	"app-store-server/internal/constants"
	"app-store-server/internal/mongo"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	es8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/update"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/golang/glog"
)

type Client struct {
	typedClient *es8.TypedClient
}

var esClient *Client

const indexName = "app_info"

func Init() error {
	addr := os.Getenv(constants.EsAddr)         //https://localhost:9200
	username := os.Getenv(constants.EsName)     //elastic
	password := os.Getenv(constants.EsPassword) //WVF+CRh+oHV+J8ZTV4lC

	err := initWithParams(addr, username, password)
	if err != nil {
		glog.Warningf("initWithParams err:%s", err.Error())
		return err
	}

	if !existIndex() {
		createIndex()
	}

	go syncAppInfosLoop()

	return nil
}

func syncAppInfosLoop() {
	for {
		err := syncAppInfosFromMongoToEs()
		if err != nil {
			glog.Warningf("syncAppInfosFromMongoToEs err:%s", err.Error())
		}
		time.Sleep(time.Duration(1) * time.Minute)
	}
}

func syncAppInfosFromMongoToEs() error {
	for page := int64(1); ; page++ {
		infos, _, err := mongo.GetAppListsFromDb(page, 1000, "")
		if err != nil {
			glog.Warningf("GetAppListsFromDb err:%s", err.Error())
			break
		}
		glog.Infof("success get %d docs from mongodb", len(infos))

		for _, info := range infos {
			err = UpsertAppInfoToDb(info)
			if err != nil {
				glog.Warningf("UpsertAppInfoToDb err:%s", err.Error())
				continue
			}
		}

		if len(infos) < 1000 {
			break
		}
	}
	return nil
}

func initWithParams(addr, username, password string) error {
	addresses := strings.Split(addr, ",")

	config := es8.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				//RootCAs:            rootCAs,
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	typedClient, err := es8.NewTypedClient(config)

	if err != nil {
		glog.Fatalf("es8.NewClient err:%s", err.Error())
		return err
	}

	esClient = &Client{
		//apiClient:   client,
		typedClient: typedClient,
	}

	suc, err := esClient.typedClient.Ping().IsSuccess(context.TODO())
	if !suc {
		return fmt.Errorf("ping failed, err:%s", err.Error())
	}

	return nil
}

func (c *Client) CreateIndex(indexName string) error {
	response, err := c.typedClient.Indices.Create(indexName).Do(context.TODO())
	if err != nil {
		glog.Warningf("CreateIndex indexName:%s, err:%s", indexName, err.Error())
		return err
	}
	glog.Infof("response:%+v", response)

	return nil
}

func (c *Client) CreateIndexWithMapping(indexName string, prop map[string]types.Property) error {
	response, err := c.typedClient.Indices.Create(indexName).
		Mappings(&types.TypeMapping{
			Properties: prop,
		}).
		Do(nil)
	if err != nil {
		glog.Warningf("es8.CreateIndex indexName:%s with map err:%s", indexName, err.Error())
		return err
	}
	glog.Infof("response:%+v", response)

	return nil
}

func (c *Client) AddDocWithID(indexName, id string, doc any) error {
	response, err := c.typedClient.Index(indexName).
		Id(id).
		Request(doc).
		Do(context.TODO())
	if err != nil {
		glog.Warningf("es8.AddDoc response:%+v,err:%s", response, err.Error())
		return err
	}
	glog.Infof("response:%+v", response)

	return nil
}

func (c *Client) GetDocById(indexName, id string) (string, error) {
	response, err := c.typedClient.Get(indexName, id).Do(context.TODO())

	if err != nil {
		glog.Warningf("es8.Get err:%s", err.Error())
		return "", err
	}
	glog.Infof("response:%+v", response)

	return string(response.Source_), nil
}

func (c *Client) DelOneDoc(indexName, id string) error {
	response, err := c.typedClient.Delete(indexName, id).Do(context.TODO())
	if err != nil {
		glog.Warningf("es8.Delete err:%s", err.Error())
		return err
	}

	glog.Infof("response:%+v", response)

	return nil
}

func (c *Client) UpdateOneDoc(indexName, id, doc string) error {
	response, err := c.typedClient.Update(indexName, id).
		Request(&update.Request{Doc: json.RawMessage(doc)}).
		Do(context.TODO())

	if err != nil {
		glog.Warningf("es8.Update err:%s", err.Error())
		return err
	}
	glog.Infof("response:%+v", response)

	return nil
}
