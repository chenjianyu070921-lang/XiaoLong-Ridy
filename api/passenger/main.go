package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"XiaoLong-Ridy/api/passenger/internal/router"
	"XiaoLong-Ridy/api/passenger/internal/svc"
)

// main 是乘客端 API 网关入口；运行时统一通过真实 gRPC 调用下游微服务。
func main() {
	configFile := flag.String("f", "etc/passenger.yaml", "the config file")
	flag.Parse()

	cfg, err := svc.LoadRuntimeConfig(*configFile)
	if err != nil {
		panic(fmt.Errorf("加载 passenger api 配置失败: %w", err))
	}
	svcCtx, err := svc.NewServiceContextFromConfig(cfg)
	if err != nil {
		panic(fmt.Errorf("初始化 passenger api 依赖失败: %w", err))
	}
	defer svcCtx.Close()

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router.NewRouter(svcCtx),
	}

	log.Printf("passenger api started at http://127.0.0.1%s", cfg.HTTPAddr)
	log.Printf("passenger api client mode=%s grpc targets: usersvc=%s ordersvc=%s pricesvc=%s paysvc=%s dispatchsvc=%s", cfg.ClientMode, cfg.UserRPCAddr, cfg.OrderRPCAddr, cfg.PriceRPCAddr, cfg.PayRPCAddr, cfg.DispatchRPCAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("启动 passenger api 失败: %w", err))
	}
}
