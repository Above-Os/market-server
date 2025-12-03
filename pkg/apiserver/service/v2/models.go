package v2

import (
	"app-store-server/pkg/models"
	"time"
)

// AppStoreInfo represents the appstore information structure
type AppStoreInfo struct {
	Apps  []models.ApplicationInfoEntry `json:"apps" bson:"apps"`
	Tops  []AppStoreTopItem             `json:"tops" bson:"tops"`
	Stats AppStoreStats                 `json:"stats" bson:"stats"`
}

// AppStoreStats represents statistics about the appstore
type AppStoreStats struct {
	TotalApps  int64  `json:"totalApps" bson:"totalApps"`
	TotalItems int64  `json:"totalItems" bson:"totalItems"`
	Hash       string `json:"hash" bson:"hash"`
}

// AppStoreTopItem represents a top ranked app item with only appid for ordering
type AppStoreTopItem struct {
	AppID string `json:"appId" bson:"appId"`
	Rank  int    `json:"rank" bson:"rank"`
}

// AppStoreInfoResponse represents the response structure for appstore info
type AppStoreInfoResponse struct {
	AppStore *AppStoreInfo `json:"appstore"`
}

// AppStoreHashResponse represents the response structure for appstore hash
type AppStoreHashResponse struct {
	Hash      string    `json:"hash"`
	UpdatedAt time.Time `json:"updatedAt"`
}
