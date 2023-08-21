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

	topicsExample         = `[{"name":"test","intro":"test","desc":"test","iconSrc":"test","detailsImgSrc":"test","richText":"test"}]`
	cateRecommendsExample = `[{"category":"all","data":[{"name":"test1","desc":"test1"},{"name":"test2","desc":"test2"}]},{"category":"cat2","data":[{"name":"test11","desc":"test11"},{"name":"test22","desc":"test22"}]}]`
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
		Reads([]models.CmsTopic{}).
		Param(ws.BodyParameter("topics", "topic list").DataFormat("json").DataType("json").PossibleValues([]string{topicsExample}).Required(true)).
		Returns(http.StatusOK, "success to set the topic list", &models.ResponseBase{}))

	ws.Route(ws.GET("/recommends").
		To(handler.getCateRecommends).
		Doc("set recommends list").
		Returns(http.StatusOK, "success to set recommend list", &models.CmsCategoryRecommendListResponse{}))

	ws.Route(ws.PUT("/recommends").
		To(handler.setCateRecommends).
		Doc("get recommend list").
		Reads([]models.CmsCategoryRecommend{}).
		Param(ws.BodyParameter("recommends", "recommend list").DataFormat("json").DataType("json").PossibleValues([]string{cateRecommendsExample}).Required(true)).
		Returns(http.StatusOK, "success to get recommend list", &models.ResponseBase{}))

	c.Add(ws)
	return nil
}
