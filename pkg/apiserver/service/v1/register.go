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
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/glog"
)

const (
	APIRootPath  = "/app-store-server"
	Version      = "v1"
	ParamAppName = "name"
)

func newWebService() *restful.WebService {
	webservice := restful.WebService{}

	webservice.Path(fmt.Sprintf("%s/%s", APIRootPath, Version)).
		Produces(restful.MIME_JSON)

	return &webservice
}

func AddToContainer(c *restful.Container) error {
	ws := newWebService()
	handler := newHandler()

	ws.Route(ws.GET("/applications").
		To(handler.handleList).
		Doc("Get application list").
		Param(ws.QueryParameter("page", "page")).
		Param(ws.QueryParameter("size", "size")).
		Param(ws.QueryParameter("category", "category")).
		Returns(http.StatusOK, "success to get application list", nil))

	glog.Infof("registered sub module: %s", ws.RootPath()+"/applications")

	ws.Route(ws.GET("/applications/types").
		To(handler.handleTypes).
		Doc("Get application type list").
		Returns(http.StatusOK, "success to get application type list", nil))

	glog.Infof("registered sub module: %s", ws.RootPath()+"/application_types")

	ws.Route(ws.GET("/applications/{"+ParamAppName+"}").
		To(handler.handleApp).
		Doc("download the application chart").
		Param(ws.PathParameter(ParamAppName, "the (chart)name of the application")).
		Returns(http.StatusOK, "Success to get the application chart", nil))

	ws.Route(ws.GET("/applications/info/{"+ParamAppName+"}").
		To(handler.handleAppInfo).
		Doc("get the application info").
		Param(ws.PathParameter(ParamAppName, "the name of the application")).
		Returns(http.StatusOK, "Success to get the application info", nil))

	ws.Route(ws.POST("/applications/update").
		To(handler.handleUpdates).
		Doc("update applications").
		Returns(http.StatusOK, "success to update applications", nil))

	glog.Infof("registered sub module: %s", ws.RootPath()+"/application_updates")

	ws.Route(ws.GET("/applications/top").
		To(handler.handleTop).
		Doc("Get top application list").
		Returns(http.StatusOK, "success to get top application list", nil))

	ws.Route(ws.GET("/applications/search/{"+ParamAppName+"}").
		To(handler.handleSearch).
		Param(ws.PathParameter(ParamAppName, "the name of the application")).
		Param(ws.QueryParameter("page", "page")).
		Param(ws.QueryParameter("size", "size")).
		Doc("search application list by name").
		Returns(http.StatusOK, "success to search application list by name", nil))

	ws.Route(ws.GET("/applications/exist/{"+ParamAppName+"}").
		To(handler.handleExist).
		Param(ws.PathParameter(ParamAppName, "the name of the application")).
		Doc("does the application exist by name").
		Returns(http.StatusOK, "success to judge the application exist by name", nil))

	c.Add(ws)
	return nil
}
