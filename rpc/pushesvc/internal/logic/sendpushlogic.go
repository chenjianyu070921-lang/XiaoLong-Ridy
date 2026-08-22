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

	// 调用真实推送通道（极光/个推/小米等，由 Push.Provider 决定）；
	// 失败自动重试一次，结果回写 push_log（result=1 成功 / 0 失败）
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		lastErr = l.svcCtx.PushProvider.Push(l.ctx, in.UserId, in.DeviceType, in.Title, in.Body, in.Extras)
		if lastErr == nil {
			break
		}
		l.Errorf("App推送通道调用失败(第%d次): userId=%d err=%v", attempt+1, in.UserId, lastErr)
	}
	result := int8(1)
	errMsg := ""
	if lastErr != nil {
		result = 0
		errMsg = lastErr.Error()
	}

	logEntry := &model.PushLog{
		UserID:   uint64(in.UserId),
		PushType: 1, // 1=App推送
		Title:    in.Title,
		Content:  in.Body,
		Target:   in.DeviceType,
		Extras:   in.Extras,
		Result:   result,
		ErrorMsg: errMsg,
	}
	if err := l.svcCtx.PushLogModel.Insert(logEntry); err != nil {
		l.Errorf("写入推送日志失败: %v", err)
		return nil, err
	}

	if lastErr != nil {
		l.Infof("App推送未完成(userId=%d)：%s", in.UserId, errMsg)
		return &pushesvc.SendPushResp{Success: false}, nil
	}
	l.Infof("App推送成功: userId=%d title=%s deviceType=%s", in.UserId, in.Title, in.DeviceType)
	return &pushesvc.SendPushResp{Success: true}, nil
}
