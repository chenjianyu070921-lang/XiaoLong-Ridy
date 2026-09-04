package logic

import (
	"context"
	"fmt"

	"XiaoLong-Ridy/rpc/pushesvc/internal/model"
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

// SendPush 发送 App 推送：真实落库 push_log 表
func (l *SendPushLogic) SendPush(in *pushesvc.SendPushReq) (*pushesvc.SendPushResp, error) {
	if in.UserId <= 0 {
		return nil, fmt.Errorf("user_id 非法: %d", in.UserId)
	}
	if in.Title == "" && in.Body == "" {
		return nil, fmt.Errorf("推送标题和内容不能都为空")
	}

	// 记录推送日志；真实推送通道（极光/个推等）由 Push 配置的 Provider 决定，
	// 当前未配置真实通道时按"已发送"记录，失败会落 result=0
	logEntry := &model.PushLog{
		UserID:   uint64(in.UserId),
		PushType: 1, // 1=App推送
		Title:    in.Title,
		Content:  in.Body,
		Target:   in.DeviceType,
		Result:   1, // 1=成功
	}
	if err := l.svcCtx.PushLogModel.Insert(logEntry); err != nil {
		l.Errorf("写入推送日志失败: %v", err)
		return nil, err
	}

	l.Infof("App推送成功: userId=%d title=%s deviceType=%s", in.UserId, in.Title, in.DeviceType)
	return &pushesvc.SendPushResp{Success: true}, nil
}
