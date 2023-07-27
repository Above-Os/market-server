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
	"app-store-server/internal/constants"
	mongo2 "app-store-server/internal/mongo"
	"app-store-server/pkg/api"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/glog"
)

const (
	ChartsPath = "./charts"
)

type Handler struct {
	appServiceClient *Client
}

func newHandler() *Handler {
	return &Handler{
		appServiceClient: newAppServiceClient(),
	}
}

func (h *Handler) handleList(req *restful.Request, resp *restful.Response) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	category := req.QueryParameter("category")

	glog.Infof("page:%s, size:%s, category:%s", page, size, category)
	pageN, _ := strconv.Atoi(page)
	if pageN < 1 {
		pageN = 1
	}

	sizeN, _ := strconv.Atoi(size)
	if sizeN < 1 {
		sizeN = 5
	}
	appList, count, err := mongo2.GetAppListsFromDb(int64(pageN), int64(sizeN), category)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.NewListResultWithCount(appList, count))
}

func (h *Handler) handleTypes(req *restful.Request, resp *restful.Response) {
	types, err := mongo2.GetAppTypesFromDb()
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.NewListResult(types))
}

func (h *Handler) handleApp(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	fileName := fmt.Sprintf("%s/%s", constants.AppGitZipLocalDir, appName)
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		api.HandleError(resp, req, err)
	}
	resp.ResponseWriter.Write(fileBytes)
}

func (h *Handler) handleUpdates(req *restful.Request, resp *restful.Response) {
	err := app.GitPullAndUpdate()

	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.Response{
		Code:    200,
		Message: "ok",
	})
}
