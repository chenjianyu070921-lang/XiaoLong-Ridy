// Package main 是司机端 HTTP API 服务的程序入口。
package main

import (
	"fmt"  // 用于格式化启动失败的错误信息
	"log"  // 用于输出服务启动日志
	"net/http" // 提供 HTTP 服务器与路由能力
	"os"  // 用于读取环境变量配置

	"XiaoLong-Ridy/api/driver/internal/handler" // 注册各业务域的 HTTP 路由处理器
	"XiaoLong-Ridy/api/driver/internal/svc"      // 提供包含 driversvc 客户端的服务上下文
)

// defaultHTTPAddress 是 driver API 的默认 HTTP 监听地址（端口 8082）。
const defaultHTTPAddress = ":8082"

// defaultDriverGRPCAddr 是下游 driversvc gRPC 服务的默认地址（本地 8080）。
const defaultDriverGRPCAddr = "127.0.0.1:8080"

// main 是程序入口：解析配置、构建服务上下文、启动 HTTP 服务。
func main() {
	// 从环境变量读取 HTTP 监听地址，未配置时使用默认值。
	address := os.Getenv("DRIVER_HTTP_ADDR")
	if address == "" {
		// 环境变量为空，回退到默认地址。
		address = defaultHTTPAddress
	}
	// 从环境变量读取 driversvc 的 gRPC 地址，未配置时使用默认值。
	grpcAddr := os.Getenv("DRIVER_GRPC_ADDR")
	if grpcAddr == "" {
		// 环境变量为空，回退到默认 gRPC 地址。
		grpcAddr = defaultDriverGRPCAddr
	}

	// 构造 HTTP 服务器，将路由处理器挂载到服务上下文之上。
	server := &http.Server{
		Addr:    address,                              // 监听地址
		Handler: newHTTPHandler(svc.NewServiceContext(grpcAddr)), // 注入持有 driversvc 客户端的上下文
	}

	// 输出启动日志，便于本地联调确认监听信息。
	log.Printf("driver api started at http://127.0.0.1%s  (driversvc gRPC: %s)", address, grpcAddr)
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
	// 司机
	// 创建司机：仅允许 POST 方法。
	mux.HandleFunc("/api/driver/v1/drivers", methodSwitch("POST", handler.CreateDriverHandler(svcCtx)))
	// 更新司机信息。
	mux.HandleFunc("/api/driver/v1/drivers/update", handler.UpdateDriverHandler(svcCtx))
	// 查询司机详情（通过 ?id= 传参）。
	mux.HandleFunc("/api/driver/v1/drivers/get", handler.GetDriverHandler(svcCtx))
	// 删除（软删）司机（通过 ?id= 传参）。
	mux.HandleFunc("/api/driver/v1/drivers/delete", handler.DeleteDriverHandler(svcCtx))
	// 分页查询司机列表。
	mux.HandleFunc("/api/driver/v1/drivers/list", handler.ListDriversHandler(svcCtx))
	// 车辆
	// 创建车辆。
	mux.HandleFunc("/api/driver/v1/vehicles", handler.CreateVehicleHandler(svcCtx))
	// 更新车辆信息。
	mux.HandleFunc("/api/driver/v1/vehicles/update", handler.UpdateVehicleHandler(svcCtx))
	// 查询车辆详情。
	mux.HandleFunc("/api/driver/v1/vehicles/get", handler.GetVehicleHandler(svcCtx))
	// 删除车辆。
	mux.HandleFunc("/api/driver/v1/vehicles/delete", handler.DeleteVehicleHandler(svcCtx))
	// 分页查询车辆列表。
	mux.HandleFunc("/api/driver/v1/vehicles/list", handler.ListVehiclesHandler(svcCtx))
	// 认证
	// 创建司机认证资料。
	mux.HandleFunc("/api/driver/v1/certifications", handler.CreateCertificationHandler(svcCtx))
	// 更新认证资料（含审核状态流转）。
	mux.HandleFunc("/api/driver/v1/certifications/update", handler.UpdateCertificationHandler(svcCtx))
	// 查询认证详情。
	mux.HandleFunc("/api/driver/v1/certifications/get", handler.GetCertificationHandler(svcCtx))
	// 删除认证资料。
	mux.HandleFunc("/api/driver/v1/certifications/delete", handler.DeleteCertificationHandler(svcCtx))
	// 分页查询认证列表。
	mux.HandleFunc("/api/driver/v1/certifications/list", handler.ListCertificationsHandler(svcCtx))
	// 服务分
	// 创建司机服务分记录。
	mux.HandleFunc("/api/driver/v1/scores", handler.CreateScoreHandler(svcCtx))
	// 更新服务分记录。
	mux.HandleFunc("/api/driver/v1/scores/update", handler.UpdateScoreHandler(svcCtx))
	// 查询服务分详情。
	mux.HandleFunc("/api/driver/v1/scores/get", handler.GetScoreHandler(svcCtx))
	// 删除服务分记录。
	mux.HandleFunc("/api/driver/v1/scores/delete", handler.DeleteScoreHandler(svcCtx))
	// 分页查询服务分列表。
	mux.HandleFunc("/api/driver/v1/scores/list", handler.ListScoresHandler(svcCtx))
	// 提现
	// 创建提现申请。
	mux.HandleFunc("/api/driver/v1/withdraws", handler.CreateWithdrawHandler(svcCtx))
	// 更新提现记录（含打款状态流转）。
	mux.HandleFunc("/api/driver/v1/withdraws/update", handler.UpdateWithdrawHandler(svcCtx))
	// 查询提现详情。
	mux.HandleFunc("/api/driver/v1/withdraws/get", handler.GetWithdrawHandler(svcCtx))
	// 删除提现记录。
	mux.HandleFunc("/api/driver/v1/withdraws/delete", handler.DeleteWithdrawHandler(svcCtx))
	// 分页查询提现列表。
	mux.HandleFunc("/api/driver/v1/withdraws/list", handler.ListWithdrawsHandler(svcCtx))
	// 将构建好的多路复用器返回给 HTTP 服务器使用。
	return mux
}

// methodSwitch 返回一个仅允许指定 HTTP 方法的包装处理器；方法不匹配时返回 405。
func methodSwitch(method string, h http.HandlerFunc) http.HandlerFunc {
	// 返回闭包处理器，在真正处理前先做方法校验。
	return func(w http.ResponseWriter, r *http.Request) {
		// 校验请求方法是否为期望的方法。
		if r.Method != method {
			// 方法不匹配，返回 405 Method Not Allowed。
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// 方法匹配，交给真实处理器处理。
		h(w, r)
	}
}
