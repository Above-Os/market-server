package models

type ResponseBase struct {
	Code int    `json:"code"`
	Msg  string `json:"message,omitempty"`
}

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"message,omitempty"`
	Data any    `json:"data,omitempty"`
}

func NewResponse(code int, msg string, data any) *Response {
	return &Response{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}
