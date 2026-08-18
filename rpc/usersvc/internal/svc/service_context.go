package svc

import (
	"context"

	"XiaoLong-Ridy/rpc/usersvc/internal/config"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
)

// SMSCodeSender 定义验证码发送能力，便于后续替换为短信服务或推送服务 RPC。
type SMSCodeSender interface {
	Send(ctx context.Context, phone string) (expireIn int64, err error)
}

// SMSCodeVerifier 定义验证码校验能力，业务逻辑只依赖抽象契约。
type SMSCodeVerifier interface {
	Verify(ctx context.Context, phone, code string) (bool, error)
}

// TokenManager 定义登录令牌管理能力，统一封装签发、刷新和注销逻辑。
type TokenManager interface {
	Issue(userID uint64, phone string, userStatus int) (token string, refreshToken string, err error)
	Refresh(refreshToken string) (token string, newRefreshToken string, err error)
	Revoke(token string) error
}

// ServiceContext 集中保存 usersvc 运行时依赖，server 层只负责转发请求。
type ServiceContext struct {
	Config      config.Config
	Users       repository.UserRepository
	Addresses   repository.AddressRepository
	Coupons     repository.CouponRepository
	SMSSender   SMSCodeSender
	SMSVerifier SMSCodeVerifier
	Tokens      TokenManager
}

// NewServiceContext 按 goctl 风格根据配置和依赖创建 usersvc 服务上下文。
func NewServiceContext(
	c config.Config,
	users repository.UserRepository,
	addresses repository.AddressRepository,
	coupons repository.CouponRepository,
	smsSender SMSCodeSender,
	smsVerifier SMSCodeVerifier,
	tokens TokenManager,
) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		Users:       users,
		Addresses:   addresses,
		Coupons:     coupons,
		SMSSender:   smsSender,
		SMSVerifier: smsVerifier,
		Tokens:      tokens,
	}
}
