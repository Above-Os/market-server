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
	"fmt"
	"io/ioutil"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/glog"
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

	glog.Infof("page:%s, size:%s, category:%s", page, size, category)

	from, sizeN := utils.VerifyFromAndSize(page, size)

	appList, count, err := mongo.GetAppLists(int64(from), int64(sizeN), category)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.NewResponse(api.OK, api.Success, api.NewListResultWithCount(appList, count)))
}

func (h *Handler) handleTypes(req *restful.Request, resp *restful.Response) {
	types, err := mongo.GetAppTypesFromDb()
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.NewResponse(api.OK, api.Success, api.NewListResult(types)))
}

func (h *Handler) handleApp(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	fileName := getChartPath(appName)
	if fileName == "" {
		api.HandleError(resp, req, fmt.Errorf("failed to get chart"))
		return
	}

	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.ResponseWriter.Write(fileBytes)
}

func (h *Handler) handleAppInfo(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)

	info, err := getInfoByName(appName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.NewResponse(api.OK, api.Success, info))
}

func (h *Handler) handleUpdates(req *restful.Request, resp *restful.Response) {
	err := app.GitPullAndUpdate()

	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.NewResponse(api.OK, api.Success, nil))
}

func (h *Handler) handleTop(req *restful.Request, resp *restful.Response) {
	//todo local cache results
	var results []models.TopResultItem
	categories := es.GetCategories()
	glog.Infof("categories:%+v", categories)
	for _, category := range categories {
		var result models.TopResultItem
		result.Category = category
		//default every category 3 apps
		infos, err := es.SearchByCategory(0, 3, category)
		if err != nil {
			continue
		}
		for _, info := range infos {
			result.Apps = append(result.Apps, info)
		}
		results = append(results, result)
	}

	resp.WriteEntity(api.NewResponse(api.OK, api.Success, api.NewListResult(results)))
}

func (h *Handler) handleSearch(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	from, sizeN := utils.VerifyFromAndSize(page, size)
	appList, count, err := es.SearchByNameFuzzy(from, sizeN, appName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.NewResponse(api.OK, api.Success, api.NewListResultWithCount(appList, count)))
}

func (h *Handler) handleExist(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	exist := gitapp.AppDirExist(appName)
	res := &models.ExistRes{
		Exist: exist,
	}

	resp.WriteEntity(api.NewResponse(api.OK, api.Success, res))
}
