package models

type ResponseBase struct {
	Code int    `json:"code"`
	Msg  string `json:"message,omitempty"`
}

type Response struct {
	ResponseBase
	Data any `json:"data,omitempty"`
}

func NewResponse(code int, msg string, data any) *Response {
	return &Response{
		ResponseBase: ResponseBase{
			Code: code,
			Msg:  msg,
		},

		Data: data,
	}
}
