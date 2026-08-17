package router

import (
	"XiaoLong-Ridy/api/passenger/internal/handler"
	"net/http"

	"XiaoLong-Ridy/api/passenger/internal/svc"
)

// NewRouter 创建乘客端 HTTP 路由入口，统一注册当前已实现的认证接口。
func NewRouter(svcCtx *svc.ServiceContext) http.Handler {
	mux := http.NewServeMux()
	registerAuthRoutes(mux, svcCtx)
	return mux
}

// registerAuthRoutes 注册乘客端登录注册相关路由，登录前接口不需要 JWT。
func registerAuthRoutes(mux *http.ServeMux, svcCtx *svc.ServiceContext) {
	mux.HandleFunc("/api/passenger/v1/auth/send-sms-code", handler.SendSMSCodeHandler(svcCtx))
	mux.HandleFunc("/api/passenger/v1/auth/login-by-sms", handler.LoginBySMSHandler(svcCtx))
	mux.HandleFunc("/api/passenger/v1/auth/refresh-token", handler.RefreshTokenHandler(svcCtx))
	mux.HandleFunc("/api/passenger/v1/auth/logout", handler.LogoutHandler(svcCtx))
}
