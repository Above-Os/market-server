package v1

import (
	"app-store-server/internal/mongo"
	"app-store-server/pkg/api"
	"app-store-server/pkg/models"
	"github.com/emicklei/go-restful/v3"
)

type Handler struct {
}

func newHandler() *Handler {
	return &Handler{}
}

func (h *Handler) getTopics(req *restful.Request, resp *restful.Response) {
	lists, err := mongo.GetTopics()
	if err != nil {
		api.HandleError(resp, req, err)
	}

	res := &models.CmsTopicListResponse{
		ResponseBase: models.ResponseBase{
			Code: 200,
			Msg:  api.Success,
		},
		Data: lists,
	}

	resp.WriteEntity(res)
}

func (h *Handler) setTopics(req *restful.Request, resp *restful.Response) {
	var lists []models.CmsTopic
	err := req.ReadEntity(&lists)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	err = mongo.SetTopics(lists)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(&models.ResponseBase{Code: api.OK, Msg: api.Success})
}
