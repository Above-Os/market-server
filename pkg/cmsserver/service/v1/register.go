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
	"app-store-server/pkg/models"
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/glog"
)

const (
	APIRootPath = "/app-store-admin-server"
	Version     = "v1"
	ParamName   = "name"

	topicsExample         = `[{"name":"test","intro":"test","desc":"test","iconSrc":"test","detailsImgSrc":"test","richText":"test"}]`
	recommendsExample     = `[{"name":"test1","desc":"test1"},{"name":"test22","desc":"test2"}]`
	cateRecommendsExample = `[{"category":"cat1","data":[{"name":"test1","desc":"test1"},{"name":"test2","desc":"test2"}]},{"category":"cat2","data":[{"name":"test11","desc":"test11"},{"name":"test22","desc":"test22"}]}]`
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

	ws.Route(ws.GET("/topics").
		To(handler.getTopics).
		Doc("get topic list").
		Returns(http.StatusOK, "success to get the topic list", &models.CmsTopicListResponse{}))

	glog.Infof("registered sub module: %s", ws.RootPath()+"/topics")

	ws.Route(ws.PUT("/topics").
		To(handler.setTopics).
		Doc("set topic list").
		Param(ws.BodyParameter("topics", "topic list").DataFormat("json").DataType("json").PossibleValues([]string{topicsExample}).Required(true)).
		Returns(http.StatusOK, "success to set the topic list", &models.ResponseBase{}))

	//ws.Route(ws.POST("/topic/{"+ParamName+"}").
	//	To(handler.addOneTopic).
	//	Doc("add/update one topic").
	//	Returns(http.StatusOK, "success to add/update one topic", &models.ResponseBase{}))
	//
	//ws.Route(ws.GET("/topic/{"+ParamName+"}").
	//	To(handler.getOneTopic).
	//	Doc("get one topic").
	//	Returns(http.StatusOK, "success to get one topic", &models.CmsTopicResponse{}))
	//
	//ws.Route(ws.DELETE("/topic/{"+ParamName+"}").
	//	To(handler.delOneTopic).
	//	Doc("delete one topic").
	//	Returns(http.StatusOK, "success to delete one topic", &models.ResponseBase{}))

	ws.Route(ws.GET("/recommends").
		To(handler.getRecommends).
		Doc("set recommends list").
		Returns(http.StatusOK, "success to set recommend list", &models.CmsRecommendListResponse{}))

	ws.Route(ws.PUT("/recommends").
		To(handler.setRecommends).
		Doc("get recommend list").
		Param(ws.BodyParameter("recommends", "recommend list").DataFormat("json").DataType("json").PossibleValues([]string{recommendsExample}).Required(true)).
		Returns(http.StatusOK, "success to get recommend list", &models.ResponseBase{}))

	ws.Route(ws.GET("/category-recommends").
		To(handler.getCateRecommends).
		Doc("set category recommend list").
		Returns(http.StatusOK, "success to set category recommend list", &models.CmsCategoryRecommendListResponse{}))

	ws.Route(ws.PUT("/category-recommends").
		To(handler.setCateRecommends).
		Doc("get category recommend list").
		Param(ws.BodyParameter("category recommends", "category recommend list").DataFormat("json").DataType("json").PossibleValues([]string{cateRecommendsExample}).Required(true)).
		Returns(http.StatusOK, "success to get category recommend list", &models.ResponseBase{}))

	c.Add(ws)
	return nil
}
