package logic

import (
	"context"
	"fmt"

	"XiaoLong-Ridy/rpc/pushesvc/internal/model"
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

// SendNotice 发送站内信：真实落库 notices 表
func (l *SendNoticeLogic) SendNotice(in *pushesvc.SendNoticeReq) (*pushesvc.SendNoticeResp, error) {
	if in.UserId <= 0 {
		return nil, fmt.Errorf("user_id 非法: %d", in.UserId)
	}
	if in.Title == "" || in.Content == "" {
		return nil, fmt.Errorf("站内信标题和内容不能为空")
	}

	n := &model.Notice{
		UserID:  uint64(in.UserId),
		Title:   in.Title,
		Content: in.Content,
		BizType: int8(in.BizType),
	}
	if err := l.svcCtx.NoticeModel.Insert(n); err != nil {
		l.Errorf("写入站内信失败: %v", err)
		return nil, err
	}

	l.Infof("发送站内信成功: noticeId=%d userId=%d title=%s", n.ID, in.UserId, in.Title)
	return &pushesvc.SendNoticeResp{NoticeId: int64(n.ID)}, nil
}
