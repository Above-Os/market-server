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

# Build stage
FROM golang:1.24.10 as builder

WORKDIR /workspace

# Copy go.mod and go.sum to cache module downloads
COPY go.mod go.sum ./ 

RUN go mod download

# Copy the remaining source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o app-store-server cmd/app-store-server/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:latest

RUN apk update && \
    apk upgrade &&  \
    apk add --no-cache bash git openssh docker-cli curl

WORKDIR /opt/app
COPY --from=builder /workspace/app-store-server .

CMD ["/opt/app/app-store-server", "-v", "4", "--logtostderr"]
