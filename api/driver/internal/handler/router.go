package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
)

// NewRouter 创建司机端 HTTP 路由入口，统一登记全部 API 路由。
// 形态与 passenger/admin 的 NewRouter(svcCtx) http.Handler 保持一致：内部创建 mux、
// 登记路由、返回 http.Handler；全局中间件（recover/CORS/InternalServiceAuth）由 main.go 包裹。
// 逐路由中间件（登录限流、RequireAuth、agent 双鉴权）在此包裹。
func NewRouter(svcCtx *svc.ServiceContext) http.Handler {
	mux := http.NewServeMux()

	// 资质文件静态服务（非 API 路由，单独登记在 mux 上）。
	mux.Handle("/api/driver/v1/certification-files/", http.StripPrefix(
		"/api/driver/v1/certification-files/",
		http.FileServer(http.Dir(certificationLocalDir())),
	))

	// 登录/注册/发码公开接口：接入 IP 级限流，防止验证码刷发与密码/验证码爆破。
	// send-sms-code 限流更严（5 次/分钟/IP）；登录与注册 10 次/分钟/IP。
	smsLimit := middleware.LoginRateLimit(5, time.Minute)
	authLimit := middleware.LoginRateLimit(10, time.Minute)
	mux.Handle("/api/driver/v1/auth/send-sms-code", smsLimit(methodSwitch("POST", SendSMSCodeHandler(svcCtx))))
	mux.Handle("/api/driver/v1/auth/login-by-password", authLimit(methodSwitch("POST", LoginByPasswordHandler(svcCtx))))
	mux.Handle("/api/driver/v1/auth/login-by-sms", authLimit(methodSwitch("POST", LoginBySMSHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/register", authLimit(methodSwitch("POST", RegisterDriverHandler(svcCtx))))
	protected := middleware.RequireAuth(svcCtx)
	mux.Handle("/api/driver/v1/upload/avatar-token", protected(methodSwitch("POST", AvatarUploadTokenHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/update", protected(methodSwitch("POST", UpdateDriverHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/get", protected(methodSwitch("GET", GetDriverHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/online", protected(methodSwitch("POST", SetOnlineHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/offline", protected(methodSwitch("POST", SetOfflineHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/heartbeat", protected(methodSwitch("POST", HeartbeatHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/location/report", protected(methodSwitch("POST", ReportLocationHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/ai-score", protected(methodSwitch("GET", GetDriverAiScoreHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/ai-score/refresh", protected(methodSwitch("POST", RefreshDriverScoreHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/certification/upload", protected(methodSwitch("POST", UploadCertificationHandler(svcCtx))))
	mux.Handle("/api/driver/v1/drivers/certification", protected(methodSwitch("GET", GetCertificationHandler(svcCtx))))
	mux.Handle("/api/driver/v1/vehicles", protected(methodSwitch("POST", CreateVehicleHandler(svcCtx))))
	mux.Handle("/api/driver/v1/vehicles/get", protected(methodSwitch("GET", GetVehicleHandler(svcCtx))))
	mux.Handle("/api/driver/v1/vehicles/update", protected(methodSwitch("POST", UpdateVehicleHandler(svcCtx))))
	mux.Handle("/api/driver/v1/vehicles/delete", protected(methodSwitch("POST", DeleteVehicleHandler(svcCtx))))
	mux.Handle("/api/driver/v1/vehicles/list", protected(methodSwitch("GET", ListVehiclesHandler(svcCtx))))
	mux.Handle("/api/driver/v1/withdraws", protected(methodSwitch("POST", CreateWithdrawHandler(svcCtx))))
	mux.Handle("/api/driver/v1/withdraws/list", protected(methodSwitch("POST", ListWithdrawsHandler(svcCtx))))
	mux.Handle("/api/driver/v1/income/summary", protected(methodSwitch("GET", GetIncomeSummaryHandler(svcCtx))))
	mux.Handle("/api/driver/v1/income/today", protected(methodSwitch("GET", GetTodayIncomeHandler(svcCtx))))
	mux.Handle("/api/driver/v1/income/week", protected(methodSwitch("GET", GetWeekIncomeHandler(svcCtx))))
	mux.Handle("/api/driver/v1/income/bills", protected(methodSwitch("POST", ListIncomeBillsHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/accept", protected(methodSwitch("POST", AcceptOrderHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/reject", protected(methodSwitch("POST", RejectOrderHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/dispatches", protected(methodSwitch("POST", ListMyDispatchesHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/available", protected(methodSwitch("POST", ListAvailableOrdersHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/grab-list", protected(methodSwitch("POST", ListGrabOrdersHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/heatmap", protected(methodSwitch("POST", GetOrderHeatmapHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/list", protected(methodSwitch("POST", ListMyOrdersHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/detail", protected(methodSwitch("POST", GetMyOrderDetailHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/trajectory", protected(methodSwitch("POST", GetOrderTrajectoryHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/start-trip", protected(methodSwitch("POST", StartTripHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/confirm-arrive", protected(methodSwitch("POST", ConfirmArriveHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/realtime-fare", protected(methodSwitch("POST", GetRealtimeFareHandler(svcCtx))))
	mux.Handle("/api/driver/v1/orders/finish-trip", protected(methodSwitch("POST", FinishTripHandler(svcCtx))))
	mux.Handle("/api/driver/v1/ws", DriverPushWSHandler(svcCtx))
	mux.Handle("/api/driver/v1/reviews/received", protected(methodSwitch("GET", ListReceivedReviewsHandler(svcCtx))))
	mux.Handle("/api/driver/v1/reviews/summary", protected(methodSwitch("GET", ReviewSummaryHandler(svcCtx))))
	mux.Handle("/api/driver/v1/bank-cards/sms-code", protected(methodSwitch("POST", SendBankCardSmsCodeHandler(svcCtx))))
	mux.Handle("/api/driver/v1/bank-cards", protected(methodSwitch("POST", BindBankCardHandler(svcCtx))))
	mux.Handle("/api/driver/v1/bank-cards/list", protected(methodSwitch("GET", ListBankCardsHandler(svcCtx))))
	mux.Handle("/api/driver/v1/bank-cards/delete", protected(methodSwitch("POST", DeleteBankCardHandler(svcCtx))))
	mux.Handle("/api/driver/v1/bank-cards/reset-password", protected(methodSwitch("POST", ResetWithdrawPasswordHandler(svcCtx))))
	mux.Handle("/api/driver/v1/agent/chat", internalOrDriverAuth(svcCtx, methodSwitch("POST", AgentChatHandler())))
	return mux
}

// methodSwitch 限定路由仅接受指定 HTTP 方法，其余返回 405。
func methodSwitch(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}

// internalOrDriverAuth 允许携带内部服务令牌直接访问（agent 服务互信），否则回落司机登录鉴权。
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

// certificationLocalDir 返回资质文件本地存储目录（与 goctl 生成解耦，保留在 handler 包）。
func certificationLocalDir() string {
	if dir := strings.TrimSpace(os.Getenv("DRIVER_CERT_LOCAL_DIR")); dir != "" {
		return dir
	}
	return filepath.Join(".run", "certifications")
}
