package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/pushesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pushesvc/pushesvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendSMSLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendSMSLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSMSLogic {
	return &SendSMSLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SendSMS 发送短信
func (l *SendSMSLogic) SendSMS(in *pushesvc.SendSMSReq) (*pushesvc.SendSMSResp, error) {
	l.Infof("SendSMS: phone=%s, bizType=%d", in.Phone, in.BizType)
	return &pushesvc.SendSMSResp{
		Success: true,
	}, nil
}
