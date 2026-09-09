package channel

import (
	"context"
	"errors"
	"testing"
)

func TestMockVerifier_Pass(t *testing.T) {
	v := &MockVerifier{}
	if err := v.Verify(context.Background(), "a=1&b=2"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestMockVerifier_Fail(t *testing.T) {
	wantErr := errors.New("sign verify failed")
	v := &MockVerifier{Err: wantErr}
	if err := v.Verify(context.Background(), "a=1"); err != wantErr {
		t.Errorf("expected %v, got %v", wantErr, err)
	}
}

func TestNewAlipayVerifier_EmptyKey(t *testing.T) {
	// 空私钥应返回错误（支付宝 SDK 校验）
	if _, err := NewAlipayVerifier("", "", "", false); err == nil {
		t.Error("expected error for empty keys")
	}
}
