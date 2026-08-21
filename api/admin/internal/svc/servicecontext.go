package svc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"XiaoLong-Ridy/api/admin/internal/config"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 是管理后台服务的依赖容器。
// handler 和 logic 层通过它复用数据库连接与 adminsvc RPC 客户端。
// HTTP 层只负责参数解析、登录态透传和响应映射，业务读写统一下沉到 adminsvc。
type ServiceContext struct {
	Config         *config.Config
	MySQL          *sql.DB
	AdminSvc       adminclient.AdminService
	AdminRPCClient zrpc.Client
}

// NewServiceContext 初始化服务依赖。
// 这里会检查 MySQL 和 Redis 是否可用，避免服务启动后才暴露连接问题。
func NewServiceContext(cfg *config.Config) (*ServiceContext, error) {
	db, err := sql.Open("mysql", cfg.MySQL.DSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	adminRPCClient, adminSvc, err := newAdminRPCClient(cfg.AdminRPC)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &ServiceContext{
		Config:         cfg,
		MySQL:          db,
		AdminSvc:       adminSvc,
		AdminRPCClient: adminRPCClient,
	}, nil
}

// Close 关闭服务依赖。
// 服务退出时调用，避免连接资源泄漏。
func (s *ServiceContext) Close() {
	if s.MySQL != nil {
		_ = s.MySQL.Close()
	}
}

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
