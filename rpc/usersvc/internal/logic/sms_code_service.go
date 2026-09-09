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

	registeredSMSHourlyLimit   = 5
	registeredSMSDailyLimit    = 10
	unregisteredSMSHourlyLimit = 3
	unregisteredSMSDailyLimit  = 5
)

// SMSRatePolicy 描述同一手机号在固定时间窗口内允许发送验证码的次数。
type SMSRatePolicy struct {
	HourLimit int
	DayLimit  int
}

// SMSCodeSender 负责向短信服务提交验证码。
type SMSCodeSender interface {
	Send(ctx context.Context, phone string) (expireIn int64, err error)
}

// SMSCodePolicySender 支持按注册状态使用不同频控阈值的验证码发送器。
type SMSCodePolicySender interface {
	SendWithPolicy(ctx context.Context, phone string, policy SMSRatePolicy) (expireIn int64, err error)
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
	sends  map[string][]time.Time
	onSent func(phone, code string)
}

// NewMemorySMSCodeService 创建本地内存短信服务。
func NewMemorySMSCodeService(onSent func(phone, code string)) *MemorySMSCodeService {
	return &MemorySMSCodeService{
		codes:  make(map[string]smsCodeRecord),
		sends:  make(map[string][]time.Time),
		onSent: onSent,
	}
}

// Send 生成并保存验证码，同时执行每手机号 60 秒一次的发送限制。
func (s *MemorySMSCodeService) Send(ctx context.Context, phone string) (int64, error) {
	return s.SendWithPolicy(ctx, phone, registeredSMSRatePolicy())
}

// SendWithPolicy 生成并保存验证码，同时执行冷却、小时和天级频控。
func (s *MemorySMSCodeService) SendWithPolicy(_ context.Context, phone string, policy SMSRatePolicy) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	// 同一手机号仍在冷却窗口内时，拒绝重复发送。
	if record, ok := s.codes[phone]; ok && now.Sub(record.sentAt) < smsCodeCooldown {
		return 0, ErrSMSCodeSendTooFrequent
	}
	if exceedsSMSRateLimit(s.sends[phone], now, policy) {
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
	s.sends[phone] = appendSMSHistory(s.sends[phone], now)
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

// registeredSMSRatePolicy 返回已注册手机号的短信频控基线。
func registeredSMSRatePolicy() SMSRatePolicy {
	return SMSRatePolicy{HourLimit: registeredSMSHourlyLimit, DayLimit: registeredSMSDailyLimit}
}

// unregisteredSMSRatePolicy 返回未注册手机号的更严格短信频控规则。
func unregisteredSMSRatePolicy() SMSRatePolicy {
	return SMSRatePolicy{HourLimit: unregisteredSMSHourlyLimit, DayLimit: unregisteredSMSDailyLimit}
}

// exceedsSMSRateLimit 判断发送历史是否已经达到小时或天级上限。
func exceedsSMSRateLimit(history []time.Time, now time.Time, policy SMSRatePolicy) bool {
	hourCount := 0
	dayCount := 0
	for _, sentAt := range history {
		if now.Sub(sentAt) < time.Hour {
			hourCount++
		}
		if now.Sub(sentAt) < 24*time.Hour {
			dayCount++
		}
	}
	return policy.HourLimit > 0 && hourCount >= policy.HourLimit ||
		policy.DayLimit > 0 && dayCount >= policy.DayLimit
}

// appendSMSHistory 追加本次发送时间，并清理超过 24 小时的历史记录。
func appendSMSHistory(history []time.Time, now time.Time) []time.Time {
	out := make([]time.Time, 0, len(history)+1)
	for _, sentAt := range history {
		if now.Sub(sentAt) < 24*time.Hour {
			out = append(out, sentAt)
		}
	}
	return append(out, now)
}
