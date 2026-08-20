// Package main 是司机端 HTTP API 服务的程序入口。
package main

import (
	"fmt"      // 用于格式化启动失败的错误信息
	"log"      // 用于输出服务启动日志
	"net/http" // 提供 HTTP 服务器与路由能力
	"os"       // 用于读取环境变量配置
	"time"     // 用于限流窗口配置

	"gopkg.in/yaml.v3" // 解析 etc/driver.yaml 配置（与其他 api 服务一致，从配置文件加载下游地址）

	"XiaoLong-Ridy/api/driver/internal/handler"    // 注册各业务域的 HTTP 路由处理器
	"XiaoLong-Ridy/api/driver/internal/middleware" // 提供 JWT 鉴权中间件
	"XiaoLong-Ridy/api/driver/internal/svc"        // 提供包含 driversvc 客户端的服务上下文
)

// defaultHTTPAddress 是 driver API 的默认 HTTP 监听地址（端口 8082）。
const defaultHTTPAddress = ":8082"

// defaultDriverGRPCAddr 是下游 driversvc gRPC 服务的默认地址（本地 5055）。
// 必须与 rpc/driversvc/etc/driversvc.yaml 的 ListenOn 保持一致，否则 driver-api 连不上 driversvc。
const defaultDriverGRPCAddr = "127.0.0.1:5055"

// defaultOrderGRPCAddr 是下游 ordersvc gRPC 服务的默认地址（本地 50051）。
const defaultOrderGRPCAddr = "127.0.0.1:50051"

const defaultDispatchGRPCAddr = "127.0.0.1:8083"

// driverConfig 是 driver.yaml 的最小配置结构，仅抽取启动所需字段（与 admin/passenger 从 etc 配置读取下游地址的方式对齐）。
type driverConfig struct {
	HTTPAddr         string `yaml:"httpAddr"`       // HTTP 监听地址
	DriverGRPCAddr   string `yaml:"driverGrpcAddr"` // 下游 driversvc 的 gRPC 地址
	DispatchGRPCAddr string `yaml:"dispatchGrpcAddr"`
}

// loadDriverConfig 从 etc/driver.yaml 读取配置；文件缺失或解析失败时返回零值，由调用方回退到默认常量。
func loadDriverConfig() driverConfig {
	var cfg driverConfig
	// 按约定从工作目录下的 etc/driver.yaml 读取（需在 api/driver 目录启动）。
	data, err := os.ReadFile("etc/driver.yaml")
	if err != nil {
		// 配置文件不存在（如其它启动方式），返回零值交由环境变量/默认常量兜底。
		return cfg
	}
	// 解析 YAML；失败同样返回零值，不阻断启动。
	_ = yaml.Unmarshal(data, &cfg)
	return cfg
}

// main 是程序入口：解析配置、构建服务上下文、启动 HTTP 服务。
func main() {
	// 从 etc/driver.yaml 读取基础配置（与其他 api 服务保持一致的配置来源）。
	fileCfg := loadDriverConfig()

	// HTTP 监听地址优先级：环境变量 DRIVER_HTTP_ADDR > driver.yaml httpAddr > 默认 :8082。
	address := os.Getenv("DRIVER_HTTP_ADDR")
	if address == "" {
		address = fileCfg.HTTPAddr
	}
	if address == "" {
		address = defaultHTTPAddress
	}

	// driversvc 的 gRPC 地址优先级：环境变量 DRIVER_GRPC_ADDR > driver.yaml driverGrpcAddr > 默认 127.0.0.1:5055。
	// 关键点：必须与 driversvc.yaml 的 ListenOn(5055) 一致，否则联调时 driver-api 找不到 driversvc。
	driverGRPCAddr := os.Getenv("DRIVER_GRPC_ADDR")
	if driverGRPCAddr == "" {
		driverGRPCAddr = fileCfg.DriverGRPCAddr
	}
	if driverGRPCAddr == "" {
		driverGRPCAddr = defaultDriverGRPCAddr
	}

	// 从环境变量读取 ordersvc 的 gRPC 地址，未配置时使用默认值。
	orderGRPCAddr := os.Getenv("ORDER_GRPC_ADDR")
	if orderGRPCAddr == "" {
		// 环境变量为空，回退到默认 gRPC 地址。
		orderGRPCAddr = defaultOrderGRPCAddr
	}
	dispatchGRPCAddr := os.Getenv("DISPATCH_GRPC_ADDR")
	if dispatchGRPCAddr == "" {
		dispatchGRPCAddr = fileCfg.DispatchGRPCAddr
	}
	if dispatchGRPCAddr == "" {
		dispatchGRPCAddr = defaultDispatchGRPCAddr
	}

	// 构造 HTTP 服务器，将路由处理器挂载到服务上下文之上。
	server := &http.Server{
		Addr:    address,                                                                                // 监听地址
		Handler: newHTTPHandler(svc.NewServiceContext(driverGRPCAddr, orderGRPCAddr, dispatchGRPCAddr)), // 注入持有下游客户端的上下文
	}

	// 输出启动日志，便于本地联调确认监听信息。
	log.Printf("driver api started at http://127.0.0.1%s  (driversvc gRPC: %s, ordersvc gRPC: %s)", address, driverGRPCAddr, orderGRPCAddr)
	// 启动 HTTP 服务并阻塞监听；仅在发生非预期错误时返回。
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// 启动失败直接 panic，由上层或进程管理器捕获。
		panic(fmt.Errorf("启动 driver api 失败: %w", err))
	}
}

// newHTTPHandler 构建并返回司机端全部 HTTP 路由的复用多路复用器（ServeMux）。
func newHTTPHandler(svcCtx *svc.ServiceContext) http.Handler {
	// 创建标准库的多路复用器，按路径分发请求到对应处理器。
	mux := http.NewServeMux()

	// 登录相关接口（无需鉴权，登录前可访问）。叠加登录/发码限流，防爆破与短信轰炸。
	loginRL := middleware.LoginRateLimit(10, time.Minute) // 单 IP 每分钟最多 10 次
	mux.Handle("/api/driver/v1/auth/send-sms-code", loginRL(methodSwitch("POST", handler.SendSMSCodeHandler(svcCtx))))
	mux.Handle("/api/driver/v1/auth/login-by-password", loginRL(methodSwitch("POST", handler.LoginByPasswordHandler(svcCtx))))
	mux.Handle("/api/driver/v1/auth/login-by-sms", loginRL(methodSwitch("POST", handler.LoginBySMSHandler(svcCtx))))

	// 受保护路由：先校验 HTTP 方法（methodSwitch 在外层，错误方法直接 405，
	// 避免被鉴权中间件抢先返回 401），再经过 JWT 鉴权中间件拦截无凭证请求。
	protected := middleware.RequireAuth(svcCtx)
	agentProtected := middleware.RequireDriverOrInternalService(svcCtx, os.Getenv("DRIVER_AGENT_SERVICE_TOKEN"))
	// 司机自注册：公开接口，无需登录（否则未注册用户永远无法创建账号）。
	// 仅限制为 POST 方法，不放进鉴权组。
	mux.Handle("/api/driver/v1/drivers/register", methodSwitch("POST", handler.RegisterDriverHandler(svcCtx)))
	// 管理员创建司机（后台建号）：需登录鉴权。
	mux.Handle("/api/driver/v1/drivers", methodSwitch("POST", protected(handler.CreateDriverHandler(svcCtx))))
	// 更新司机信息。
	mux.Handle("/api/driver/v1/drivers/update", protected(handler.UpdateDriverHandler(svcCtx)))
	// 查询司机详情（通过 ?id= 传参）。
	mux.Handle("/api/driver/v1/drivers/get", protected(handler.GetDriverHandler(svcCtx)))
	// 删除（软删）司机（通过 ?id= 传参）。
	mux.Handle("/api/driver/v1/drivers/delete", protected(handler.DeleteDriverHandler(svcCtx)))
	// 查询司机 AI 智能推荐得分（综合分 + 影响因子，含降级）。
	mux.Handle("/api/driver/v1/drivers/ai-score", protected(handler.GetDriverAiScoreHandler(svcCtx)))
	// 司机上线（置为在线状态）。
	mux.Handle("/api/driver/v1/drivers/online", methodSwitch("POST", protected(handler.SetOnlineHandler(svcCtx))))
	// 司机下线（置为离线状态）。
	mux.Handle("/api/driver/v1/drivers/offline", methodSwitch("POST", protected(handler.SetOfflineHandler(svcCtx))))
	// 司机心跳上报（刷新在线状态 + 多端互踢判定）。
	mux.Handle("/api/driver/v1/drivers/heartbeat", methodSwitch("POST", protected(handler.HeartbeatHandler(svcCtx))))
	// 司机资质上传（图片直传 MinIO 并落库）。
	mux.Handle("/api/driver/v1/drivers/certification/upload", methodSwitch("POST", protected(handler.UploadCertificationHandler(svcCtx))))
	// 司机资质查询。
	mux.Handle("/api/driver/v1/drivers/certification", methodSwitch("GET", protected(handler.GetCertificationHandler(svcCtx))))
	// 司机接单（driver_id 取自 JWT，order_id 取自请求体）。
	mux.Handle("/api/driver/v1/orders/accept", methodSwitch("POST", protected(handler.AcceptOrderHandler(svcCtx))))
	// 司机拒绝派单：当前等待 dispatchsvc.RejectDispatch。
	mux.Handle("/api/driver/v1/orders/reject", methodSwitch("POST", protected(handler.RejectOrderHandler(svcCtx))))
	// 我的派单列表：当前等待 dispatchsvc 按 driver_id 查询。
	mux.Handle("/api/driver/v1/orders/dispatches", methodSwitch("POST", protected(handler.ListMyDispatchesHandler(svcCtx))))
	// 司机开始行程。
	mux.Handle("/api/driver/v1/orders/start-trip", methodSwitch("POST", protected(handler.StartTripHandler(svcCtx))))
	// 司机确认到达上车点。
	mux.Handle("/api/driver/v1/orders/confirm-arrive", methodSwitch("POST", protected(handler.ConfirmArriveHandler(svcCtx))))
	// 司机结束行程并上报实际里程/时长/金额。
	mux.Handle("/api/driver/v1/orders/finish-trip", methodSwitch("POST", protected(handler.FinishTripHandler(svcCtx))))
	mux.Handle("/api/driver/v1/agent/chat", methodSwitch("POST", agentProtected(handler.AgentChatHandler())))

	// 将构建好的多路复用器返回给 HTTP 服务器使用。
	return mux
}

// methodSwitch 返回一个仅允许指定 HTTP 方法的包装处理器：方法不匹配时直接返回 405，
// 匹配则交给内部处理器。入参/出参均为 http.Handler，以便与鉴权中间件任意顺序组合嵌套
// （通常放在外层，让 405 先于 401 返回，避免被鉴权误拦）。
func methodSwitch(method string, h http.Handler) http.Handler {
	// 返回闭包处理器，在真正处理前先做方法校验。
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 校验请求方法是否为期望的方法。
		if r.Method != method {
			// 方法不匹配，返回 405 Method Not Allowed。
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// 方法匹配，交给真实处理器处理。
		h.ServeHTTP(w, r)
	})
}
