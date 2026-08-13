package logic

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

const (
	smsCodeTTL      = 5 * time.Minute
	smsCodeCooldown = time.Minute
)

// SMSCodeSender 负责向短信服务提交验证码。
type SMSCodeSender interface {
	Send(ctx context.Context, phone string) (expireIn int64, err error)
}

// SMSCodeVerifier 负责校验手机号和验证码是否匹配。
type SMSCodeVerifier interface {
	Verify(ctx context.Context, phone, code string) (bool, error)
}

type smsCodeRecord struct {
	code      string
	expiresAt time.Time
	sentAt    time.Time
}

// MemorySMSCodeService 是短信服务的本地实现，仅用于当前开发和联调。
type MemorySMSCodeService struct {
	mu     sync.Mutex
	codes  map[string]smsCodeRecord
	onSent func(phone, code string)
}

// NewMemorySMSCodeService 创建本地内存短信服务。
func NewMemorySMSCodeService(onSent func(phone, code string)) *MemorySMSCodeService {
	return &MemorySMSCodeService{
		codes:  make(map[string]smsCodeRecord),
		onSent: onSent,
	}
}

// Send 生成并保存验证码，同时执行每手机号 60 秒一次的发送限制。
func (s *MemorySMSCodeService) Send(_ context.Context, phone string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	// 同一手机号仍在冷却窗口内时，拒绝重复发送。
	if record, ok := s.codes[phone]; ok && now.Sub(record.sentAt) < smsCodeCooldown {
		return 0, ErrSMSCodeSendTooFrequent
	}

	// 验证码只保存在内存中，后续接入 Redis 时保持相同的 TTL 和冷却语义。
	code, err := generateSMSCode()
	if err != nil {
		return 0, err
	}
	s.codes[phone] = smsCodeRecord{
		code:      code,
		sentAt:    now,
		expiresAt: now.Add(smsCodeTTL),
	}
	if s.onSent != nil {
		s.onSent(phone, code)
	}
	return int64(smsCodeTTL / time.Second), nil
}

// Verify 校验验证码；验证码成功使用后会立即失效。
func (s *MemorySMSCodeService) Verify(_ context.Context, phone, code string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.codes[phone]
	// 缓存中没有记录或超过 5 分钟有效期，都统一视为验证码过期。
	if !ok || time.Now().After(record.expiresAt) {
		delete(s.codes, phone)
		return false, ErrSMSCodeExpired
	}
	if record.code != code {
		return false, nil
	}
	// 验证成功后立即删除，避免同一验证码被重复使用。
	delete(s.codes, phone)
	return true, nil
}

func generateSMSCode() (string, error) {
	var bytes [4]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	value := uint32(bytes[0])<<24 | uint32(bytes[1])<<16 | uint32(bytes[2])<<8 | uint32(bytes[3])
	return fmt.Sprintf("%06d", value%1000000), nil
}
