package logic

import (
	"context"
	"sync"
)

type MemorySMSCodeVerifier struct {
	mu    sync.RWMutex
	codes map[string]string
}

// NewMemorySMSCodeVerifier 创建一个简单的内存验证码校验器，便于本地联调。
func NewMemorySMSCodeVerifier() *MemorySMSCodeVerifier {
	return &MemorySMSCodeVerifier{codes: make(map[string]string)}
}

// SetCode 为测试或本地调试设置手机号和验证码。
func (v *MemorySMSCodeVerifier) SetCode(phone, code string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.codes[phone] = code
}

// Verify 对比预置验证码和调用方传入的验证码。
func (v *MemorySMSCodeVerifier) Verify(_ context.Context, phone, code string) (bool, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	expected, ok := v.codes[phone]
	return ok && expected == code, nil
}
