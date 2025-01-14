package appadmin

import (
	"app-store-server/internal/constants"
	"app-store-server/pkg/models"
	"app-store-server/pkg/utils"

	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
)

type Topic struct {
	ID           int                           `json:"id"`
	Name         string                        `json:"name"`
	Introduction string                        `json:"introduction"`
	Desc         string                        `json:"des"`
	IconImg      string                        `json:"iconimg"`
	DetailImg    string                        `json:"detailimg"`
	RichText     string                        `json:"richtext"`
	CreateAt     time.Time                     `json:"createat"`
	UpdateAt     time.Time                     `json:"updateat"`
	Apps         string                        `json:"apps"`
	IsDelete     bool                          `json:"isdelete"`
	AppList      []models.ApplicationInfoEntry `json:"appList"`
}

type TopResponse struct {
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
	Data    struct {
		All []Topic `json:"All"`
	} `json:"data"`
}

type CategoriesResponse struct {
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
	Data    []string    `json:"data"`
}

type ContentItem struct {
	Name         string                        `json:"name"`
	Introduction string                        `json:"introduction"`
	Des          string                        `json:"des"`
	Iconimg      string                        `json:"iconimg"`
	Detailimg    string                        `json:"detailimg"`
	Richtext     string                        `json:"richtext"`
	Apps         []models.ApplicationInfoEntry `json:"apps"`
}

type DataItem struct {
	Type        string        `json:"type"`
	TopicType   string        `json:"topicType"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Content     []interface{} `json:"content"`
}

type CategoryData struct {
	Category string     `json:"category"`
	Data     []DataItem `json:"data"`
}

type Response struct {
	Code    int            `json:"code"`
	Message interface{}    `json:"message"`
	Data    []CategoryData `json:"data"`
}

func getAppAdminServiceHost() string {
	return os.Getenv(constants.AppAdminHostEnv)
}

func getAppAdminServicePort() string {
	return os.Getenv(constants.AppAdminPortEnv)
}

func getAppGitBotHost() string {
	return os.Getenv(constants.AppGitBotHostEnv)
}

func getAppGitBotPort() string {
	return os.Getenv(constants.AppGitBotPortEnv)
}

func sendHttpRequest(method, url string, reader io.Reader) (string, error) {
	httpReq, err := http.NewRequest(method, url, reader)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return "", err
	}
	if reader != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	return utils.SendHttpRequest(httpReq)
}
