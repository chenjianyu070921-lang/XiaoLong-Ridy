package logic

import (
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
)

// TestLogicErrorsAliasModelErrors 校验 logic 层对业务错误的统一导出。
// 乘客端和 usersvc 逻辑层都依赖这些错误名，避免合并时漏导出导致全仓无法编译。
func TestLogicErrorsAliasModelErrors(t *testing.T) {
	cases := []struct {
		name string
		got  error
		want error
	}{
		{name: "短信验证码过期", got: ErrSMSCodeExpired, want: model.ErrSMSCodeExpired},
		{name: "短信发送过于频繁", got: ErrSMSCodeSendTooFrequent, want: model.ErrSMSCodeSendTooFrequent},
		{name: "无效令牌", got: ErrInvalidToken, want: model.ErrInvalidToken},
		{name: "令牌过期", got: ErrTokenExpired, want: model.ErrTokenExpired},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !errors.Is(tc.got, tc.want) {
				t.Fatalf("logic error = %v, want alias of %v", tc.got, tc.want)
			}
		})
	}
}
