package v2

import (
	"app-store-server/internal/constants"
	"app-store-server/internal/mongo"
	"app-store-server/pkg/api"
	"app-store-server/pkg/models"
	"app-store-server/pkg/utils"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/glog"
)

type Handler struct {
}

func newHandler() *Handler {
	return &Handler{}
}

// calculateHash calculates MD5 hash for apps and tops data
func calculateHash(apps []models.ApplicationInfoFullData, tops []AppStoreTopItem) string {
	data := struct {
		Apps []models.ApplicationInfoFullData `json:"apps"`
		Tops []AppStoreTopItem                `json:"tops"`
	}{
		Apps: apps,
		Tops: tops,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		glog.Errorf("Failed to marshal data for hash calculation: %v", err)
		return ""
	}

	hash := md5.Sum(jsonData)
	return fmt.Sprintf("%x", hash)
}

// getTopsData gets top applications data and converts to AppStoreTopItem format
func getTopsData() ([]AppStoreTopItem, error) {
	// Get top applications from database, similar to handleTop in v1
	// Use default parameters: category="", type="", excludedLabels=empty, size=10000
	excludedLabels := []string{}
	sizeN := 10000 // Default top size

	infos, err := mongo.GetTopApplicationInfos("", "", excludedLabels, sizeN)
	if err != nil {
		glog.Errorf("Failed to get top application infos: %v", err)
		return nil, err
	}

	// Convert to AppStoreTopItem format
	var tops []AppStoreTopItem
	for i, info := range infos {
		tops = append(tops, AppStoreTopItem{
			AppID: info.Name, // Use application name as AppID
			Rank:  i + 1,     // Rank starts from 1
		})
	}

	glog.Infof("Retrieved %d top applications", len(tops))
	return tops, nil
}

// filterVersionForApps filters apps based on version constraints similar to v1
func filterVersionForApps(apps []*models.ApplicationInfoFullData, version string) ([]models.ApplicationInfoEntry, error) {
	var result []models.ApplicationInfoEntry

	// Parse the passed version string
	v, err := semver.NewVersion(version)
	if err != nil {
		glog.Infof("error version:%s", version)
		return nil, err
	}

	for _, app := range apps {
		var maxEntry *models.ApplicationInfoEntry

		// Traversing the history
		for _, entry := range app.History {
			// Traversing dependencies
			for _, dep := range entry.Options.Dependencies {
				if dep.Name == "olares" && dep.Type == "system" {
					// Parsing version strings
					constraint, err := semver.NewConstraint(dep.Version)
					if err != nil {
						glog.Infof("error version:%s", dep.Version)
						return nil, err
					}

					// Check if the version meets the constraints
					if constraint.Check(v) {
						appv, err := semver.NewVersion(entry.Version)
						if err != nil {
							glog.Infof("error version:%s, error app:%s", entry.Version, entry.Name)
							continue
						}

						// If the conditions are met, check whether it is the largest version
						if maxEntry == nil || appv.GreaterThan(semver.MustParse(maxEntry.Version)) {
							entryCopy := entry
							maxEntry = &entryCopy
						}
					}
				}
			}
		}

		// If the largest version matching the criteria is found, add it to the result
		if maxEntry != nil {
			result = append(result, *maxEntry)
		}
	}

	return result, nil
}

// getTopsDataWithVersion gets top applications data with version filtering
func getTopsDataWithVersion(version string) ([]AppStoreTopItem, error) {
	// Get top applications from database, similar to handleTop in v1
	excludedLabels := []string{}
	sizeN := 10000 // Default top size

	infos, err := mongo.GetTopApplicationInfos("", "", excludedLabels, sizeN)
	if err != nil {
		glog.Errorf("Failed to get top application infos: %v", err)
		return nil, err
	}

	// Convert to app pointers for version filtering
	var appPointers []*models.ApplicationInfoFullData
	for i := range infos {
		appPointers = append(appPointers, &infos[i])
	}

	// Filter by version
	filteredEntries, err := filterVersionForApps(appPointers, version)
	if err != nil {
		glog.Errorf("Failed to filter top apps by version: %v", err)
		return nil, err
	}

	// Convert to AppStoreTopItem format
	var tops []AppStoreTopItem
	for i, entry := range filteredEntries {
		tops = append(tops, AppStoreTopItem{
			AppID: entry.Name, // Use application name as AppID
			Rank:  i + 1,      // Rank starts from 1
		})
	}

	glog.Infof("Retrieved %d filtered top applications", len(tops))
	return tops, nil
}

// getAppStoreData gets apps, tops, and stats data for appstore with version filtering
func getAppStoreData(page, size, version string) (*AppStoreInfo, error) {
	// Verify and convert page parameters
	from, sizeN := utils.VerifyFromAndSize(page, size)

	// Get apps data from database with pagination, use empty category and type
	appList, totalCount, err := mongo.GetAppLists(int64(from), int64(sizeN), "", "")
	if err != nil {
		glog.Errorf("Failed to get app lists: %v", err)
		return nil, err
	}

	glog.Infof("Retrieved %d apps from database", len(appList))

	// Filter apps based on version constraints
	appEntryList, err := filterVersionForApps(appList, version)
	if err != nil {
		glog.Errorf("Failed to filter apps by version: %v", err)
		return nil, err
	}

	glog.Infof("Filtered %d apps after version filtering", len(appEntryList))

	// Convert []ApplicationInfoEntry to []ApplicationInfoFullData for hash calculation
	var apps []models.ApplicationInfoFullData
	for _, entry := range appEntryList {
		// Create a minimal ApplicationInfoFullData with the filtered entry
		fullData := models.ApplicationInfoFullData{
			Name:      entry.Name,
			AppLabels: entry.AppLabels,
			History:   make(map[string]models.ApplicationInfoEntry),
		}
		fullData.History[entry.Version] = entry
		apps = append(apps, fullData)
	}

	// Get tops data from database with version filtering
	tops, err := getTopsDataWithVersion(version)
	if err != nil {
		glog.Errorf("Failed to get tops data: %v", err)
		return nil, err
	}

	// Calculate hash from apps and tops data
	hash := calculateHash(apps, tops)

	// Create stats with filtered count
	stats := AppStoreStats{
		TotalApps:  totalCount,               // Original total count from database
		TotalItems: int64(len(appEntryList)), // Filtered items count
		Hash:       hash,
	}

	// Create AppStoreInfo with filtered entry list
	appStoreInfo := &AppStoreInfo{
		Apps:  appEntryList, // Use filtered entry list instead of full data
		Tops:  tops,
		Stats: stats,
	}

	return appStoreInfo, nil
}

// handleAppStoreInfo handles the request to get appstore information
func (h *Handler) handleAppStoreInfo(req *restful.Request, resp *restful.Response) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	version := req.QueryParameter("version")

	if version == "" {
		version = "1.10.9-0"
	}

	if version == "undefined" {
		version = "1.10.9-0"
	}

	if version == "latest" {
		version = os.Getenv("LATEST_VERSION")
	}

	glog.Infof("handleAppStoreInfo - page:%s, size:%s, version:%s", page, size, version)

	// Get appstore data with version filtering
	appStoreInfo, err := getAppStoreData(page, size, version)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	response := &AppStoreInfoResponse{
		AppStore: appStoreInfo,
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, response))
}

// handleChartDownload handles the request to download application chart
func (h *Handler) handleChartDownload(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	version := req.QueryParameter("version")
	fileName := req.QueryParameter("fileName") // Get fileName from query parameter

	if appName == "" {
		api.HandleError(resp, req, errors.New("app name is required"))
		return
	}

	if version == "" {
		version = "1.10.9-0"
	}

	if version == "undefined" {
		version = "1.10.9-0"
	}

	if version == "latest" {
		version = os.Getenv("LATEST_VERSION")
	}

	glog.Infof("handleChartDownload for app: %s, version: %s, fileName: %s", appName, version, fileName)

	// Determine file path based on fileName parameter
	var filePath string
	if fileName != "" {
		// Use the new getChartPath method with fileName only
		filePath = getChartPathByFileName(fileName)
	} else {
		// Fallback to original logic if fileName not provided
		// This would require importing v1 package or implementing similar logic
		api.HandleError(resp, req, errors.New("fileName parameter is required"))
		return
	}

	glog.Infof("Chart file path: %s", filePath)

	if filePath == "" {
		api.HandleError(resp, req, fmt.Errorf("failed to get chart path"))
		return
	}

	// Read and return the file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		glog.Errorf("Failed to read chart file: %v", err)
		api.HandleError(resp, req, err)
		return
	}

	resp.ResponseWriter.Write(fileBytes)
}

// getChartPathByFileName gets the chart file path by fileName only
func getChartPathByFileName(fileName string) string {
	return path.Join(constants.AppGitZipLocalDir, fileName)
}

// handleAppStoreHash handles the request to get appstore information hash
func (h *Handler) handleAppStoreHash(req *restful.Request, resp *restful.Response) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	version := req.QueryParameter("version")

	if version == "" {
		version = "1.10.9-0"
	}

	if version == "undefined" {
		version = "1.10.9-0"
	}

	if version == "latest" {
		version = os.Getenv("LATEST_VERSION")
	}

	glog.Infof("handleAppStoreHash - page:%s, size:%s, version:%s", page, size, version)

	// Get appstore data with version filtering (same as handleAppStoreInfo)
	appStoreInfo, err := getAppStoreData(page, size, version)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	hashResponse := &AppStoreHashResponse{
		Hash:      appStoreInfo.Stats.Hash,
		UpdatedAt: time.Now(),
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, hashResponse))
}
