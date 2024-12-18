package appadmin

import (
	"app-store-server/internal/constants"
	"fmt"
	"sync"
)

const (
	pageKey = "PAGE"
)

var (
	cache sync.Map
)

func init() {
	//todo extract func, not in this init
	// go getPagesDetailLoop()
}

// func getPagesDetailLoop() {
// 	callLoop(getPagesDetailFromAdmin)
// }

// func callLoop(f func() error) {
// 	err := f()
// 	if err != nil {
// 		glog.Warningf("err:%s", err.Error())
// 	}
// 	ticker := time.NewTicker(time.Minute)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ticker.C:
// 			err := f()
// 			if err != nil {
// 				glog.Warningf("err:%s", err.Error())
// 			}
// 		}
// 	}
// }

func GetPagesDetailFromAdmin(version string) (interface{}, error) {
	// url := fmt.Sprintf(constants.AppAdminServicePagesDetailURLTempl, getAppAdminServiceHost(), getAppAdminServicePort())
	url := fmt.Sprintf(constants.AppAdminServicePagesDetailURLTemplV2, getAppAdminServiceHost(), version)

	bodyStr, err := sendHttpRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// cache.Store(pageKey, bodyStr)
	return bodyStr, nil
}

// func GetPagesDetail() interface{} {
// 	value, _ := cache.Load(pageKey)
// 	if value == nil {
// 		_ = getPagesDetailFromAdmin()
// 		value, _ = cache.Load(pageKey)
// 	}
// 	return value
// }
