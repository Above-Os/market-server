package models

type CmsTopic struct {
	Name          string `json:"name"`
	Intro         string `json:"intro"`
	Desc          string `json:"desc"`
	IconSrc       string `json:"iconSrc"`
	DetailsImgSrc string `json:"detailsImgSrc"`
	RichText      string `json:"richText"`
}

type CmsTopicResponse struct {
	ResponseBase
	Data CmsTopic `json:"data"`
}

type CmsTopicListResponse struct {
	ResponseBase
	Data []*CmsTopic `json:"data"`
}

type CmsRecommend struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type CmsRecommendListResponse struct {
	ResponseBase
	Data []*CmsRecommend `json:"data"`
}

type CmsCategoryRecommend struct {
	Category string         `json:"category"`
	Data     []CmsRecommend `json:"data"`
}

type CmsCategoryRecommendListResponse struct {
	ResponseBase
	Data []*CmsCategoryRecommend `json:"data"`
}
