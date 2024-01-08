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
	"os"

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
	ty := req.QueryParameter("type")
	if ty == "" {
		ty = "app"
	}

	glog.Infof("page:%s, size:%s, category:%s", page, size, category)

	from, sizeN := utils.VerifyFromAndSize(page, size)

	appList, count, err := mongo.GetAppLists(int64(from), int64(sizeN), category, ty)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, models.NewListResultWithCount(appList, count)))
}

func (h *Handler) handleTypes(req *restful.Request, resp *restful.Response) {
	types, err := mongo.GetAppTypesFromDb()
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, models.NewListResult(types)))
}

func (h *Handler) handleApp(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	fileName := getChartPath(appName)
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

	info, err := getInfoByName(appName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, info))
}

func (h *Handler) handleUpdates(req *restful.Request, resp *restful.Response) {
	err := app.GitPullAndUpdate()

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
	if ty == "" {
		ty = "app"
	}
	size := req.QueryParameter("size")
	sizeN := utils.VerifyTopSize(size)
	infos, err := mongo.GetTopApplicationInfos(category, ty, sizeN)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(models.NewResponse(api.OK, api.Success, models.NewListResult(infos)))
}

func (h *Handler) handleSearch(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	from, sizeN := utils.VerifyFromAndSize(page, size)
	appList, count, err := es.SearchByNameWildcard(from, sizeN, appName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, models.NewListResultWithCount(appList, count)))
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

	resp.WriteEntity(models.NewResponse(api.OK, api.Success, mapInfo))
}
