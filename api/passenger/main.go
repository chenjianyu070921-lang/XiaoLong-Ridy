package main

import (
	"fmt"
	"log"
	"net/http"

	"XiaoLong-Ridy/api/passenger/internal/router"
	"XiaoLong-Ridy/api/passenger/internal/svc"
)

// main 是乘客端 API 网关入口；配置 RPC 地址时调用真实服务，未配置时回退本地联调客户端。
func main() {
	cfg := svc.LoadRuntimeConfigFromEnv()
	svcCtx, err := svc.NewServiceContextFromConfigWithSMSLogger(cfg, func(phone, code string) {
		// 本地联调没有真实短信通道，将验证码输出到日志，便于 Postman 或 curl 调试。
		log.Printf("本地短信验证码：phone=%s code=%s", phone, code)
	})
	if err != nil {
		panic(fmt.Errorf("初始化 passenger api 依赖失败: %w", err))
	}
	defer svcCtx.Close()

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router.NewRouter(svcCtx),
	}

	log.Printf("passenger api started at http://127.0.0.1%s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("启动 passenger api 失败: %w", err))
	}
}
