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
)

type TencentCloudConfig struct {
	SecretID  string
	SecretKey string
	Region    string
}

type TencentCloudRealNameVerifier struct {
	config TencentCloudConfig
	client *http.Client
}

func NewTencentCloudRealNameVerifier(cfg TencentCloudConfig) *TencentCloudRealNameVerifier {
	if strings.TrimSpace(cfg.SecretID) == "" || strings.TrimSpace(cfg.SecretKey) == "" {
		return nil
	}
	if cfg.Region == "" {
		cfg.Region = "ap-beijing"
	}
	return &TencentCloudRealNameVerifier{config: cfg, client: &http.Client{Timeout: 10 * time.Second}}
}

func (v *TencentCloudRealNameVerifier) Verify(ctx context.Context, name, idCardNo string) (*VerifyResult, error) {
	if v == nil || v.client == nil {
		return nil, fmt.Errorf("real-name verifier is not configured")
	}
	name = strings.TrimSpace(name)
	idCardNo = strings.TrimSpace(idCardNo)
	if name == "" || idCardNo == "" {
		return nil, fmt.Errorf("name and id card number are required")
	}
	authorization, date := calculateAuthorization(v.config.SecretID, v.config.SecretKey)
	body, err := json.Marshal(map[string]string{"cardNo": idCardNo, "realName": name})
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("https://%s.cloudmarket-apigw.com/service-18c38npd/idcard/VerifyIdcardv2", v.config.Region)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Date", date)

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("real-name provider returned status %d", resp.StatusCode)
	}
	return parseResponse(responseBody), nil
}

func calculateAuthorization(secretID, secretKey string) (string, string) {
	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	mac := hmac.New(sha1.New, []byte(secretKey))
	_, _ = mac.Write([]byte("x-date: " + date))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf(`{"id":"%s", "x-date":"%s", "signature":"%s"}`, secretID, date, signature), date
}

func parseResponse(responseBody []byte) *VerifyResult {
	result := &VerifyResult{Result: "-4", Description: "provider response could not be parsed"}
	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			IsMatch bool `json:"isMatch"`
		} `json:"data"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return result
	}
	if response.Code == 0 && response.Data.IsMatch {
		return &VerifyResult{Result: "0", Description: "name and identity-card number match"}
	}
	if response.Code == 0 {
		return &VerifyResult{Result: "-1", Description: "name and identity-card number do not match"}
	}
	return &VerifyResult{Result: fmt.Sprintf("%d", response.Code), Description: response.Message}
}
