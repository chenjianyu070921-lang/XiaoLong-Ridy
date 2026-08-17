package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"XiaoLong-Ridy/api/passenger/internal/router"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	"XiaoLong-Ridy/rpc/usersvc/client"
)

// TestAuthHTTPFlow 验证乘客端短信登录、刷新令牌与退出登录的完整 HTTP 流程。
func TestAuthHTTPFlow(t *testing.T) {
	var smsCode string
	userClient := client.NewLocalClient("test-signing-key", func(_ string, code string) {
		smsCode = code
	})
	server := httptest.NewServer(router.NewRouter(svc.NewServiceContext(userClient)))
	defer server.Close()

	sendResponse := callJSON(t, http.MethodPost, server.URL+"/api/passenger/v1/auth/send-sms-code", map[string]string{
		"phone": "13800138000",
	}, "")
	if sendResponse.Code != 0 {
		t.Fatalf("SendSMSCode response = %+v", sendResponse)
	}
	if smsCode == "" {
		t.Fatal("SendSMSCode did not generate a local code")
	}

	loginResponse := callJSON(t, http.MethodPost, server.URL+"/api/passenger/v1/auth/login-by-sms", map[string]string{
		"phone": "13800138000",
		"code":  smsCode,
	}, "")
	if loginResponse.Code != 0 {
		t.Fatalf("LoginBySMS response = %+v", loginResponse)
	}
	loginData := decodeData[types.LoginBySMSResponse](t, loginResponse.Data)
	if !loginData.IsNewUser || loginData.Token == "" || loginData.RefreshToken == "" {
		t.Fatalf("LoginBySMS data = %+v", loginData)
	}

	refreshResponse := callJSON(t, http.MethodPost, server.URL+"/api/passenger/v1/auth/refresh-token", map[string]string{
		"refreshToken": loginData.RefreshToken,
	}, "")
	if refreshResponse.Code != 0 {
		t.Fatalf("RefreshToken response = %+v", refreshResponse)
	}
	refreshData := decodeData[types.RefreshTokenResponse](t, refreshResponse.Data)
	if refreshData.Token == "" || refreshData.RefreshToken == "" {
		t.Fatalf("RefreshToken data = %+v", refreshData)
	}

	logoutResponse := callJSON(t, http.MethodPost, server.URL+"/api/passenger/v1/auth/logout", nil, refreshData.Token)
	if logoutResponse.Code != 0 {
		t.Fatalf("Logout response = %+v", logoutResponse)
	}
}

func callJSON(t *testing.T, method, url string, body any, token string) types.Response {
	t.Helper()

	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
		requestBody = bytes.NewReader(payload)
	}

	request, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("http.Do() error = %v", err)
	}
	defer response.Body.Close()

	var result types.Response
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("json.Decode() error = %v", err)
	}
	return result
}

func decodeData[T any](t *testing.T, data any) T {
	t.Helper()

	payload, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("json.Marshal(data) error = %v", err)
	}
	var result T
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v", err)
	}
	return result
}
