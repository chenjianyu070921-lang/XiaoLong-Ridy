package aiagent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// arkDefaultBaseURL 是火山方舟（Volcengine Ark）推理 API 的 OpenAI 兼容基址。
// 方舟聊天补全路径为 {baseURL}/chat/completions。
const arkDefaultBaseURL = "https://ark.cn-beijing.volces.com/api/v3"

// ArkAPIKeyEnv 是火山方舟 API Key 的环境变量名。
// 密钥仅从环境变量读取，绝不写入配置或日志（与设计文档的安全边界一致）。
const ArkAPIKeyEnv = "ARK_API_KEY"

// ArkModel 是火山方舟的 ModelClient 实现，遵循 OpenAI 兼容的聊天补全协议。
// 当缺少模型名（Model.Name）或缺少 ARK_API_KEY 环境变量时，Generate 返回
// errModelUnavailable，编排层据此降级到本地模板报告，行为与未配置模型一致。
type ArkModel struct {
	baseURL    string
	model      string
	apiKey     string
	httpClient *http.Client
}

// NewArkModel 构造火山方舟模型客户端。
// baseURL 取 ModelConfig.Endpoint，为空时回退到方舟默认地址；
// 模型名取 ModelConfig.Name（火山方舟推理接入点 ID 或模型名）；
// 密钥从 ARK_API_KEY 环境变量读取。
func NewArkModel(cfg ModelConfig) *ArkModel {
	apiKey := os.Getenv(ArkAPIKeyEnv)
	base := strings.TrimRight(cfg.Endpoint, "/")
	if base == "" {
		base = arkDefaultBaseURL
	}
	cli := &http.Client{Timeout: 30 * time.Second}
	if cfg.TimeoutSeconds > 0 {
		cli.Timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}
	// 启动期记录模型就绪状态，便于确认接入点与密钥是否真正注入（不打印密钥本身）。
	logx.Infof("aiagent ark client ready: model=%q apiKeySet=%v endpoint=%s", cfg.Name, apiKey != "", base)
	return &ArkModel{
		baseURL:    base,
		model:      cfg.Name,
		apiKey:     apiKey,
		httpClient: cli,
	}
}

// Generate 调用火山方舟聊天补全接口，返回模型原始文本（预期为 JSON 对象字符串）。
func (m *ArkModel) Generate(ctx context.Context, system, user string) (string, error) {
	if m.model == "" || m.apiKey == "" {
		logx.Errorf("aiagent ark model unavailable: modelSet=%v apiKeySet=%v", m.model != "", m.apiKey != "")
		return "", errModelUnavailable
	}
	reqBody := arkChatRequest{
		Model: m.model,
		Messages: []arkMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.3,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	url := m.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		logx.Errorf("ark chat completion request failed: model=%s err=%v", m.model, err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		snippet := string(body)
		if len(snippet) > 300 {
			snippet = snippet[:300]
		}
		logx.Errorf("ark chat completion failed: status=%d model=%s body=%s", resp.StatusCode, m.model, snippet)
		return "", fmt.Errorf("ark chat completion failed: status=%d: %s", resp.StatusCode, snippet)
	}
	var parsed arkChatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("ark chat completion returned no choices")
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return "", errors.New("ark chat completion returned empty content")
	}
	return content, nil
}

// GenerateStream 流式调用火山方舟：每收到一段增量文本即回调 onDelta，最终返回拼接后的完整文本。
// 与 Generate 走同一接入点与鉴权，仅请求体带 stream=true 并按 SSE 帧解析。
func (m *ArkModel) GenerateStream(ctx context.Context, system, user string, onDelta func(string)) (string, error) {
	if m.model == "" || m.apiKey == "" {
		logx.Errorf("aiagent ark model unavailable(stream): modelSet=%v apiKeySet=%v", m.model != "", m.apiKey != "")
		return "", errModelUnavailable
	}
	reqBody := arkChatRequest{
		Model: m.model,
		Messages: []arkMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.3,
		Stream:      true,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		logx.Errorf("ark stream request failed: model=%s err=%v", m.model, err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		snippet := string(body)
		if len(snippet) > 300 {
			snippet = snippet[:300]
		}
		logx.Errorf("ark stream failed: status=%d model=%s body=%s", resp.StatusCode, m.model, snippet)
		return "", fmt.Errorf("ark chat completion failed: status=%d: %s", resp.StatusCode, snippet)
	}

	var sb strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var chunk arkStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		sb.WriteString(delta)
		if onDelta != nil {
			onDelta(delta)
		}
	}
	if err := scanner.Err(); err != nil {
		logx.Errorf("ark stream read failed: model=%s err=%v", m.model, err)
		return sb.String(), err
	}
	return sb.String(), nil
}

// arkStreamChunk 是方舟流式响应中的单个 data 帧（仅取增量文本）。
type arkStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// arkMessage 是聊天消息。
type arkMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// arkChatRequest 是 OpenAI 兼容的聊天补全请求体。
type arkChatRequest struct {
	Model       string       `json:"model"`
	Messages    []arkMessage `json:"messages"`
	Temperature float64      `json:"temperature"`
	// Stream 为 true 时方舟以 SSE 增量返回，供网关逐字渲染。
	Stream bool `json:"stream,omitempty"`
}

// arkChatResponse 是火山方舟聊天补全响应（仅取所需字段）。
type arkChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
