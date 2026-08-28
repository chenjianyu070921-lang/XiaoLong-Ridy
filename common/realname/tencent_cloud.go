package realname

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TencentCloudConfig 保存腾讯云云市场实名认证配置。
type TencentCloudConfig struct {
	SecretID  string `yaml:"secretId" json:"secretId"`
	SecretKey string `yaml:"secretKey" json:"secretKey"`
	Region    string `yaml:"region" json:"region"`
}

// TencentCloudRealNameVerifier 是腾讯云实名认证客户端。
type TencentCloudRealNameVerifier struct {
	config TencentCloudConfig
	client *http.Client
}

// NewTencentCloudRealNameVerifier 创建客户端，密钥为空时返回 nil 以支持本地开发。
func NewTencentCloudRealNameVerifier(c TencentCloudConfig) *TencentCloudRealNameVerifier {
	if strings.TrimSpace(c.SecretID) == "" || strings.TrimSpace(c.SecretKey) == "" {
		return nil
	}
	if c.Region == "" {
		c.Region = "ap-beijing"
	}
	return &TencentCloudRealNameVerifier{c, &http.Client{Timeout: 10 * time.Second}}
}

// Verify 调用云市场二要素接口并转换响应。
func (v *TencentCloudRealNameVerifier) Verify(ctx context.Context, name, id string) (*VerifyResult, error) {
	if v == nil {
		return nil, fmt.Errorf("实名服务未初始化")
	}
	name, id = strings.TrimSpace(name), strings.TrimSpace(id)
	if name == "" || id == "" {
		return nil, fmt.Errorf("姓名和身份证号不能为空")
	}
	a, d := calcAuthorization(v.config.SecretID, v.config.SecretKey)
	u := fmt.Sprintf("https://%s.cloudmarket-apigw.com/service-8lifidsz/idcard/validate", v.config.Region)
	f := url.Values{"idcard_number": {id}, "name": {name}}
	q, e := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(f.Encode()))
	if e != nil {
		return nil, e
	}
	q.Header.Set("Authorization", a)
	q.Header.Set("X-Date", d)
	q.Header.Set("request-id", uuid.NewString())
	q.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r, e := v.client.Do(q)
	if e != nil {
		return nil, e
	}
	defer r.Body.Close()
	b, e := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if e != nil {
		return nil, e
	}
	if r.StatusCode < 200 || r.StatusCode >= 300 {
		return nil, fmt.Errorf("实名认证接口返回 HTTP %d: %s", r.StatusCode, strings.TrimSpace(string(b)))
	}
	return parseResponse(b), nil
}

// calcAuthorization 生成云市场 HMAC-SHA1 签名。
func calcAuthorization(id, key string) (string, string) {
	d := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	m := hmac.New(sha1.New, []byte(key))
	_, _ = m.Write([]byte("x-date: " + d))
	s := base64.StdEncoding.EncodeToString(m.Sum(nil))
	return fmt.Sprintf(`{"id":"%s", "x-date":"%s", "signature":"%s"}`, id, d, s), d
}

// parseResponse 兼容云市场常见响应格式。
func parseResponse(b []byte) *VerifyResult {
	r := &VerifyResult{Result: "-4", Description: strings.TrimSpace(string(b))}
	var x struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			IsMatch bool `json:"isMatch"`
		} `json:"data"`
		ErrorCode int    `json:"error_code"`
		Reason    string `json:"reason"`
		Legacy    struct {
			IsOK bool `json:"isok"`
		} `json:"result"`
	}
	if json.Unmarshal(b, &x) != nil {
		return r
	}
	if x.Code != 0 || x.Message != "" {
		// 部分云市场版本使用 code=200 表示 HTTP/业务成功，而不是 code=0。
		// 没有返回 isMatch 时，200 仍代表二要素核验通过。
		if (x.Code == 0 || x.Code == 200) && (x.Data.IsMatch || x.Code == 200) {
			r.Result, r.Description = "0", "姓名和身份证号一致"
		} else if x.Code == 0 {
			r.Result, r.Description = "-1", x.Message
		} else {
			r.Result, r.Description = fmt.Sprintf("%d", x.Code), x.Message
		}
		if r.Description == "" {
			r.Description = "实名认证未通过"
		}
		return r
	}
	if x.ErrorCode == 0 && x.Legacy.IsOK {
		r.Result, r.Description = "0", "姓名和身份证号一致"
	} else if x.Reason != "" {
		r.Result, r.Description = fmt.Sprintf("%d", x.ErrorCode), x.Reason
	}
	return r
}
