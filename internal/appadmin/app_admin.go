package appadmin

import (
	"app-store-server/internal/constants"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
)

const (
	PageKey = "PAGE"
)

var (
	cache sync.Map
)

func init() {
	go getPagesDetailLoop()
}

func getPagesDetailLoop() {
	callLoop(getPagesDetailFromAdmin)
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

func getPagesDetailFromAdmin() error {
	url := fmt.Sprintf(constants.AppAdminServicePagesDetailURLTempl, getAppAdminServiceHost(), getAppAdminServicePort())
	bodyStr, err := sendHttpRequest("GET", url, nil)
	if err != nil {
		return err
	}

	cache.Store(PageKey, bodyStr)
	return nil
}

func GetPagesDetail() interface{} {
	value, _ := cache.Load(PageKey)
	if value == nil {
		_ = getPagesDetailFromAdmin()
		value, _ = cache.Load(PageKey)
	}
	return value
}
