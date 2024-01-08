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

package utils

import (
	"app-store-server/internal/constants"
	"bytes"
	"encoding/json"
	"strconv"
)

func ToJSON(v any) string {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		panic(err)
	}
	return buf.String()
}

func PrettyJSON(v any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	return buf.String()
}

func VerifyFromAndSize(page, size string) (int, int) {
	pageN, err := strconv.Atoi(page)
	if pageN < 1 || err != nil {
		pageN = constants.DefaultPage
	}

	sizeN, err := strconv.Atoi(size)
	if sizeN < 1 || err != nil {
		sizeN = constants.DefaultPageSize
	}

	from := (pageN - 1) * sizeN
	if from < 0 {
		from = constants.DefaultFrom
	}

	return from, sizeN
}

func VerifyTopSize(size string) int {
	sizeN, err := strconv.Atoi(size)
	if sizeN < 1 || sizeN > 2*constants.DefaultTopCount || err != nil {
		sizeN = constants.DefaultTopCount
	}

	return sizeN
}
