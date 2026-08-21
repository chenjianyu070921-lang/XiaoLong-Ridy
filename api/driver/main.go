package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"XiaoLong-Ridy/api/driver/internal/handler"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
)

const defaultHTTPAddress = ":8082"

// driversvc listens on rpc/driversvc/etc/driversvc.yaml ListenOn by default.
const defaultDriverGRPCAddr = "127.0.0.1:5055"

const defaultOrderGRPCAddr = "127.0.0.1:50051"

func main() {
	address := os.Getenv("DRIVER_HTTP_ADDR")
	if address == "" {
		address = defaultHTTPAddress
	}

	driverGRPCAddr := os.Getenv("DRIVER_GRPC_ADDR")
	if driverGRPCAddr == "" {
		driverGRPCAddr = defaultDriverGRPCAddr
	}

	orderGRPCAddr := os.Getenv("ORDER_GRPC_ADDR")
	if orderGRPCAddr == "" {
		orderGRPCAddr = defaultOrderGRPCAddr
	}

	server := &http.Server{
		Addr:    address,
		Handler: newHTTPHandler(svc.NewServiceContext(driverGRPCAddr, orderGRPCAddr)),
	}

	log.Printf("driver api started at http://127.0.0.1%s  (driversvc gRPC: %s, ordersvc gRPC: %s)", address, driverGRPCAddr, orderGRPCAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("start driver api: %w", err))
	}
}

func newHTTPHandler(svcCtx *svc.ServiceContext) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/driver/v1/auth/send-sms-code", methodSwitch("POST", handler.SendSMSCodeHandler(svcCtx)))
	mux.HandleFunc("/api/driver/v1/auth/login-by-password", methodSwitch("POST", handler.LoginByPasswordHandler(svcCtx)))
	mux.HandleFunc("/api/driver/v1/auth/login-by-sms", methodSwitch("POST", handler.LoginBySMSHandler(svcCtx)))
	mux.HandleFunc("/api/driver/v1/drivers/register", methodSwitch("POST", handler.RegisterDriverHandler(svcCtx)))

	protected := middleware.RequireAuth(svcCtx)
	mux.Handle("/api/driver/v1/drivers", protected(methodSwitch("POST", handler.CreateDriverHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/update", protected(handler.UpdateDriverHandler(svcCtx)))
	mux.Handle("/api/driver/v1/drivers/get", protected(handler.GetDriverHandler(svcCtx)))
	mux.Handle("/api/driver/v1/drivers/delete", protected(handler.DeleteDriverHandler(svcCtx)))
	mux.Handle("/api/driver/v1/drivers/online", protected(methodSwitch("POST", handler.SetOnlineHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/offline", protected(methodSwitch("POST", handler.SetOfflineHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/heartbeat", protected(methodSwitch("POST", handler.HeartbeatHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/location/report", protected(methodSwitch("POST", handler.ReportLocationHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/ai-score", protected(methodSwitch("GET", handler.GetDriverAiScoreHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/certification/upload", protected(methodSwitch("POST", handler.UploadCertificationHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/certification", protected(methodSwitch("GET", handler.GetCertificationHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/accept", protected(methodSwitch("POST", handler.AcceptOrderHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/start-trip", protected(methodSwitch("POST", handler.StartTripHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/confirm-arrive", protected(methodSwitch("POST", handler.ConfirmArriveHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/finish-trip", protected(methodSwitch("POST", handler.FinishTripHandler(svcCtx))))
	mux.Handle("/api/driver/v1/agent/chat", internalOrDriverAuth(svcCtx, methodSwitch("POST", handler.AgentChatHandler())))

	return mux
}

func internalOrDriverAuth(svcCtx *svc.ServiceContext, h http.HandlerFunc) http.Handler {
	protected := middleware.RequireAuth(svcCtx)(h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serviceToken := os.Getenv("DRIVER_AGENT_SERVICE_TOKEN")
		if serviceToken != "" && r.Header.Get("X-Internal-Service-Token") == serviceToken {
			h(w, r)
			return
		}
		protected.ServeHTTP(w, r)
	})
}

func methodSwitch(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}
