package appadmin

import (
	"app-store-server/internal/constants"
	"fmt"
)

func GetAppHistory(appName string) (string, error) {
	url := fmt.Sprintf(constants.AppGitBotAppVersionHistoryURLTempl, getAppGitBotHost(), getAppGitBotPort(), appName)
	bodyStr, err := sendHttpRequest("GET", url, nil)
	return bodyStr, err
}
