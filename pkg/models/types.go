// Copyright 2022 bytetrade
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"app-store-server/pkg/models/tapr"
)

type TopResultItem struct {
	Category string             `json:"category"`
	Apps     []*ApplicationInfo `json:"apps"`
}

type ApplicationInfo struct {
	Id string `yaml:"id" json:"id" bson:"_id"`

	Name        string `yaml:"name" json:"name" bson:"name"`
	ChartName   string `yaml:"chartName" json:"chartName" bson:"chartName"`
	Icon        string `yaml:"icon" json:"icon" bson:"icon"`
	Description string `yaml:"desc" json:"desc" bson:"desc"`
	AppID       string `yaml:"appid" json:"appid" bson:"appid"`
	Title       string `yaml:"title" json:"title" bson:"title"`
	Version     string `yaml:"version" json:"version" bson:"version"`
	Categories  string `yaml:"categories" json:"categories" bson:"categories"` //[]string
	VersionName string `yaml:"versionName" json:"versionName" bson:"versionName"`

	FullDescription    string           `yaml:"fullDescription" json:"fullDescription" bson:"fullDescription"`
	UpgradeDescription string           `yaml:"upgradeDescription" json:"upgradeDescription" bson:"upgradeDescription"`
	PromoteImage       []string         `yaml:"promoteImage" json:"promoteImage" bson:"promoteImage"`
	PromoteVideo       string           `yaml:"promoteVideo" json:"promoteVideo" bson:"promoteVideo"`
	SubCategory        string           `yaml:"subCategory" json:"subCategory" bson:"subCategory"`
	Language           []string         `yaml:"language" json:"language" bson:"language"`
	Developer          string           `yaml:"developer" json:"developer" bson:"developer"`
	RequiredMemory     string           `yaml:"requiredMemory" json:"requiredMemory" bson:"requiredMemory"`
	RequiredDisk       string           `yaml:"requiredDisk" json:"requiredDisk" bson:"requiredDisk"`
	SupportClient      SupportClient    `yaml:"supportClient" json:"supportClient" bson:"supportClient"`
	RequiredGPU        string           `yaml:"requiredGpu" json:"requiredGpu,omitempty" bson:"requiredGpu"`
	RequiredCPU        string           `yaml:"requiredCpu" json:"requiredCpu" bson:"requiredCpu"`
	Rating             float32          `yaml:"rating" json:"rating" bson:"rating"`
	Target             string           `yaml:"target" json:"target" bson:"target"`
	Permission         Permission       `yaml:"permission" json:"permission"  bson:"permission" description:"app permission request"`
	Entrance           AppService       `yaml:"entrance" json:"entrance" bson:"entrance"`
	Middleware         *tapr.Middleware `yaml:"middleware" json:"middleware" bson:"middleware" description:"app middleware request"`
	Options            Options          `yaml:"options" json:"options" bson:"options" description:"app options"`

	Submitter string       `yaml:"submitter" json:"submitter" bson:"submitter"`
	Doc       string       `yaml:"doc" json:"doc" bson:"doc"`
	Website   string       `yaml:"website" json:"website" bson:"website"`
	License   []TextAndURL `yaml:"license" json:"license" bson:"license"`
	Legal     []TextAndURL `yaml:"legal" json:"legal" bson:"legal"`

	LastCommitHash string `yaml:"-" json:"lastCommitHash" bson:"lastCommitHash"`
	CreateTime     int64  `yaml:"-" json:"createTime" bson:"createTime"`
	UpdateTime     int64  `yaml:"-" json:"updateTime" bson:"updateTime"`
	Status         string `yaml:"status" json:"status" bson:"status"`
}

type AppService struct {
	Name string `yaml:"name" json:"name" bson:"name"`
	Port int32  `yaml:"port" json:"port" bson:"port"`
}

type AppSpec struct {
	VersionName        string        `yaml:"versionName" json:"versionName"`
	FullDescription    string        `yaml:"fullDescription" json:"fullDescription"`
	UpgradeDescription string        `yaml:"upgradeDescription" json:"upgradeDescription"`
	PromoteImage       []string      `yaml:"promoteImage" json:"promoteImage"`
	PromoteVideo       string        `yaml:"promoteVideo" json:"promoteVideo"`
	SubCategory        string        `yaml:"subCategory" json:"subCategory"`
	Language           []string      `yaml:"language" json:"language"`
	Developer          string        `yaml:"developer" json:"developer"`
	RequiredMemory     string        `yaml:"requiredMemory" json:"requiredMemory"`
	RequiredDisk       string        `yaml:"requiredDisk" json:"requiredDisk"`
	SupportClient      SupportClient `yaml:"supportClient" json:"supportClient"`
	RequiredGPU        string        `yaml:"requiredGpu" json:"requiredGpu,omitempty"`
	RequiredCPU        string        `yaml:"requiredCpu" json:"requiredCpu"`

	Submitter string       `yaml:"submitter" json:"submitter"`
	Doc       string       `yaml:"doc" json:"doc"`
	Website   string       `yaml:"website" json:"website"`
	License   []TextAndURL `yaml:"license" json:"license"`
	Legal     []TextAndURL `yaml:"legal" json:"legal"`
}

type TextAndURL struct {
	Text string `yaml:"text" json:"text" bson:"text"`
	URL  string `yaml:"url" json:"url" bson:"url"`
}

type SupportClient struct {
	Edge    string `yaml:"edge" json:"edge" bson:"edge"`
	Android string `yaml:"android" json:"android" bson:"android"`
	Ios     string `yaml:"ios" json:"ios" bson:"ios"`
	Windows string `yaml:"windows" json:"windows" bson:"windows"`
	Mac     string `yaml:"mac" json:"mac" bson:"mac"`
	Linux   string `yaml:"linux" json:"linux" bson:"linux"`
}

type Permission struct {
	AppData bool         `yaml:"appData" json:"appData" bson:"appData"  description:"app data permission for writing"`
	SysData []SysDataCfg `yaml:"sysData" json:"sysData" bson:"sysData"  description:"system shared data permission for accessing"`
}

type SysDataCfg struct {
	Group    string   `yaml:"group" json:"group"`
	DataType string   `yaml:"dataType" json:"dataType"`
	Version  string   `yaml:"version" json:"version"`
	Ops      []string `yaml:"ops" json:"ops"`
}

type Policy struct {
	Description string `yaml:"description" json:"description" bson:"description" description:"the description of the policy"`
	URIRegex    string `yaml:"uriRegex" json:"uriRegex" description:"uri regular expression"`
	Level       string `yaml:"level" json:"level"`
	OneTime     bool   `yaml:"oneTime" json:"oneTime"`
	Duration    string `yaml:"validDuration" json:"validDuration"`
}

type Analytics struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type Options struct {
	Policies     []Policy     `yaml:"policies" json:"policies" bson:"policies"`
	Analytics    Analytics    `yaml:"analytics" json:"analytics" bson:"analytics"`
	Dependencies []Dependency `yaml:"dependencies" json:"dependencies" bson:"dependencies"`
}

type Dependency struct {
	Name    string `yaml:"name" json:"name" bson:"name"`
	Version string `yaml:"version" json:"version" bson:"version"`
	// dependency type: system, application.
	Type string `yaml:"type" json:"type" bson:"type"`
}

type ExistRes struct {
	Exist bool `json:"exist"`
}
