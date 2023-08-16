# Copyright 2022 bytetrade
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: app-store-server fmt vet

all: app-store-server

tidy: 
	go mod tidy
	
fmt: ;$(info $(M)...Begin to run go fmt against code.) @
	go fmt ./...

vet: ;$(info $(M)...Begin to run go vet against code.) @
	go vet ./...

app-store-server: tidy fmt vet ;$(info $(M)...Begin to build app-store-server.) @
	go build -o output/app-store-server cmd/app-store-server/main.go


run: fmt vet; $(info $(M)...Run app-store-server.)
	go run cmd/app-store-server/main.go -v 4 --logtostderr

dev: fmt vet; $(info $(M)...Run app-store-server.)
	export MONGODB_URI='mongodb://root:123456@localhost:27017' && export ES_ADDR='https://localhost:9200' && export ES_NAME='elastic' && export ES_PASSWORD='WVF+CRh+oHV+J8ZTV4lC' && go run cmd/app-store-server/main.go -v 4 --logtostderr


.PHONY: cms
cms:
	go build -o output/app-store-admin-server ./cmd/app-store-admin-server/main.go

cmsdev:
	export MONGODB_URI='mongodb://root:123456@localhost:27017' && go run cmd/app-store-admin-server/main.go -v 4 --logtostderr
