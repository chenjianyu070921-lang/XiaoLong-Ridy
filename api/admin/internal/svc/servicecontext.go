package svc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"XiaoLong-Ridy/api/admin/internal/config"
	"XiaoLong-Ridy/api/admin/internal/repository"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
	payclient "XiaoLong-Ridy/rpc/paysvc/pay"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 是管理后台服务的依赖容器。
// handler 和 logic 层通过它复用数据库、缓存和各类仓储对象。
type ServiceContext struct {
	Config                 *config.Config
	MySQL                  *sql.DB
	Redis                  *redis.Client
	AdminSvc               adminclient.AdminService
	AdminRPCClient         zrpc.Client
	PaySvc                 payclient.Pay
	PayRPCClient           zrpc.Client
	AdminRepository        *repository.AdminRepository
	SessionRepository      *repository.SessionRepository
	OperationLogRepository *repository.OperationLogRepository
	UserRepository         *repository.UserRepository
	DriverRepository       *repository.DriverRepository
	OrderRepository        *repository.OrderRepository
	CouponRepository       *repository.CouponRepository
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

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		_ = db.Close()
		_ = redisClient.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	sessionTTL := time.Duration(cfg.Auth.SessionTTLHours) * time.Hour
	adminRepo := repository.NewAdminRepository(db)
	sessionRepo := repository.NewSessionRepository(redisClient, cfg.Auth.TokenPrefix, sessionTTL)
	adminRPCClient, adminSvc, err := newAdminRPCClient(cfg.AdminRPC)
	if err != nil {
		_ = db.Close()
		_ = redisClient.Close()
		return nil, err
	}

	payRPCClient, paySvc, err := newPayRPCClient(cfg.PayRPC)
	if err != nil {
		_ = db.Close()
		_ = redisClient.Close()
		return nil, err
	}

	return &ServiceContext{
		Config:                 cfg,
		MySQL:                  db,
		Redis:                  redisClient,
		AdminSvc:               adminSvc,
		AdminRPCClient:         adminRPCClient,
		PaySvc:                 paySvc,
		PayRPCClient:           payRPCClient,
		AdminRepository:        adminRepo,
		SessionRepository:      sessionRepo,
		OperationLogRepository: repository.NewOperationLogRepository(db),
		UserRepository:         repository.NewUserRepository(db),
		DriverRepository:       repository.NewDriverRepository(db),
		OrderRepository:        repository.NewOrderRepository(db),
		CouponRepository:       repository.NewCouponRepository(db),
	}, nil
}

// Close 关闭服务依赖。
// 服务退出时调用，避免连接资源泄漏。
func (s *ServiceContext) Close() {
	if s.Redis != nil {
		_ = s.Redis.Close()
	}
	if s.MySQL != nil {
		_ = s.MySQL.Close()
	}
}

// newAdminRPCClient 按配置初始化 adminsvc 的 gRPC 客户端。
// HTTP 网关只负责鉴权和参数转换，真正的业务操作通过该客户端下沉到 adminsvc。
func newAdminRPCClient(cfg zrpc.RpcClientConf) (zrpc.Client, adminclient.AdminService, error) {
	if len(cfg.Endpoints) == 0 && cfg.Target == "" {
		cfg.Target = "127.0.0.1:8080"
	}
	client, err := zrpc.NewClient(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("new admin rpc client: %w", err)
	}
	return client, adminclient.NewAdminService(client), nil
}

// newPayRPCClient 按配置初始化 paysvc 的 gRPC 客户端。
// 网关层支付回调（/api/pay/callback/alipay）通过它透传到 paysvc.NotifyPayment。
func newPayRPCClient(cfg zrpc.RpcClientConf) (zrpc.Client, payclient.Pay, error) {
	if len(cfg.Endpoints) == 0 && cfg.Target == "" {
		cfg.Target = "127.0.0.1:50054"
	}
	client, err := zrpc.NewClient(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("new pay rpc client: %w", err)
	}
	return client, payclient.NewPay(client), nil
}
