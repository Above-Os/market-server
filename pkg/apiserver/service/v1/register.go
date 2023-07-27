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
		//Param(ws.HeaderParameter("X-Authorization", "Auth token")).
		Returns(http.StatusOK, "success to get application list", nil))

	glog.Infof("registered sub module: %s", ws.RootPath()+"/applications")

	ws.Route(ws.GET("/application_types").
		To(handler.handleTypes).
		Doc("Get application type list").
		//Param(ws.HeaderParameter("X-Authorization", "Auth token")).
		Returns(http.StatusOK, "success to get application type list", nil))

	glog.Infof("registered sub module: %s", ws.RootPath()+"/application_types")

	ws.Route(ws.GET("/application/{"+ParamAppName+"}").
		To(handler.handleApp).
		Doc("Get the application").
		Param(ws.PathParameter(ParamAppName, "the name of a application")).
		//Param(ws.HeaderParameter("X-Authorization", "Auth token")).
		Returns(http.StatusOK, "Success to get a application", nil))

	ws.Route(ws.POST("/application_updates").
		To(handler.handleUpdates).
		Doc("update applications").
		//Param(ws.HeaderParameter("X-Authorization", "Auth token")).
		Returns(http.StatusOK, "success to update applications", nil))

	glog.Infof("registered sub module: %s", ws.RootPath()+"/application_updates")

	// ws.Route(ws.GET("/{"+ParamKey+"}/prefix").
	// 	To(handler.handleGetKeyPrefix).
	// 	Doc("Multi get the data with a key prefix ").
	// 	Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
	// 	Param(ws.PathParameter(ParamKey, "key prefix of the stored data")).
	// 	Param(ws.HeaderParameter("X-Authorization", "Auth token")).
	// 	Returns(http.StatusOK, ok, nil))

	// ws.Route(ws.PUT("/{"+ParamKey+"}").
	// 	To(handler.handleSet).
	// 	Doc("Set the data of a key").
	// 	Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
	// 	Param(ws.PathParameter(ParamKey, "key of the stored data")).
	// 	Param(ws.HeaderParameter("X-Authorization", "Auth token")).
	// 	Returns(http.StatusOK, ok, nil))

	// ws.Route(ws.DELETE("/{"+ParamKey+"}").
	// 	To(handler.handleDelete).
	// 	Doc("Delete the data of a key").
	// 	Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
	// 	Param(ws.PathParameter(ParamKey, "key of the stored data")).
	// 	Param(ws.HeaderParameter("X-Authorization", "Auth token")).
	// 	Returns(http.StatusOK, ok, nil))

	c.Add(ws)
	return nil
}
