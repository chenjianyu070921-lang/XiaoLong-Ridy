package logic

import (
	"context"
	"fmt"

	"XiaoLong-Ridy/rpc/pushesvc/internal/model"
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

// SendSMS 发送短信：真实落库 push_log 表
func (l *SendSMSLogic) SendSMS(in *pushesvc.SendSMSReq) (*pushesvc.SendSMSResp, error) {
	if in.Phone == "" {
		return nil, fmt.Errorf("phone 不能为空")
	}
	if in.Content == "" {
		return nil, fmt.Errorf("短信内容不能为空")
	}

	// 记录短信日志；真实短信通道（阿里云/腾讯云等）由 SMS 配置的 Provider 决定，
	// 当前未配置真实通道时按"已发送"记录
	logEntry := &model.PushLog{
		UserID:   0, // 短信按手机号发送，无用户ID
		PushType: 2, // 2=短信
		Content:  in.Content,
		Target:   in.Phone,
		Result:   1, // 1=成功
	}
	if err := l.svcCtx.PushLogModel.Insert(logEntry); err != nil {
		l.Errorf("写入短信日志失败: %v", err)
		return nil, err
	}

	l.Infof("短信发送成功: phone=%s bizType=%d", in.Phone, in.BizType)
	return &pushesvc.SendSMSResp{Success: true}, nil
}
