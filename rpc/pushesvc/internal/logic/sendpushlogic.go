package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/pushesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pushesvc/pushesvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendPushLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendPushLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendPushLogic {
	return &SendPushLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SendPush 发送 App 推送
func (l *SendPushLogic) SendPush(in *pushesvc.SendPushReq) (*pushesvc.SendPushResp, error) {
	l.Infof("SendPush: userId=%d, title=%s, deviceType=%s", in.UserId, in.Title, in.DeviceType)
	return &pushesvc.SendPushResp{
		Success: true,
	}, nil
}
