package v1

import (
	"app-store-server/internal/appadmin"
	"app-store-server/pkg/api"
	"errors"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/glog"
)

func (h *Handler) recommendsDetail(req *restful.Request, resp *restful.Response) {
	detail := appadmin.GetRecommendsDetail()
	//todo deal with error
	if detail == nil {
		api.HandleError(resp, req, errors.New("get empty detail"))
		return
	}

	_, err := resp.Write([]byte(detail.(string)))
	if err != nil {
		glog.Warningf("err:%s", err)
	}
}

func (h *Handler) topicsDetail(req *restful.Request, resp *restful.Response) {
	detail := appadmin.GetTopicsDetail()
	//todo deal with error
	if detail == nil {
		api.HandleError(resp, req, errors.New("get empty detail"))
		return
	}

	_, err := resp.Write([]byte(detail.(string)))
	if err != nil {
		glog.Warningf("err:%s", err)
	}
}

func (h *Handler) categories(req *restful.Request, resp *restful.Response) {
	detail := appadmin.GetCategoriesDetail()
	//todo deal with error
	if detail == nil {
		api.HandleError(resp, req, errors.New("get empty detail"))
		return
	}

	_, err := resp.Write([]byte(detail.(string)))
	if err != nil {
		glog.Warningf("err:%s", err)
	}
}
