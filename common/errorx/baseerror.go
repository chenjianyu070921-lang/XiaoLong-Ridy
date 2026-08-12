package errorx

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultCode = 1001

// CodeError 统一错误类型
type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// CodeErrorResponse 统一错误返回格式
type CodeErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewCodeError(code int, msg string) error {
	return &CodeError{Code: code, Msg: msg}
}

func NewDefaultError(msg string) error {
	return NewCodeError(defaultCode, msg)
}

func (e *CodeError) Error() string {
	return e.Msg
}

func (e *CodeError) Data() *CodeErrorResponse {
	return &CodeErrorResponse{
		Code: e.Code,
		Msg:  e.Msg,
	}
}

// GrpcError 用于 gRPC 错误转换
func GrpcError(err error) error {
	if err == nil {
		return nil
	}
	if codeErr, ok := err.(*CodeError); ok {
		return status.Error(codes.Code(codeErr.Code), codeErr.Msg)
	}
	return status.Error(codes.Code(defaultCode), err.Error())
}
