package realname

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// TencentCloudConfig 表示腾讯云云市场实名认证服务配置。
type TencentCloudConfig struct {
	SecretID  string // 云市场 SecretId
	SecretKey string // 云市场 SecretKey
	Region    string // 区域，默认 ap-beijing
}

// TencentCloudRealNameVerifier 表示腾讯云云市场实名认证实现。
type TencentCloudRealNameVerifier struct {
	config TencentCloudConfig
	client *http.Client
}

// NewTencentCloudRealNameVerifier 根据配置创建腾讯云实名认证客户端。
//
// 当 SecretID 或 SecretKey 为空时返回 nil，表示本地开发环境跳过实名认证。
func NewTencentCloudRealNameVerifier(cfg TencentCloudConfig) *TencentCloudRealNameVerifier {
	if cfg.SecretID == "" || cfg.SecretKey == "" {
		return nil
	}

	if cfg.Region == "" {
		cfg.Region = "ap-beijing"
	}

	return &TencentCloudRealNameVerifier{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Verify 调用腾讯云云市场身份证实名认证接口。
//
// 该接口的请求体应以 application/x-www-form-urlencoded 提交；
// JSON 请求容易导致网关拿不到参数，从而返回“无效请求”。
func (v *TencentCloudRealNameVerifier) Verify(ctx context.Context, name, idCardNo string) (*VerifyResult, error) {
	auth, datetime, err := calcAuthorization(v.config.SecretID, v.config.SecretKey)
	if err != nil {
		logx.WithContext(ctx).Errorf("计算实名认证签名失败: %v", err)
		return nil, fmt.Errorf("计算实名认证签名失败: %w", err)
	}

	reqURL := fmt.Sprintf("https://%s.cloudmarket-apigw.com/service-18c38npd/idcard/VerifyIdcardv2", v.config.Region)

	form := url.Values{}
	form.Set("cardNo", idCardNo)
	form.Set("realName", name)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		logx.WithContext(ctx).Errorf("创建实名认证请求失败: %v", err)
		return nil, fmt.Errorf("创建实名认证请求失败: %w", err)
	}

	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Date", datetime)
	req.Header.Set("request-id", fmt.Sprintf("realname-%d", time.Now().UnixNano()))
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := v.client.Do(req)
	if err != nil {
		logx.WithContext(ctx).Errorf("调用实名认证API失败: %v", err)
		return nil, fmt.Errorf("调用实名认证API失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logx.WithContext(ctx).Errorf("读取实名认证响应失败: %v", err)
		return nil, fmt.Errorf("读取实名认证响应失败: %w", err)
	}

	logx.WithContext(ctx).Infof("实名认证API响应: status=%d body=%s", resp.StatusCode, string(respBody))

	return parseResponse(string(respBody)), nil
}

// calcAuthorization 根据云市场要求拼接 HMAC-SHA1 签名。
func calcAuthorization(secretID, secretKey string) (auth string, datetime string, err error) {
	timeLocation, _ := time.LoadLocation("Etc/GMT")
	datetime = time.Now().In(timeLocation).Format("Mon, 02 Jan 2006 15:04:05 GMT")

	signStr := fmt.Sprintf("x-date: %s", datetime)

	mac := hmac.New(sha1.New, []byte(secretKey))
	mac.Write([]byte(signStr))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	auth = fmt.Sprintf(`{"id":"%s", "x-date":"%s", "signature":"%s"}`, secretID, datetime, sign)
	return auth, datetime, nil
}

// parseResponse 将云市场返回转换成统一的 VerifyResult。
//
// 兼容两类常见响应：
// 1. code/message/data.isMatch
// 2. error_code/reason/result.isok
func parseResponse(respBody string) *VerifyResult {
	result := &VerifyResult{
		Result:      "-4",
		Description: respBody,
	}

	var marketResp struct {
		ErrorCode int    `json:"error_code"`
		Reason    string `json:"reason"`
		Result    struct {
			IsOK bool `json:"isok"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(respBody), &marketResp); err == nil && (marketResp.ErrorCode != 0 || marketResp.Reason != "" || marketResp.Result.IsOK) {
		if marketResp.ErrorCode == 0 && marketResp.Result.IsOK {
			result.Result = "0"
			result.Description = "姓名和身份证号一致"
		} else if marketResp.ErrorCode == 0 && !marketResp.Result.IsOK {
			result.Result = "-1"
			result.Description = "姓名和身份证号不一致"
		} else {
			result.Result = fmt.Sprintf("%d", marketResp.ErrorCode)
			result.Description = marketResp.Reason
		}
		return result
	}

	var apiResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			IsMatch bool `json:"isMatch"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(respBody), &apiResp); err != nil {
		result.Description = fmt.Sprintf("响应解析失败: %s", respBody)
		return result
	}

	if apiResp.Code == 0 && apiResp.Data.IsMatch {
		result.Result = "0"
		result.Description = "姓名和身份证号一致"
	} else if apiResp.Code == 0 && !apiResp.Data.IsMatch {
		result.Result = "-1"
		result.Description = "姓名和身份证号不一致"
	} else {
		result.Result = fmt.Sprintf("%d", apiResp.Code)
		result.Description = apiResp.Message
	}

	return result
}
