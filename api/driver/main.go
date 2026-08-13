package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"XiaoLong-Ridy/api/driver/internal/handler"
	"XiaoLong-Ridy/api/driver/internal/svc"
)

const defaultHTTPAddress = ":8082"
const defaultDriverGRPCAddr = "127.0.0.1:8080"

func main() {
	address := os.Getenv("DRIVER_HTTP_ADDR")
	if address == "" {
		address = defaultHTTPAddress
	}
	grpcAddr := os.Getenv("DRIVER_GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = defaultDriverGRPCAddr
	}

	server := &http.Server{
		Addr:    address,
		Handler: newHTTPHandler(svc.NewServiceContext(grpcAddr)),
	}

	log.Printf("driver api started at http://127.0.0.1%s  (driversvc gRPC: %s)", address, grpcAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("启动 driver api 失败: %w", err))
	}
}

// newHTTPHandler 注册司机端全部 HTTP 路由。
func newHTTPHandler(svcCtx *svc.ServiceContext) http.Handler {
	mux := http.NewServeMux()
	// 司机
	mux.HandleFunc("/api/driver/v1/drivers", methodSwitch("POST", handler.CreateDriverHandler(svcCtx)))
	mux.HandleFunc("/api/driver/v1/drivers/update", handler.UpdateDriverHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/drivers/get", handler.GetDriverHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/drivers/delete", handler.DeleteDriverHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/drivers/list", handler.ListDriversHandler(svcCtx))
	// 车辆
	mux.HandleFunc("/api/driver/v1/vehicles", handler.CreateVehicleHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/vehicles/update", handler.UpdateVehicleHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/vehicles/get", handler.GetVehicleHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/vehicles/delete", handler.DeleteVehicleHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/vehicles/list", handler.ListVehiclesHandler(svcCtx))
	// 认证
	mux.HandleFunc("/api/driver/v1/certifications", handler.CreateCertificationHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/certifications/update", handler.UpdateCertificationHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/certifications/get", handler.GetCertificationHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/certifications/delete", handler.DeleteCertificationHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/certifications/list", handler.ListCertificationsHandler(svcCtx))
	// 服务分
	mux.HandleFunc("/api/driver/v1/scores", handler.CreateScoreHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/scores/update", handler.UpdateScoreHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/scores/get", handler.GetScoreHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/scores/delete", handler.DeleteScoreHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/scores/list", handler.ListScoresHandler(svcCtx))
	// 提现
	mux.HandleFunc("/api/driver/v1/withdraws", handler.CreateWithdrawHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/withdraws/update", handler.UpdateWithdrawHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/withdraws/get", handler.GetWithdrawHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/withdraws/delete", handler.DeleteWithdrawHandler(svcCtx))
	mux.HandleFunc("/api/driver/v1/withdraws/list", handler.ListWithdrawsHandler(svcCtx))
	return mux
}

// methodSwitch 限定路由仅接受指定方法，其余返回 405。
func methodSwitch(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}
