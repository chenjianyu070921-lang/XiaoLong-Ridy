package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/chat/internal/middleware"
	"XiaoLong-Ridy/api/chat/internal/svc"
)

// NewRouter 注册聊天的 4 个 HTTP 端点，均经 JWT 鉴权。
func NewRouter(svcCtx *svc.ServiceContext) http.Handler {
	mux := http.NewServeMux()
	auth := middleware.RequireAuth(svcCtx)

	mux.Handle("/api/chat/v1/conversation", auth(methodSwitch("GET", GetConversationHandler(svcCtx))))
	mux.Handle("/api/chat/v1/messages", auth(methodSwitch("GET", ListMessagesHandler(svcCtx))))
	mux.Handle("/api/chat/v1/send", auth(methodSwitch("POST", SendMessageHandler(svcCtx))))
	mux.Handle("/api/chat/v1/read", auth(methodSwitch("POST", MarkReadHandler(svcCtx))))

	return mux
}

// methodSwitch 仅允许指定 HTTP 方法，否则 405。
func methodSwitch(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"code":40500,"message":"method not allowed","data":null}`))
			return
		}
		h(w, r)
	}
}
