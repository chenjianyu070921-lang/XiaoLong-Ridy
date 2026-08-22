package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"XiaoLong-Ridy/api/driver/internal/handler"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/ws"
)

// envOr 读取环境变量，缺失时回退到默认值。
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	address := envOr("DRIVER_HTTP_ADDR", ":8082")
	grpcAddr := envOr("DRIVER_GRPC_ADDR", "127.0.0.1:8080")
	locationGRPCAddr := envOr("LOCATION_GRPC_ADDR", "127.0.0.1:9001")
	redisAddr := envOr("DRIVER_REDIS_ADDR", "127.0.0.1:6379")

	server := &http.Server{
		Addr:    address,
		Handler: newHTTPHandler(svc.NewServiceContext(grpcAddr, locationGRPCAddr, redisAddr)),
	}

	log.Printf("driver api started at http://127.0.0.1%s (driversvc gRPC: %s, locationsvc gRPC: %s, redis: %s)",
		address, grpcAddr, locationGRPCAddr, redisAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("启动 driver api 失败: %w", err))
	}
}

// newHTTPHandler 注册司机端所有 HTTP/WS 路由。
func newHTTPHandler(svcCtx *svc.ServiceContext) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/driver/v1/drivers", handler.CreateDriverHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/drivers/update", handler.UpdateDriverHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/drivers/get", handler.GetDriverHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/drivers/delete", handler.DeleteDriverHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/drivers/report-location", handler.ReportLocationHandler(svcCtx))
	mux.HandleFunc("/ws/location", ws.LocationWSHandler(svcCtx.Redis))
	return mux
}
