package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// SendSMSCodeLogic 处理发送短信验证码 RPC 的业务流程。
type SendSMSCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewSendSMSCodeLogic 创建发送验证码逻辑实例，并绑定请求上下文与服务依赖。
func NewSendSMSCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSMSCodeLogic {
	return &SendSMSCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SendSMSCode 校验手机号后向短信服务发送验证码。
func (l *SendSMSCodeLogic) SendSMSCode(in *userproto.SendSMSCodeRequest) (*userproto.SendSMSCodeResponse, error) {
	phone := strings.TrimSpace(in.GetPhone())
	if !IsValidPhone(phone) {
		return nil, ErrInvalidPhone
	}

	// 发送能力由 ServiceContext 注入，便于本地内存实现和真实短信通道平滑切换。
	expireIn, err := l.sendSMSCode(phone)
	if err != nil {
		return nil, err
	}

	return &userproto.SendSMSCodeResponse{
		Success:  true,
		ExpireIn: expireIn,
	}, nil
}

// sendSMSCode 根据手机号是否已注册选择不同的短信频控策略。
func (l *SendSMSCodeLogic) sendSMSCode(phone string) (int64, error) {
	policy := registeredSMSRatePolicy()
	if isUnregisteredPhone(l.ctx, l.svcCtx, phone) {
		policy = unregisteredSMSRatePolicy()
	}
	if sender, ok := l.svcCtx.SMSSender.(SMSCodePolicySender); ok {
		return sender.SendWithPolicy(l.ctx, phone, policy)
	}
	return l.svcCtx.SMSSender.Send(l.ctx, phone)
}

// isUnregisteredPhone 判断手机号是否尚未注册，仓储不可用时按已注册基线处理。
func isUnregisteredPhone(ctx context.Context, svcCtx *svc.ServiceContext, phone string) bool {
	if svcCtx == nil || svcCtx.Users == nil {
		return false
	}
	_, err := svcCtx.Users.FindByPhone(ctx, phone)
	return errors.Is(err, repository.ErrUserNotFound)
}
