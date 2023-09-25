package models

import (
	"strings"

	"app-store-server/pkg/models/tapr"
)

const (
	AppCfgFileName = "app.cfg"
)

/*
app.cfg

app.cfg.version: v1
metadata:
  name: <chart name>
  description: <desc>
  icon: <icon file uri>
  appid: <app register id>
  version: <app version>
  title: <app title>
*/

type AppMetaData struct {
	Name        string   `yaml:"name" json:"name"`
	Icon        string   `yaml:"icon" json:"icon"`
	Description string   `yaml:"description" json:"description"`
	AppID       string   `yaml:"appid" json:"appid"`
	Title       string   `yaml:"title" json:"title"`
	Version     string   `yaml:"version" json:"version"`
	Categories  []string `yaml:"categories" json:"categories"`
	Rating      float32  `yaml:"rating" json:"rating"`
	Target      string   `yaml:"target" json:"target"`
}

type AppConfiguration struct {
	ConfigVersion string           `yaml:"app.cfg.version" json:"app.cfg.version"`
	Metadata      AppMetaData      `yaml:"metadata" json:"metadata"`
	Entrance      AppService       `yaml:"entrance" json:"entrance"`
	Spec          AppSpec          `yaml:"spec" json:"spec"`
	Permission    Permission       `yaml:"permission" json:"permission" description:"app permission request"`
	Middleware    *tapr.Middleware `yaml:"middleware" json:"middleware" description:"app middleware request"`
	Options       Options          `yaml:"options" json:"options" description:"app options"`
}

func (ac *AppConfiguration) ToAppInfo() *ApplicationInfo {
	return &ApplicationInfo{
		AppID:              ac.Metadata.AppID,
		Name:               ac.Metadata.Name,
		Icon:               ac.Metadata.Icon,
		Description:        ac.Metadata.Description,
		Title:              ac.Metadata.Title,
		Version:            ac.Metadata.Version,
		Categories:         strings.Join(ac.Metadata.Categories, ","),
		VersionName:        ac.Spec.VersionName,
		FullDescription:    ac.Spec.FullDescription,
		UpgradeDescription: ac.Spec.UpgradeDescription,
		PromoteImage:       ac.Spec.PromoteImage,
		PromoteVideo:       ac.Spec.PromoteVideo,
		SubCategory:        ac.Spec.SubCategory,
		Developer:          ac.Spec.Developer,
		RequiredMemory:     ac.Spec.RequiredMemory,
		RequiredDisk:       ac.Spec.RequiredDisk,
		SupportClient:      ac.Spec.SupportClient,
		RequiredGPU:        ac.Spec.RequiredGPU,
		RequiredCPU:        ac.Spec.RequiredCPU,
		Rating:             ac.Metadata.Rating,
		Target:             ac.Metadata.Target,
		Permission:         ac.Permission,
		Entrance:           ac.Entrance,
		Middleware:         ac.Middleware,
		Options:            ac.Options,
	}
}
