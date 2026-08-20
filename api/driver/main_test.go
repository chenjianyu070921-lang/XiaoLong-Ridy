package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/common/jwtx"
)

func TestAgentChatEndpointRequiresDriverTokenAndRunsAgent(t *testing.T) {
	const signingKey = "agent-chat-test-key"
	const serviceToken = "agent-chat-service-token"
	t.Setenv("DRIVER_AGENT_SERVICE_TOKEN", serviceToken)
	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: signingKey})

	unauthorizedRequest := httptest.NewRequest(http.MethodPost, "/api/driver/v1/agent/chat", bytes.NewBufferString(`{"question":"price for product 1001"}`))
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorizedResponse.Code, http.StatusUnauthorized)
	}

	serviceRequest := httptest.NewRequest(http.MethodPost, "/api/driver/v1/agent/chat", bytes.NewBufferString(`{"question":"price for product 1001"}`))
	serviceRequest.Header.Set("X-Internal-Service-Token", serviceToken)
	serviceResponse := httptest.NewRecorder()
	handler.ServeHTTP(serviceResponse, serviceRequest)
	if serviceResponse.Code != http.StatusOK {
		t.Fatalf("service status = %d, want %d: %s", serviceResponse.Code, http.StatusOK, serviceResponse.Body.String())
	}

	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     1,
		AccountType:   "driver",
		AccountStatus: 2,
		TTL:           time.Minute,
	}, signingKey)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/driver/v1/agent/chat", bytes.NewBufferString(`{"question":"price for product 1001"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}

	var body struct {
		Code int `json:"code"`
		Data struct {
			Answer       string `json:"answer"`
			LoopCount    int    `json:"loopCount"`
			Mode         string `json:"mode"`
			Observations []struct {
				ToolName string `json:"toolName"`
			} `json:"observations"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 || body.Data.Answer == "" || body.Data.LoopCount != 2 || body.Data.Mode != "scripted" || len(body.Data.Observations) != 1 || body.Data.Observations[0].ToolName != "get_product_price" {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}
