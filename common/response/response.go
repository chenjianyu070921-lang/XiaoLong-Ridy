package response

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"XiaoLong-Ridy/common/errorx"
)

// Body 统一返回结构
type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

var bgCtx = context.Background()

// Success 成功返回
func Success(w http.ResponseWriter, data interface{}) {
	httpx.OkJsonCtx(bgCtx, w, &Body{
		Code: errorx.OK,
		Msg:  errorx.GetMsg(errorx.OK),
		Data: data,
	})
}

// Error 错误返回
func Error(w http.ResponseWriter, code int, msg string) {
	httpx.WriteJsonCtx(bgCtx, w, http.StatusOK, &Body{
		Code: code,
		Msg:  msg,
	})
}

// ErrorWithCode 根据错误码返回
func ErrorWithCode(w http.ResponseWriter, code int) {
	httpx.WriteJsonCtx(bgCtx, w, http.StatusOK, &Body{
		Code: code,
		Msg:  errorx.GetMsg(code),
	})
}

// ParamError 参数错误快捷返回
func ParamError(w http.ResponseWriter, msg string) {
	if msg == "" {
		msg = errorx.GetMsg(errorx.ParamError)
	}
	Error(w, errorx.ParamError, msg)
}
