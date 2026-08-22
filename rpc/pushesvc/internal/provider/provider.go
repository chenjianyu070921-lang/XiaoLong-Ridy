package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/pushesvc/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
)

// SMSProvider 短信通道抽象：屏蔽阿里云/腾讯云等供应商差异
type SMSProvider interface {
	Send(ctx context.Context, phone, content string) error
}

// PushProvider App 推送通道抽象：屏蔽极光/个推/小米等供应商差异
type PushProvider interface {
	Push(ctx context.Context, userID int64, deviceType, title, body, extras string) error
}

// NewSMSProvider 按配置构建短信通道：未配置 Provider 时走本地 noop（开发/演示用）
func NewSMSProvider(cfg config.SMSConfig) SMSProvider {
	if cfg.Provider == "" || cfg.Provider == "mock" {
		return &noopSMS{}
	}
	return &httpSMS{cfg: cfg, client: &http.Client{Timeout: 5 * time.Second}}
}

// NewPushProvider 按配置构建推送通道：未配置 Provider 时走本地 noop（开发/演示用）
func NewPushProvider(cfg config.PushConfig) PushProvider {
	if cfg.Provider == "" || cfg.Provider == "mock" {
		return &noopPush{}
	}
	return &httpPush{cfg: cfg, client: &http.Client{Timeout: 5 * time.Second}}
}

// ---- noop 实现：未配置真实通道时模拟渠道成功，保证本地可跑通 ----
type noopSMS struct{}

func (n *noopSMS) Send(ctx context.Context, phone, content string) error {
	logx.Infof("[SMS-noop] 模拟短信发送 phone=%s content=%s", phone, content)
	return nil
}

type noopPush struct{}

func (n *noopPush) Push(ctx context.Context, userID int64, deviceType, title, body, extras string) error {
	logx.Infof("[Push-noop] 模拟App推送 userId=%d device=%s title=%s", userID, deviceType, title)
	return nil
}

// ---- HTTP 实现：占位演示用，向第三方网关发请求并解析响应体。 ----
// 注意：真实接入应替换为各厂商官方 SDK（阿里云/腾讯云/极光/个推等），
// 因为它们需要按签名算法生成 sign、且响应体字段各异。本实现仅做 HTTP 兜底判断，
// 用于暴露「HTTP 200 但业务失败」的假成功问题，不能用于生产环境。
var smsEndpoints = map[string]string{
	"aliyun":  "https://dysmsapi.aliyuncs.com",
	"tencent": "https://sms.tencentcloudapi.com",
}

var pushEndpoints = map[string]string{
	"jpush":  "https://api.jpush.cn/v3/push",
	"xiaomi": "https://api.xmpush.xiaomi.com/v3/message/regid",
	"getui":  "https://restapi.getui.com/v2/push",
}

type httpSMS struct {
	cfg    config.SMSConfig
	client *http.Client
}

func (h *httpSMS) Send(ctx context.Context, phone, content string) error {
	url, ok := smsEndpoints[h.cfg.Provider]
	if !ok {
		return fmt.Errorf("未支持的短信通道: %s", h.cfg.Provider)
	}
	body, _ := json.Marshal(map[string]any{
		"provider":  h.cfg.Provider,
		"accessKey": h.cfg.AccessKey,
		"signName":  h.cfg.SignName,
		"phone":     phone,
		"content":   content,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("短信网关请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("短信网关返回错误状态码: %d", resp.StatusCode)
	}
	// 读取并解析业务响应体：仅看 HTTP 状态码会漏判「HTTP 200 + 业务失败」的假成功。
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取短信网关响应失败: %w", err)
	}
	if ok, detail := parseGatewayResult(respBody); !ok {
		return fmt.Errorf("短信网关业务失败: %s", detail)
	}
	return nil
}

type httpPush struct {
	cfg    config.PushConfig
	client *http.Client
}

func (h *httpPush) Push(ctx context.Context, userID int64, deviceType, title, body, extras string) error {
	url, ok := pushEndpoints[h.cfg.Provider]
	if !ok {
		return fmt.Errorf("未支持的推送通道: %s", h.cfg.Provider)
	}
	raw, _ := json.Marshal(map[string]any{
		"appKey":       h.cfg.AppKey,
		"masterSecret": h.cfg.MasterSecret,
		"userId":       userID,
		"deviceType":   deviceType,
		"title":        title,
		"body":         body,
		"extras":       extras,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("推送网关请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("推送网关返回错误状态码: %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取推送网关响应失败: %w", err)
	}
	if ok, detail := parseGatewayResult(respBody); !ok {
		return fmt.Errorf("推送网关业务失败: %s", detail)
	}
	return nil
}

// parseGatewayResult 解析第三方网关的 JSON 响应，判断业务是否真正成功。
// 各厂商字段不同，这里按常见约定兜底：
//   - code/errcode/error_code/statusCode 存在且非 0/"OK" 视为失败
//   - success=false 视为失败
//   - message/msg 含 fail/error/invalid/denied 关键字视为失败
//
// 真实接入时应替换为各厂商官方 SDK（含签名与专有响应解析），本函数仅用于
// 兜底暴露「HTTP 200 + 业务失败」的假成功。
func parseGatewayResult(respBody []byte) (bool, string) {
	var m map[string]any
	if err := json.Unmarshal(respBody, &m); err != nil {
		// 非 JSON 响应：HTTP 状态码已通过，按成功处理（兼容纯文本网关）
		return true, ""
	}
	for _, key := range []string{"code", "errcode", "error_code", "statusCode"} {
		if v, ok := m[key]; ok && isNonZero(v) {
			return false, fmt.Sprintf("%s=%v", key, v)
		}
	}
	for _, key := range []string{"code", "status", "subCode"} {
		if v, ok := m[key]; ok {
			s := fmt.Sprintf("%v", v)
			if s != "" && s != "OK" && s != "ok" && s != "Success" && s != "success" && s != "0" {
				return false, fmt.Sprintf("%s=%s", key, s)
			}
		}
	}
	if v, ok := m["success"]; ok {
		if b, ok := v.(bool); ok && !b {
			return false, "success=false"
		}
	}
	for _, key := range []string{"message", "msg"} {
		if v, ok := m[key]; ok {
			s := strings.ToLower(fmt.Sprintf("%v", v))
			for _, kw := range []string{"fail", "error", "invalid", "denied"} {
				if strings.Contains(s, kw) {
					return false, fmt.Sprintf("%v", v)
				}
			}
		}
	}
	return true, ""
}

// isNonZero 判断一个值是否表示非零（失败）错误码。
func isNonZero(v any) bool {
	switch n := v.(type) {
	case float64:
		return n != 0
	case int:
		return n != 0
	case int64:
		return n != 0
	case string:
		return n != "" && n != "0" && n != "OK" && n != "ok"
	}
	return false
}
