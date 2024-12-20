package v1

import (
	"app-store-server/internal/appadmin"
	"app-store-server/pkg/api"
	"errors"
	"os"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/glog"
)

func (h *Handler) pagesDetail(req *restful.Request, resp *restful.Response) {
	version := req.QueryParameter("version")
	if version == "" {
		version = "1.10.9-0"
	}

	if version == "latest" {
		version = os.Getenv("LATEST_VERSION")
	}

	detail, err := appadmin.GetPagesDetailFromAdmin(version)
	if err != nil {
		glog.Warningf("err:%s", err)
	}
	//todo deal with error
	if detail == nil {
		api.HandleError(resp, req, errors.New("get empty detail"))
		return
	}

	_, err = resp.Write([]byte(detail.(string)))
	if err != nil {
		glog.Warningf("err:%s", err)
	}
}

func (h *Handler) handleVersionHistory(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	if appName == "" {
		api.HandleError(resp, req, errors.New("param invalid"))
		return
	}

	respBody, err := appadmin.GetAppHistory(appName)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	_, err = resp.Write([]byte(respBody))
	if err != nil {
		glog.Warningf("err:%s", err)
	}
}
