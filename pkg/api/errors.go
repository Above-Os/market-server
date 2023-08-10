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

package api

import (
	"app-store-server/pkg/utils"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/glog"
)

type ErrorType = string

const (
	ErrorInternalServerError ErrorType = "internal_server_error"
	ErrorInvalidGrant        ErrorType = "invalid_grant"
	ErrorBadRequest          ErrorType = "bad_request"
	ErrorUnknown             ErrorType = "unknown_error"
	ErrorIamOperator         ErrorType = "iam_operator"
)

const (
	OK = 0

	Success = "success"
)

// Avoid emitting errors that look like valid HTML. Quotes are okay.
var sanitizer = strings.NewReplacer(`&`, "&amp;", `<`, "&lt;", `>`, "&gt;")

func HandleInternalError(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusInternalServerError, response, req, err)
}

// HandleBadRequest writes http.StatusBadRequest and log error
func HandleBadRequest(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusBadRequest, response, req, err)
}

func HandleNotFound(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusNotFound, response, req, err)
}

func HandleForbidden(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusForbidden, response, req, err)
}

func HandleUnauthorized(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusUnauthorized, response, req, err)
}

func HandleTooManyRequests(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusTooManyRequests, response, req, err)
}

func HandleConflict(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusConflict, response, req, err)
}

func HandleError(response *restful.Response, req *restful.Request, err error) {
	var statusCode int
	switch t := err.(type) {
	case restful.ServiceError:
		statusCode = t.Code
	default:
		statusCode = http.StatusInternalServerError
	}
	handle(statusCode, response, req, err)
}

func handle(statusCode int, resp *restful.Response, req *restful.Request, err error) {
	_, fn, line, _ := runtime.Caller(2)
	glog.Errorf("%s:%d %v", fn, line, err)

	var errType ErrorType
	var errDesc string

	if t, ok := err.(Error); ok {
		resp.WriteHeaderAndEntity(statusCode, t)
		return
	}

	switch statusCode {
	case http.StatusBadRequest:
		errType = ErrorBadRequest
	case http.StatusUnauthorized, http.StatusForbidden:
		errType = ErrorInvalidGrant
	case http.StatusInternalServerError:
		errType = ErrorInternalServerError
	default:
		errType = ErrorUnknown
	}
	errDesc = err.Error()
	resp.WriteHeaderAndEntity(statusCode, Error{
		ErrorType:        errType,
		ErrorDescription: errDesc,
	})
}

type Error struct {
	ErrorType        string `json:"error_type"`
	ErrorDescription string `json:"error_description"`
}

func (e Error) Error() string {
	return utils.PrettyJSON(e)
}

func NewError(t string, errs ...string) Error {
	var desc string
	if len(errs) > 0 {
		desc = errs[0]
	}
	return Error{ErrorType: t, ErrorDescription: desc}
}

func ErrorWithMessage(err error, message string) error {
	return fmt.Errorf("%v: %v", message, err.Error())
}

type ErrorMessage struct {
	Message string `json:"message"`
}

func (e ErrorMessage) Error() string {
	return e.Message
}

var ErrorNone = ErrorMessage{Message: "success"}
