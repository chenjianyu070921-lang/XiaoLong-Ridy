package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/pushesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pushesvc/pushesvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendNoticeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendNoticeLogic {
	return &SendNoticeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SendNotice 发送站内信
func (l *SendNoticeLogic) SendNotice(in *pushesvc.SendNoticeReq) (*pushesvc.SendNoticeResp, error) {
	l.Infof("SendNotice: userId=%d, title=%s, bizType=%d", in.UserId, in.Title, in.BizType)
	return &pushesvc.SendNoticeResp{
		NoticeId: 1,
	}, nil
}
