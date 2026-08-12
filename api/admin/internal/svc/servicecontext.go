package svc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"XiaoLong-Ridy/api/admin/internal/config"
	"XiaoLong-Ridy/api/admin/internal/repository"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

// ServiceContext 是管理后台服务的依赖容器。
// handler 和 logic 层通过它复用数据库、缓存和各类仓储对象。
type ServiceContext struct {
	Config                 *config.Config
	MySQL                  *sql.DB
	Redis                  *redis.Client
	AdminRepository        *repository.AdminRepository
	SessionRepository      *repository.SessionRepository
	OperationLogRepository *repository.OperationLogRepository
	UserRepository         *repository.UserRepository
	DriverRepository       *repository.DriverRepository
	OrderRepository        *repository.OrderRepository
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

	return &ServiceContext{
		Config:                 cfg,
		MySQL:                  db,
		Redis:                  redisClient,
		AdminRepository:        adminRepo,
		SessionRepository:      sessionRepo,
		OperationLogRepository: repository.NewOperationLogRepository(db),
		UserRepository:         repository.NewUserRepository(db),
		DriverRepository:       repository.NewDriverRepository(db),
		OrderRepository:        repository.NewOrderRepository(db),
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
