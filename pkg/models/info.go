package models

import (
	"app-store-server/pkg/models/tapr"
)

/*
OlaresManifest.yaml

olaresManifest.version: v1
olaresManifest.type: app/recommend/model/agent
metadata:
  name: <chart name>
  description: <desc>
  icon: <icon file uri>
  appid: <app register id>
  version: <app version>
  title: <app title>
*/

type AppMetaData struct {
	Name string `yaml:"name" json:"name"`
	//Type        string   `yaml:"type" json:"type"`
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
	ConfigVersion string           `yaml:"olaresManifest.version" json:"olaresManifest.version"`
	ConfigType    string           `yaml:"olaresManifest.type" json:"olaresManifest.type"`
	Metadata      AppMetaData      `yaml:"metadata" json:"metadata"`
	Entrances     []Entrance       `yaml:"entrances" json:"entrances"`
	Spec          AppSpec          `yaml:"spec" json:"spec"`
	Permission    Permission       `yaml:"permission" json:"permission" description:"app permission request"`
	Middleware    *tapr.Middleware `yaml:"middleware" json:"middleware" description:"app middleware request"`
	Options       Options          `yaml:"options" json:"options" description:"app options"`
}

type Entrance struct {
	Name      string `yaml:"name" json:"name" bson:"name"`
	Host      string `yaml:"host" json:"host" bson:"host"`
	Port      int32  `yaml:"port" json:"port" bson:"port"`
	Icon      string `yaml:"icon,omitempty" json:"icon,omitempty" bson:"icon,omitempty"`
	Title     string `yaml:"title" json:"title" bson:"title"`
	AuthLevel string `yaml:"authLevel,omitempty" json:"authLevel,omitempty" bson:"authLevel,omitempty"`
	Invisible bool   `yaml:"invisible,omitempty" json:"invisible,omitempty" bson:"invisible,omitempty"`
}

func (ac *AppConfiguration) ToAppInfo() *ApplicationInfoEntry {
	return &ApplicationInfoEntry{
		AppID:              ac.Metadata.AppID,
		CfgType:            ac.ConfigType,
		Name:               ac.Metadata.Name,
		Icon:               ac.Metadata.Icon,
		Description:        ac.Metadata.Description,
		Title:              ac.Metadata.Title,
		Version:            ac.Metadata.Version,
		Categories:         ac.Metadata.Categories,
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
		SupportArch:        ac.Spec.SupportArch,
		RequiredGPU:        ac.Spec.RequiredGPU,
		RequiredCPU:        ac.Spec.RequiredCPU,
		Rating:             ac.Metadata.Rating,
		Target:             ac.Metadata.Target,
		Permission:         ac.Permission,
		Entrances:          ac.Entrances,
		Middleware:         ac.Middleware,
		Options:            ac.Options,
		Locale:             ac.Spec.Locale,
		Submitter:          ac.Spec.Submitter,
		Doc:                ac.Spec.Doc,
		Website:            ac.Spec.Website,
		FeaturedImage:      ac.Spec.FeaturedImage,
		SourceCode:         ac.Spec.SourceCode,
		License:            ac.Spec.License,
		Legal:              ac.Spec.Legal,
		ModelSize:          ac.Spec.ModelSize,
		Namespace:          ac.Spec.Namespace,
		OnlyAdmin:          ac.Spec.OnlyAdmin,
	}
}
