package appadmin

import (
	"app-store-server/internal/constants"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
)

const (
	RecommendKey = "RECOMMEND"
	TopicKey     = "TOPIC"
	CategoryKey  = "CATEGORY"
)

var (
	cache sync.Map
)

func init() {
	go getRecommendsDetailLoop()
	go getTopicsDetailLoop()
	go getCategoriesDetailLoop()
}

func getRecommendsDetailLoop() {
	callLoop(getTopicsDetailFromAdmin)
}

func getTopicsDetailLoop() {
	callLoop(getRecommendsDetailFromAdmin)
}

func getCategoriesDetailLoop() {
	callLoop(getCategoriesDetailFromAdmin)
}

func callLoop(f func() error) {
	err := f()
	if err != nil {
		glog.Warningf("err:%s", err.Error())
	}
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := f()
			if err != nil {
				glog.Warningf("err:%s", err.Error())
			}
		}
	}
}

func getRecommendsDetailStrFromAdmin() (string, error) {
	url := fmt.Sprintf(constants.AppAdminServiceRecommendsDetailURLTempl, getAppAdminServiceHost(), getAppAdminServicePort())
	bodyStr, err := sendHttpRequest("GET", url, nil)
	return bodyStr, err
}

func getRecommendsDetailFromAdmin() error {
	bodyStr, err := getRecommendsDetailStrFromAdmin()
	if err != nil {
		return err
	}

	cache.Store(RecommendKey, bodyStr)
	return nil
}

func getTopicsDetailFromAdmin() error {
	url := fmt.Sprintf(constants.AppAdminServiceTopicsDetailURLTempl, getAppAdminServiceHost(), getAppAdminServicePort())
	bodyStr, err := sendHttpRequest("GET", url, nil)
	if err != nil {
		return err
	}

	cache.Store(TopicKey, bodyStr)
	return nil
}

func getCategoriesDetailFromAdmin() error {
	url := fmt.Sprintf(constants.AppAdminServiceCategoriesURLTempl, getAppAdminServiceHost(), getAppAdminServicePort())
	bodyStr, err := sendHttpRequest("GET", url, nil)
	if err != nil {
		return err
	}

	cache.Store(CategoryKey, bodyStr)
	return nil
}

func GetCategoriesDetail() interface{} {
	value, _ := cache.Load(CategoryKey)
	if value == nil {
		_ = getCategoriesDetailFromAdmin()
		value, _ = cache.Load(CategoryKey)
	}
	return value
}

func GetRecommendsDetail() interface{} {
	value, _ := cache.Load(RecommendKey)
	if value == nil {
		_ = getRecommendsDetailFromAdmin()
		value, _ = cache.Load(RecommendKey)
	}
	return value
}

func GetTopicsDetail() interface{} {
	value, _ := cache.Load(TopicKey)
	if value == nil {
		_ = getTopicsDetailFromAdmin()
		value, _ = cache.Load(TopicKey)
	}

	return value
}
