// Package sms 提供可复用的短信发送能力封装。
package sms

import (
	"context"
	"fmt"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20190711"
)

const (
	defaultTencentRegion  = "ap-guangzhou"
	defaultMainlandPrefix = "+86"
	tencentEndpoint       = "sms.tencentcloudapi.com"
	tencentSuccessCode    = "Ok"
)

// Sender 定义短信发送器的统一能力，便于 usersvc 或其他服务复用。
type Sender interface {
	Send(ctx context.Context, phone, code string) error
}

// TencentConfig 保存腾讯云短信发送器的必要参数。
type TencentConfig struct {
	SecretID    string
	SecretKey   string
	Region      string
	SmsSdkAppID string
	SignName    string
	TemplateID  string
}

// TencentSender 使用腾讯云短信 SDK 发送验证码短信。
type TencentSender struct {
	client      *tencentsms.Client
	smsSdkAppID string
	signName    string
	templateID  string
}

// NewTencentSender 创建腾讯云短信发送器。
func NewTencentSender(cfg TencentConfig) (*TencentSender, error) {
	cfg = normalizeTencentConfig(cfg)
	if err := validateTencentConfig(cfg); err != nil {
		return nil, err
	}

	credential := common.NewCredential(cfg.SecretID, cfg.SecretKey)
	clientProfile := profile.NewClientProfile()
	clientProfile.HttpProfile.Endpoint = tencentEndpoint

	client, err := tencentsms.NewClient(credential, cfg.Region, clientProfile)
	if err != nil {
		return nil, err
	}
	return &TencentSender{
		client:      client,
		smsSdkAppID: cfg.SmsSdkAppID,
		signName:    cfg.SignName,
		templateID:  cfg.TemplateID,
	}, nil
}

// Send 调用腾讯云短信接口发送验证码。
func (s *TencentSender) Send(ctx context.Context, phone, code string) error {
	_ = ctx
	if s == nil || s.client == nil {
		return fmt.Errorf("tencent sms sender is not initialized")
	}

	req := tencentsms.NewSendSmsRequest()
	req.SmsSdkAppid = common.StringPtr(s.smsSdkAppID)
	req.Sign = common.StringPtr(s.signName)
	req.TemplateID = common.StringPtr(s.templateID)
	req.PhoneNumberSet = []*string{common.StringPtr(formatPhone(phone))}
	req.TemplateParamSet = []*string{common.StringPtr(code)}

	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	if resp == nil || resp.Response == nil || len(resp.Response.SendStatusSet) == 0 {
		return fmt.Errorf("tencent sms returned empty send status")
	}
	status := resp.Response.SendStatusSet[0]
	if stringValue(status.Code) != tencentSuccessCode {
		return fmt.Errorf("tencent sms send failed: code=%s message=%s requestId=%s",
			stringValue(status.Code),
			stringValue(status.Message),
			stringValue(resp.Response.RequestId),
		)
	}
	return nil
}

// normalizeTencentConfig 补齐默认地域并清理配置空白字符。
func normalizeTencentConfig(cfg TencentConfig) TencentConfig {
	cfg.SecretID = strings.TrimSpace(cfg.SecretID)
	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)
	cfg.Region = strings.TrimSpace(cfg.Region)
	cfg.SmsSdkAppID = strings.TrimSpace(cfg.SmsSdkAppID)
	cfg.SignName = strings.TrimSpace(cfg.SignName)
	cfg.TemplateID = strings.TrimSpace(cfg.TemplateID)
	if cfg.Region == "" {
		cfg.Region = defaultTencentRegion
	}
	return cfg
}

// validateTencentConfig 校验腾讯云短信发送所需的最小配置集。
func validateTencentConfig(cfg TencentConfig) error {
	if cfg.SecretID == "" {
		return fmt.Errorf("tencent sms secret id is required")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("tencent sms secret key is required")
	}
	if cfg.SmsSdkAppID == "" {
		return fmt.Errorf("tencent sms sdk app id is required")
	}
	if cfg.SignName == "" {
		return fmt.Errorf("tencent sms sign name is required")
	}
	if cfg.TemplateID == "" {
		return fmt.Errorf("tencent sms template id is required")
	}
	return nil
}

// formatPhone 将手机号转换成腾讯云要求的 E.164 格式。
func formatPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if strings.HasPrefix(phone, "+") {
		return phone
	}
	return defaultMainlandPrefix + phone
}

// stringValue 读取 SDK 返回指针字段，避免空指针。
func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
