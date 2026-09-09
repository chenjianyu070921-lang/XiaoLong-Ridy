package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	// 在独立 goroutine 中启动 HTTP 服务，主 goroutine 专门负责接收退出信号。
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	// 捕获 Ctrl+C/SIGTERM，给正在处理的请求最多 10 秒完成清理。
	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("启动 passenger api 失败: %w", err))
		}
	case <-stopCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("passenger api 优雅关闭失败: %v", err)
			return
		}
		log.Printf("passenger api 已完成优雅关闭")
	}
}
