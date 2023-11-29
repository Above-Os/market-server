package utils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
)

var (
	client *http.Client
	once   sync.Once
)

func init() {
	once.Do(func() {
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   time.Duration(3) * time.Second,
					KeepAlive: time.Duration(60) * time.Second,
				}).DialContext,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     time.Duration(60) * time.Second,
			},
			Timeout: time.Duration(3) * time.Second,
		}
	})
}

func GetHttpClient() *http.Client {
	return client
}

func SendHttpRequest(req *http.Request) (string, error) {
	resp, err := GetHttpClient().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	//glog.Infof("%s, res:%s\n", req.URL, string(body))

	if resp.StatusCode != http.StatusOK {
		glog.Warningf("res:%s, resp.StatusCode:%d", string(body), resp.StatusCode)
		return string(body), fmt.Errorf("http status not 200 %d msg:%s", resp.StatusCode, string(body))
	}

	return string(body), nil
}
