package v1

import (
	"app-store-server/internal/mongo"
	"app-store-server/pkg/api"
	"app-store-server/pkg/models"
	"github.com/emicklei/go-restful/v3"
)

func (h *Handler) getCateRecommends(req *restful.Request, resp *restful.Response) {
	lists, err := mongo.GetCategoryRecommends()
	if err != nil {
		api.HandleError(resp, req, err)
	}

	res := &models.CmsCategoryRecommendListResponse{
		ResponseBase: models.ResponseBase{
			Code: 200,
			Msg:  api.Success,
		},
		Data: lists,
	}

	resp.WriteEntity(res)
}

func (h *Handler) setCateRecommends(req *restful.Request, resp *restful.Response) {
	var lists []models.CmsCategoryRecommend
	err := req.ReadEntity(&lists)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	err = mongo.SetCategoryRecommends(lists)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(&models.ResponseBase{Code: api.OK, Msg: api.Success})
}
