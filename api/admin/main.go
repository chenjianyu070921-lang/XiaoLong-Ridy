// 管理后台服务的启动入口。
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

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
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	log.Printf("admin api listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
