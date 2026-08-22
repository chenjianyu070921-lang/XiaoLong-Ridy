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

	// 调用真实短信通道（阿里云/腾讯云等，由 SMS.Provider 决定）；
	// 失败自动重试一次，结果回写 push_log（result=1 成功 / 0 失败）
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		lastErr = l.svcCtx.SMSProvider.Send(l.ctx, in.Phone, in.Content)
		if lastErr == nil {
			break
		}
		l.Errorf("短信通道调用失败(第%d次): phone=%s err=%v", attempt+1, in.Phone, lastErr)
	}
	result := int8(1)
	errMsg := ""
	if lastErr != nil {
		result = 0
		errMsg = lastErr.Error()
	}

	logEntry := &model.PushLog{
		UserID:   0, // 短信按手机号发送，无用户ID
		PushType: 2, // 2=短信
		BizType:  int8(in.BizType),
		Content:  in.Content,
		Target:   in.Phone,
		Result:   result,
		ErrorMsg: errMsg,
	}
	if err := l.svcCtx.PushLogModel.Insert(logEntry); err != nil {
		l.Errorf("写入短信日志失败: %v", err)
		return nil, err
	}

	if lastErr != nil {
		l.Infof("短信未发送(phone=%s)：%s", in.Phone, errMsg)
		return &pushesvc.SendSMSResp{Success: false}, nil
	}
	l.Infof("短信发送成功: phone=%s bizType=%d", in.Phone, in.BizType)
	return &pushesvc.SendSMSResp{Success: true}, nil
}
