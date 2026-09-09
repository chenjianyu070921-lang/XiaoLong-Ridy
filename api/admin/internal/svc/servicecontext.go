package svc

import (
	"fmt"

	"XiaoLong-Ridy/api/admin/internal/config"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"

	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 是管理后台服务的依赖容器。
// handler 和 logic 层通过它复用 adminsvc RPC 客户端。
// HTTP 层只负责参数解析、登录态透传和响应映射，业务读写统一下沉到 adminsvc。
type ServiceContext struct {
	Config         *config.Config
	AdminSvc       adminclient.AdminService
	AdminRPCClient zrpc.Client
}

// NewServiceContext 初始化网关依赖。
// 网关不持有业务数据库连接，避免绕过 adminsvc 的服务边界。
func NewServiceContext(cfg *config.Config) (*ServiceContext, error) {
	adminRPCClient, adminSvc, err := newAdminRPCClient(cfg.AdminRPC)
	if err != nil {
		return nil, err
	}

	return &ServiceContext{
		Config:         cfg,
		AdminSvc:       adminSvc,
		AdminRPCClient: adminRPCClient,
	}, nil
}

// Close 预留服务退出资源释放入口。
// adminsvc RPC 客户端由 go-zero 生命周期管理，网关不再持有数据库连接。
func (s *ServiceContext) Close() {}

// newAdminRPCClient 按配置初始化 adminsvc 的 gRPC 客户端。
// HTTP 网关只负责鉴权和参数转换，真正的业务操作通过该客户端下沉到 adminsvc。
func newAdminRPCClient(cfg zrpc.RpcClientConf) (zrpc.Client, adminclient.AdminService, error) {
	if len(cfg.Endpoints) == 0 && cfg.Target == "" {
		cfg.Target = "127.0.0.1:8084"
	}
	client, err := zrpc.NewClient(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("new admin rpc client: %w", err)
	}
	return client, adminclient.NewAdminService(client), nil
}
