package main

import (
	"XiaoLong-Ridy/api/passenger/internal/router"
	"fmt"
	"log"
	"net/http"
	"os"

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
		Handler: router.NewRouter(svc.NewServiceContext(userClient)),
	}

	log.Printf("passenger api started at http://127.0.0.1%s", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("启动 passenger api 失败: %w", err))
	}
}
