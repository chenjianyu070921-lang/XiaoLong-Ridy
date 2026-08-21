package realname

import (
	"context"
)

// Verifier 定义实名认证能力接口，支持多种实现（腾讯云、阿里云等）。
type Verifier interface {
	// Verify 校验姓名和身份证号的真实性和一致性。
	// 返回的 VerifyResult.Result 为 "0" 表示认证通过，其他值表示失败或异常。
	Verify(ctx context.Context, name, idCardNo string) (*VerifyResult, error)
}

// VerifyResult 实名认证结果，包含腾讯云返回的结果码和描述信息。
type VerifyResult struct {
	Result      string // 结果码: "0"=通过, "-1"=不一致, "-2"~"-7"=异常
	Description string // 结果描述（如"姓名和身份证号一致"）
}
