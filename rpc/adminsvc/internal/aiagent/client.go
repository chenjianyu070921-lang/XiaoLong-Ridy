package aiagent

import (
	"context"
	"errors"
)

// errModelUnavailable 表示模型未配置或调用不可用，触发降级到本地模板。
var errModelUnavailable = errors.New("model unavailable")

// ModelClient 是外部 LLM 的可插拔边界。
//
// 密钥由实现方从环境变量读取，绝不进入配置或日志。
// 返回错误时编排层降级到本地模板报告。
type ModelClient interface {
	// Generate 根据系统提示词与用户提示词返回模型原始文本。
	Generate(ctx context.Context, system, user string) (string, error)
}

// DisabledModel 在未配置模型时使用，固定返回错误，强制走降级链路。
type DisabledModel struct{}

// Generate 实现 ModelClient。
func (DisabledModel) Generate(context.Context, string, string) (string, error) {
	return "", errModelUnavailable
}
