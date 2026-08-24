package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"XiaoLong-Ridy/api/driver/internal/handler"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"

	"gopkg.in/yaml.v3"
)

const defaultHTTPAddress = ":8082"

// driversvc listens on rpc/driversvc/etc/driversvc.yaml ListenOn by default.
const defaultDriverGRPCAddr = "127.0.0.1:5055"

const defaultOrderGRPCAddr = "127.0.0.1:50051"

const defaultDispatchGRPCAddr = "127.0.0.1:8083"

const defaultRedisAddr = ""

type driverConfig struct {
	HTTPAddr         string `yaml:"httpAddr"`
	DriverGRPCAddr   string `yaml:"driverGrpcAddr"`
	OrderGRPCAddr    string `yaml:"orderGrpcAddr"`
	DispatchGRPCAddr string `yaml:"dispatchGrpcAddr"`
	RedisAddr        string `yaml:"redisAddr"`
}

func main() {
	configPath := flag.String("f", "etc/driver.yaml", "driver api config file")
	flag.Parse()

	cfg, err := loadDriverConfig(*configPath)
	if err != nil {
		panic(fmt.Errorf("load driver api config: %w", err))
	}
	address := envOr("DRIVER_HTTP_ADDR", cfg.HTTPAddr)
	driverGRPCAddr := envOr("DRIVER_GRPC_ADDR", cfg.DriverGRPCAddr)
	orderGRPCAddr := envOr("ORDER_GRPC_ADDR", cfg.OrderGRPCAddr)
	dispatchGRPCAddr := envOr("DISPATCH_GRPC_ADDR", cfg.DispatchGRPCAddr)
	redisAddr := envOr("DRIVER_REDIS_ADDR", cfg.RedisAddr)

	svcCtx := svc.NewServiceContext(driverGRPCAddr, orderGRPCAddr, dispatchGRPCAddr, redisAddr)
	if err := svcCtx.ValidateSigningKey(); err != nil {
		panic(fmt.Errorf("driver api signing key check: %w", err))
	}

	server := &http.Server{
		Addr:    address,
		Handler: newHTTPHandler(svcCtx),
	}

	log.Printf("driver api started at http://127.0.0.1%s  (driversvc gRPC: %s, ordersvc gRPC: %s, dispatchsvc gRPC: %s, redis: %s)", address, driverGRPCAddr, orderGRPCAddr, dispatchGRPCAddr, redisAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("start driver api: %w", err))
	}
}

func loadDriverConfig(path string) (driverConfig, error) {
	cfg := driverConfig{
		HTTPAddr:         defaultHTTPAddress,
		DriverGRPCAddr:   defaultDriverGRPCAddr,
		OrderGRPCAddr:    defaultOrderGRPCAddr,
		DispatchGRPCAddr: defaultDispatchGRPCAddr,
		RedisAddr:        defaultRedisAddr,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = defaultHTTPAddress
	}
	if cfg.DriverGRPCAddr == "" {
		cfg.DriverGRPCAddr = defaultDriverGRPCAddr
	}
	if cfg.OrderGRPCAddr == "" {
		cfg.OrderGRPCAddr = defaultOrderGRPCAddr
	}
	if cfg.DispatchGRPCAddr == "" {
		cfg.DispatchGRPCAddr = defaultDispatchGRPCAddr
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
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
	mux.Handle("/api/driver/v1/vehicles", protected(methodSwitch("POST", handler.CreateVehicleHandler(svcCtx))))
	mux.Handle("/api/driver/v1/vehicles/get", protected(methodSwitch("GET", handler.GetVehicleHandler(svcCtx))))
	mux.Handle("/api/driver/v1/withdraws", protected(methodSwitch("POST", handler.CreateWithdrawHandler(svcCtx))))
	mux.Handle("/api/driver/v1/withdraws/list", protected(methodSwitch("POST", handler.ListWithdrawsHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/accept", protected(methodSwitch("POST", handler.AcceptOrderHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/reject", protected(methodSwitch("POST", handler.RejectOrderHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/dispatches", protected(methodSwitch("POST", handler.ListMyDispatchesHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/list", protected(methodSwitch("POST", handler.ListMyOrdersHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/start-trip", protected(methodSwitch("POST", handler.StartTripHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/confirm-arrive", protected(methodSwitch("POST", handler.ConfirmArriveHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/finish-trip", protected(methodSwitch("POST", handler.FinishTripHandler(svcCtx))))
	mux.Handle("/api/driver/v1/ws", handler.DriverPushWSHandler(svcCtx))
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
