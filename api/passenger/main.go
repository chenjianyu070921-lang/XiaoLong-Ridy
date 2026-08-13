package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"XiaoLong-Ridy/api/passenger/internal/handler"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/rpc/usersvc/client"
)

const defaultHTTPAddress = ":8091"

func main() {
	address := os.Getenv("PASSENGER_HTTP_ADDR")
	if address == "" {
		address = defaultHTTPAddress
	}

	userClient := client.NewLocalClient("local-development-signing-key", func(phone, code string) {
		// 本地联调没有真实短信通道，将验证码输出到日志供 Postman/curl 使用。
		log.Printf("本地短信验证码：phone=%s code=%s", phone, code)
	})
	server := &http.Server{
		Addr:    address,
		Handler: newHTTPHandler(svc.NewServiceContext(userClient)),
	}

	log.Printf("passenger api started at http://127.0.0.1%s", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("启动 passenger api 失败: %w", err))
	}
}

// newHTTPHandler 注册乘客端当前已实现的认证路由。
func newHTTPHandler(svcCtx *svc.ServiceContext) http.Handler {
	mux := http.NewServeMux()
	// 登录前接口不需要 JWT；后续受保护接口统一在此处挂载鉴权中间件。
	mux.HandleFunc("/api/passenger/v1/auth/send-sms-code", handler.SendSMSCodeHandler(svcCtx))
	mux.HandleFunc("/api/passenger/v1/auth/login-by-sms", handler.LoginBySMSHandler(svcCtx))
	mux.HandleFunc("/api/passenger/v1/auth/refresh-token", handler.RefreshTokenHandler(svcCtx))
	mux.HandleFunc("/api/passenger/v1/auth/logout", handler.LogoutHandler(svcCtx))
	return mux
}
