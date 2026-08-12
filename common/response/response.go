package response

import (
	"encoding/json"
	"net/http"

	"XiaoLong-Ridy/common/errorx"
)

// Result 统一返回格式
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Ok 返回成功
func Ok(w http.ResponseWriter, data interface{}) {
	b, _ := json.Marshal(Result{Code: errorx.OK, Msg: errorx.Msg(errorx.OK), Data: data})
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// Fail 返回失败
func Fail(w http.ResponseWriter, code int) {
	b, _ := json.Marshal(Result{Code: code, Msg: errorx.Msg(code)})
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
