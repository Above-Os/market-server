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

package constants

const (
	APIServerListenAddress = ":8081"
	CMSServerListenAddress = ":8082"

	AppGitLocalDir    = "./app_git"
	AppGitZipLocalDir = "./charts"
	AppCfgFileName    = "app.cfg"

	TimeFormatStr = "Mon Jan 2 15:04:05 2006 -0700"

	DefaultPage     = 1
	DefaultPageSize = 100
	DefaultFrom     = 0

	MongoDBUri = "MONGODB_URI"

	EsAddr     = "ES_ADDR"
	EsName     = "ES_NAME"
	EsPassword = "ES_PASSWORD"
)

const (
	RemoveFile  = ".remove"
	SuspendFile = ".suspend"
	NsfwFile    = ".nsfw"

	RemoveLabel  = "remove"
	SuspendLabel = "suspend"
	NsfwLabel    = "nsfw"
)

var (
	//`{
	//   "appTypes":[
	//	   "AI",
	//	   "Mining",
	//	   "Protocol",
	//	   "Home",
	//	   "Data",
	//	   "Developer",
	//	   "Productivity",
	//	   "Multimedia",
	//	   "Utilities",
	//	   "Security",
	//	   "Game"
	//   ]
	//}`

	DefaultAppTypes = []string{"AI", "Mining", "Protocol", "Home",
		"Data", "Developer", "Productivity", "Multimedia",
		"Utilities", "Security", "Game"}
	DefaultAppTypeMaps = map[string]byte{"AI": 1, "Mining": 1, "Protocol": 1, "Home": 1,
		"Data": 1, "Developer": 1, "Productivity": 1, "Multimedia": 1,
		"Utilities": 1, "Security": 1, "Game": 1}
)
