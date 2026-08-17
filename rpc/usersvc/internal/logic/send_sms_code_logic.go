package logic

import (
	"context"
	"strings"

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
	expireIn, err := l.svcCtx.SMSSender.Send(l.ctx, phone)
	if err != nil {
		return nil, err
	}

	return &userproto.SendSMSCodeResponse{
		Success:  true,
		ExpireIn: expireIn,
	}, nil
}
