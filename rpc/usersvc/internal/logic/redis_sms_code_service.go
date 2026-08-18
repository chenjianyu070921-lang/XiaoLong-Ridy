package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	commonSMS "XiaoLong-Ridy/common/sms"

	"github.com/redis/go-redis/v9"
)

// RedisSMSCodeService 使用 Redis 保存短信验证码和发送冷却状态。
type RedisSMSCodeService struct {
	client *redis.Client
	sender commonSMS.Sender
	onSent func(phone, code string)
}

// NewRedisSMSCodeService 创建生产环境使用的验证码服务。
func NewRedisSMSCodeService(client *redis.Client, sender commonSMS.Sender, onSent func(phone, code string)) *RedisSMSCodeService {
	return &RedisSMSCodeService{
		client: client,
		sender: sender,
		onSent: onSent,
	}
}

// Send 生成验证码并写入 Redis，同时为手机号设置冷却键。
func (s *RedisSMSCodeService) Send(ctx context.Context, phone string) (int64, error) {
	return s.SendWithPolicy(ctx, phone, registeredSMSRatePolicy())
}

// SendWithPolicy 生成验证码并写入 Redis，同时执行冷却、小时和天级频控。
func (s *RedisSMSCodeService) SendWithPolicy(ctx context.Context, phone string, policy SMSRatePolicy) (int64, error) {
	exists, err := s.client.Exists(ctx, smsCooldownKey(phone)).Result()
	if err != nil {
		return 0, err
	}
	if exists > 0 {
		return 0, ErrSMSCodeSendTooFrequent
	}
	if limited, err := s.exceedsRateLimit(ctx, phone, policy); err != nil {
		return 0, err
	} else if limited {
		return 0, ErrSMSCodeSendTooFrequent
	}

	code, err := generateSMSCode()
	if err != nil {
		return 0, err
	}
	if s.sender != nil {
		if err := s.sender.Send(ctx, phone, code); err != nil {
			return 0, err
		}
	}
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, smsCodeKey(phone), code, smsCodeTTL)
	pipe.Set(ctx, smsCooldownKey(phone), "1", smsCodeCooldown)
	pipe.Incr(ctx, smsHourCountKey(phone))
	pipe.Expire(ctx, smsHourCountKey(phone), time.Hour)
	pipe.Incr(ctx, smsDayCountKey(phone))
	pipe.Expire(ctx, smsDayCountKey(phone), 24*time.Hour)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	if s.onSent != nil {
		s.onSent(phone, code)
	}
	return int64(smsCodeTTL.Seconds()), nil
}

// exceedsRateLimit 判断 Redis 中记录的小时和天级发送次数是否已达到阈值。
func (s *RedisSMSCodeService) exceedsRateLimit(ctx context.Context, phone string, policy SMSRatePolicy) (bool, error) {
	hourCount, err := s.client.Get(ctx, smsHourCountKey(phone)).Int()
	if errors.Is(err, redis.Nil) {
		hourCount = 0
	} else if err != nil {
		return false, err
	}

	dayCount, err := s.client.Get(ctx, smsDayCountKey(phone)).Int()
	if errors.Is(err, redis.Nil) {
		dayCount = 0
	} else if err != nil {
		return false, err
	}

	return policy.HourLimit > 0 && hourCount >= policy.HourLimit ||
		policy.DayLimit > 0 && dayCount >= policy.DayLimit, nil
}

// Verify 校验 Redis 中保存的验证码，验证成功后立即删除验证码和冷却键。
func (s *RedisSMSCodeService) Verify(ctx context.Context, phone, code string) (bool, error) {
	savedCode, err := s.client.Get(ctx, smsCodeKey(phone)).Result()
	if errors.Is(err, redis.Nil) {
		return false, ErrSMSCodeExpired
	}
	if err != nil {
		return false, err
	}
	if savedCode != code {
		return false, nil
	}
	if err := s.client.Del(ctx, smsCodeKey(phone), smsCooldownKey(phone)).Err(); err != nil {
		return false, err
	}
	return true, nil
}

// smsCodeKey 返回短信验证码 Redis key。
func smsCodeKey(phone string) string {
	return fmt.Sprintf("usersvc:sms:code:%s", phone)
}

// smsCooldownKey 返回短信发送冷却 Redis key。
func smsCooldownKey(phone string) string {
	return fmt.Sprintf("usersvc:sms:cooldown:%s", phone)
}

// smsHourCountKey 返回同一手机号 1 小时发送次数 Redis key。
func smsHourCountKey(phone string) string {
	return fmt.Sprintf("usersvc:sms:hour:%s", phone)
}

// smsDayCountKey 返回同一手机号 24 小时发送次数 Redis key。
func smsDayCountKey(phone string) string {
	return fmt.Sprintf("usersvc:sms:day:%s", phone)
}
