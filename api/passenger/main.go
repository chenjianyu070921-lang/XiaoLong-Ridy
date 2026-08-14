package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"XiaoLong-Ridy/api/passenger/internal/router"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	orderclient "XiaoLong-Ridy/rpc/ordersvc/client"
	priceclient "XiaoLong-Ridy/rpc/pricesvc/client"
	"XiaoLong-Ridy/rpc/usersvc/client"
)

const defaultHTTPAddress = ":8091"
const localSigningKey = "local-development-signing-key"

// main 是乘客端 API 网关入口，默认注入本地联调版 usersvc、pricesvc 和 ordersvc 客户端。
func main() {
	address := os.Getenv("PASSENGER_HTTP_ADDR")
	if address == "" {
		address = defaultHTTPAddress
	}

	userClient := client.NewLocalClient(localSigningKey, func(phone, code string) {
		// 本地联调没有真实短信通道，将验证码输出到日志，便于 Postman 或 curl 调试。
		log.Printf("本地短信验证码：phone=%s code=%s", phone, code)
	})
	svcCtx := svc.NewServiceContext(
		userClient,
		svc.WithOrderClient(orderclient.NewLocalClient()),
		svc.WithPriceClient(priceclient.NewLocalClient()),
		svc.WithTokenSigningKey(localSigningKey),
	)
	server := &http.Server{
		Addr:    address,
		Handler: router.NewRouter(svcCtx),
	}

	log.Printf("passenger api started at http://127.0.0.1%s", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("启动 passenger api 失败: %w", err))
	}
}
