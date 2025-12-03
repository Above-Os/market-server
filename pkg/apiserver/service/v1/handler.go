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
	"app-store-server/internal/app"
	"app-store-server/internal/es"
	"app-store-server/internal/gitapp"
	"app-store-server/internal/mongo"
	"app-store-server/pkg/api"
	"app-store-server/pkg/models"
	"app-store-server/pkg/utils"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/emicklei/go-restful/v3"
)

type Handler struct {
}

func newHandler() *Handler {
	return &Handler{}
}

func (h *Handler) handleList(req *restful.Request, resp *restful.Response) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	category := req.QueryParameter("category")
	ty := req.QueryParameter("type")
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

	from, sizeN := utils.VerifyFromAndSize(page, size)

	appList, count, err := mongo.GetAppLists(int64(from), int64(sizeN), category, ty)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	appEntryList, err := filterVersionForApps(appList, version)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, models.NewListResultWithCount(appEntryList, count)))
}

func (h *Handler) handleTypes(req *restful.Request, resp *restful.Response) {

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

	appList, _, err := mongo.GetAppLists(0, 10000, "", "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	appEntryList, err := filterVersionForApps(appList, version)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	var categorieList []string
	for _, app := range appEntryList {
		categorieList = append(categorieList, app.Categories...)
	}

	categorieList = utils.RemoveDuplicates(categorieList)

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, models.NewListResultWithCount(categorieList, int64(len(categorieList)))))
}

func (h *Handler) handleApp(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
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

	fileName := getChartPath(appName, version)

	if fileName == "" {
		api.HandleError(resp, req, fmt.Errorf("failed to get chart"))
		return
	}

	fileBytes, err := os.ReadFile(fileName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.ResponseWriter.Write(fileBytes)
}

func (h *Handler) handleAppInfo(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
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

	if appName == "" {
		api.HandleError(resp, req, errors.New("empty app name"))
		return
	}

	info, err := getInfoByName(appName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	appEntry, err := filterVersionForApp(info, version)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, appEntry))
}

func (h *Handler) handleReadme(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	if appName == "" {
		api.HandleError(resp, req, errors.New("empty app name"))
		return
	}

	content, err := gitapp.ReadMe(appName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.Write(content)
}

func (h *Handler) handleUpdate(req *restful.Request, resp *restful.Response) {
	err := app.GitPullAndUpdate(true)

	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, nil))
}

func (h *Handler) handleTop(req *restful.Request, resp *restful.Response) {
	//todo local cache results
	category := req.QueryParameter("category")
	ty := req.QueryParameter("type")
	size := req.QueryParameter("size")
	excludedLabels := req.QueryParameter("excludedLabels")
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

	excludedLabelsSlice := strings.Split(excludedLabels, ",")
	sizeN := utils.VerifyTopSize(size)
	infos, err := mongo.GetTopApplicationInfos(category, ty, excludedLabelsSlice, sizeN)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	var appPointers []*models.ApplicationInfoFullData
	for i := range infos {
		appPointers = append(appPointers, &infos[i])
	}

	appEntryList, err := filterVersionForApps(appPointers, version)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, models.NewListResult(appEntryList)))
}

func (h *Handler) handleSearch(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
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

	from, sizeN := utils.VerifyFromAndSize(page, size)
	appList, count, err := es.SearchByNameWildcard(from, sizeN, appName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	appEntryList, err := pickVersionForApps(appList, version)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, models.NewListResultWithCount(appEntryList, count)))
}

func (h *Handler) handleExist(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	exist := gitapp.AppDirExist(appName)
	res := &models.ExistRes{
		Exist: exist,
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, res))
}

func (h *Handler) handleCount(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	err := mongo.SetAppInstallCount(appName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, nil))
}

func (h *Handler) handleInfos(req *restful.Request, resp *restful.Response) {
	version := req.PathParameter("version")
	if version == "" {
		version = "1.10.9-0"
	}

	if version == "undefined" {
		version = "1.10.9-0"
	}

	if version == "latest" {
		version = os.Getenv("LATEST_VERSION")
	}

	var names []string
	err := req.ReadEntity(&names)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	mapInfo, err := mongo.GetAppInfos(names)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	appEntryList, err := pickVersionForAppsWithMap(mapInfo, version)

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, appEntryList))
}
