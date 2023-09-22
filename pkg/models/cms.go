package models

type CmsTopic struct {
	Id           string            `json:"id"`
	Name         string            `json:"name"`
	Introduction string            `json:"introduction"`
	Des          string            `json:"des"`
	IconImg      string            `json:"iconimg"`
	DetailImg    string            `json:"detailimg"`
	RichText     string            `json:"richtext"`
	AppList      []ApplicationInfo `json:"appList"`
}

type CmsTopicListResponse struct {
	ResponseBase
	Data []*CmsTopic `json:"data"`
}

type CmsRecommend struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type CmsCategoryRecommend struct {
	Category string         `json:"category"`
	Data     []CmsRecommend `json:"data"`
}

type CmsCategoryRecommendListResponse struct {
	ResponseBase
	Data []*CmsCategoryRecommend `json:"data"`
}
