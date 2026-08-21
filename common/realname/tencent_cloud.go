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
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// TencentCloudConfig 腾讯云市场实名认证服务配置。
type TencentCloudConfig struct {
	SecretID  string // 云市场分配的密钥 Id
	SecretKey string // 云市场分配的密钥 Key
	Region    string // 地域，默认 ap-beijing
}

// TencentCloudRealNameVerifier 腾讯云市场身份证实名认证实现。
type TencentCloudRealNameVerifier struct {
	config TencentCloudConfig
	client *http.Client
}

// NewTencentCloudRealNameVerifier 根据配置创建腾讯云市场实名认证客户端实例。
//
// 若 SecretID 或 SecretKey 为空，返回 nil 表示未启用实名认证（兼容本地开发环境）。
func NewTencentCloudRealNameVerifier(cfg TencentCloudConfig) *TencentCloudRealNameVerifier {
	if cfg.SecretID == "" || cfg.SecretKey == "" {
		return nil
	}

	if cfg.Region == "" {
		cfg.Region = "ap-beijing"
	}

	return &TencentCloudRealNameVerifier{
		config: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Verify 调用腾讯云市场身份证实名认证接口校验姓名和身份证号的真实性和一致性。
//
// 接口地址：https://ap-beijing.cloudmarket-apigw.com/service-18c38npd/idcard/VerifyIdcardv2
//
// 返回结果说明：
//   - Result = "0": 认证通过 ✅
//   - Result = "-1": 认证失败 ❌（姓名与身份证号不匹配）
//   - 其他值：系统异常或参数错误
func (v *TencentCloudRealNameVerifier) Verify(ctx context.Context, name, idCardNo string) (*VerifyResult, error) {
	// 生成签名
	auth, datetime, err := calcAuthorization(v.config.SecretID, v.config.SecretKey)
	if err != nil {
		logx.WithContext(ctx).Errorf("生成签名失败: %v", err)
		return nil, fmt.Errorf("生成签名失败: %w", err)
	}

	// 构建请求
	reqURL := fmt.Sprintf("https://%s.cloudmarket-apigw.com/service-18c38npd/idcard/VerifyIdcardv2", v.config.Region)

	// 构建JSON请求体
	reqBody := map[string]string{
		"cardNo":   idCardNo,
		"realName": name,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		logx.WithContext(ctx).Errorf("构建请求体失败: %v", err)
		return nil, fmt.Errorf("构建请求体失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		logx.WithContext(ctx).Errorf("创建请求失败: %v", err)
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Date", datetime)

	// 发送请求
	resp, err := v.client.Do(req)
	if err != nil {
		logx.WithContext(ctx).Errorf("调用实名认证API失败: %v", err)
		return nil, fmt.Errorf("调用实名认证API失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logx.WithContext(ctx).Errorf("读取响应失败: %v", err)
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	logx.WithContext(ctx).Infof("实名认证API响应: status=%d body=%s", resp.StatusCode, string(respBody))

	// 解析响应并返回标准格式
	result := parseResponse(string(respBody))

	return result, nil
}

// calcAuthorization 计算 HMAC-SHA1 签名。
func calcAuthorization(secretID, secretKey string) (auth string, datetime string, err error) {
	timeLocation, _ := time.LoadLocation("Etc/GMT")
	datetime = time.Now().In(timeLocation).Format("Mon, 02 Jan 2006 15:04:05 GMT")

	signStr := fmt.Sprintf("x-date: %s", datetime)

	mac := hmac.New(sha1.New, []byte(secretKey))
	mac.Write([]byte(signStr))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	auth = fmt.Sprintf(`{"id":"%s", "x-date":"%s", "signature":"%s"}`,
		secretID, datetime, sign)

	return auth, datetime, nil
}

// parseResponse 解析API响应并转换为标准结果格式。
//
// 注意：此函数需根据实际API返回的JSON格式进行调整。
// 常见返回格式示例：
//
// 成功：{"code":0,"message":"success","data":{"isMatch":true}}
// 失败：{"code":-1,"message":"姓名与身份证号不匹配"}
func parseResponse(respBody string) *VerifyResult {
	result := &VerifyResult{
		Result:      "-4",
		Description: respBody,
	}

	// 尝试解析 JSON 响应
	var apiResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			IsMatch bool `json:"isMatch"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(respBody), &apiResp); err != nil {
		// JSON 解析失败，返回原始响应
		result.Description = fmt.Sprintf("响应解析失败: %s", respBody)
		return result
	}

	// 根据业务码判断结果
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
