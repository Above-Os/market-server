package v2

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/glog"
)

const (
	APIRootPath  = "/app-store-server"
	Version      = "v2"
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

	// Get appstore information
	ws.Route(ws.GET("/appstore/info").
		To(handler.handleAppStoreInfo).
		Doc("Get appstore information").
		Param(ws.QueryParameter("version", "version of the system")).
		Param(ws.QueryParameter("page", "page number for pagination")).
		Param(ws.QueryParameter("size", "page size for pagination")).
		Returns(http.StatusOK, "success to get appstore information", AppStoreInfoResponse{}))

	glog.Infof("registered sub module: %s", ws.RootPath()+"/appstore/info")

	// Download application chart
	ws.Route(ws.GET("/applications/{"+ParamAppName+"}/chart").
		To(handler.handleChartDownload).
		Doc("Download application chart package").
		Param(ws.PathParameter(ParamAppName, "the name of the application")).
		Param(ws.QueryParameter("version", "version of the system")).
		Returns(http.StatusOK, "success to download application chart", nil))

	glog.Infof("registered sub module: %s", ws.RootPath()+"/applications/{name}/chart")

	// Get appstore information hash
	ws.Route(ws.GET("/appstore/hash").
		To(handler.handleAppStoreHash).
		Doc("Get appstore information hash").
		Param(ws.QueryParameter("version", "version of the system")).
		Param(ws.QueryParameter("page", "page number for pagination")).
		Param(ws.QueryParameter("size", "page size for pagination")).
		Returns(http.StatusOK, "success to get appstore hash", AppStoreHashResponse{}))

	glog.Infof("registered sub module: %s", ws.RootPath()+"/appstore/hash")

	c.Add(ws)
	return nil
} 