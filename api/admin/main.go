// 管理后台服务的启动入口。
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"XiaoLong-Ridy/api/admin/internal/config"
	"XiaoLong-Ridy/api/admin/internal/handler"
	"XiaoLong-Ridy/api/admin/internal/svc"
)

// main 负责加载配置、初始化 adminsvc RPC 客户端、注册 HTTP 路由并启动服务。
// 按 README 的目录约定，应在 api/admin 目录中执行启动命令。
func main() {
	cfgPath := filepath.Join("etc", "admin.json")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// ServiceContext 只持有 RPC 依赖，所有业务数据访问统一由 adminsvc 承担。
	ctx, err := svc.NewServiceContext(cfg)
	if err != nil {
		log.Fatalf("init service context: %v", err)
	}
	defer ctx.Close()

	// NewRouter 注册管理后台所有 HTTP API 路由及鉴权中间件。
	router := handler.NewRouter(ctx)
	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 优雅关停：监听 SIGINT/SIGTERM，收到信号后停止接收新连接并等待在途请求处理完成，
	// 避免发版/下线时请求被硬断。ListenAndServe 的错误通过独立 channel 传回主流程处理。
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("admin api listening on %s", cfg.HTTPAddr)
		serverErr <- server.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-quit:
		log.Printf("received signal %v, shutting down admin api gracefully", sig)
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("server stopped unexpectedly: %v", err)
			os.Exit(1)
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// 二次信号兜底：Shutdown 等待期间若再次收到信号或超时，强制退出，避免进程挂死。
	go func() {
		<-quit
		log.Println("received second signal, forcing exit")
		os.Exit(1)
	}()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed, forcing exit: %v", err)
		os.Exit(1)
	}
	log.Println("admin api shutdown complete")
}
