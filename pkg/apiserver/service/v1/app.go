package v1

import (
	"app-store-server/internal/app"
	"app-store-server/internal/constants"
	"app-store-server/internal/es"
	"app-store-server/internal/mongo"
	"app-store-server/pkg/models"
	"app-store-server/pkg/utils"
	"path"
)

func getChartPath(appName string) string {
	filePathName := path.Join(constants.AppGitZipLocalDir, appName)
	if exist, _ := utils.PathExists(filePathName); exist {
		return filePathName
	}

	//not a chart name, search chart name
	info, err := getInfoByName(appName)
	if err == nil && info != nil && info.ChartName != "" {
		filePathName = path.Join(constants.AppGitZipLocalDir, info.ChartName)
		if exist, _ := utils.PathExists(filePathName); exist {
			return filePathName
		}
	}

	return ""
}

func getInfoByName(appName string) (*models.ApplicationInfo, error) {
	info, err := es.SearchByNameAccurate(appName)
	if err == nil && info != nil {
		return info, nil
	}

	info, err = mongo.GetAppInfoByName(appName)
	if err == nil && info != nil {
		return info, nil
	}

	info, err = app.ReadAppInfo(appName)

	return info, err
}
