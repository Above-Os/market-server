// Copyright 2022 bytetrade
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang/glog"
)

const (
	AuthorizationTokenKey = "X-Authorization"
)

type Client struct {
	httpClient *http.Client
}

const (
	AppServiceGetURLTempl  = "http://%s:%s/app-service/v1/applications/%s/%s"
	AppServiceListURLTempl = "http://%s:%s/app-service/v1/applications"
	AppSeviceHostEnv       = "APP_SERVICE_SERVICE_HOST"
	AppSevicePortEnv       = "APP_SERVICE_SERVICE_PORT"
)

func newAppServiceClient() *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: time.Second * 2},
	}

	return c
}

func (c *Client) fetchAppListFromAppService(token string) ([]map[string]interface{}, error) {
	appServiceHost := os.Getenv(AppSeviceHostEnv)
	appServicePort := os.Getenv(AppSevicePortEnv)
	urlStr := fmt.Sprintf(AppServiceListURLTempl, appServiceHost, appServicePort)

	return c.doHttpGetList(urlStr, token)
}

func (c *Client) doHttpGetResponse(urlStr, token string) (*http.Response, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: http.MethodGet,
		Header: http.Header{
			AuthorizationTokenKey: []string{token},
		},
		URL: url,
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		glog.Error("do request error: ", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		glog.Error("app info response not ok, ", resp.Status)
		return nil, errors.New("app info not found")
	}

	return resp, nil
}

func (c *Client) doHttpGetList(urlStr, token string) ([]map[string]interface{}, error) {
	resp, err := c.doHttpGetResponse(urlStr, token)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apps []map[string]interface{} // simple get. TODO: application struct
	err = json.Unmarshal(data, &apps)
	if err != nil {
		glog.Error("parse response error: ", err, string(data))
		return nil, err
	}

	return apps, nil

}
