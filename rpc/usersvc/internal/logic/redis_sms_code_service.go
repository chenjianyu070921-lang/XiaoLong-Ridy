package logic

import (
	"context"
	"errors"
	"fmt"
	"log"
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

// verifyScript 在 Redis 服务器端原子完成“比对验证码并删除”操作，
// 避免并发请求在 GET 与 DEL 之间同时读到同一验证码。
var verifyScript = redis.NewScript(`
local saved = redis.call('GET', KEYS[1])
if not saved then return -1 end
if saved ~= ARGV[1] then return 0 end
redis.call('DEL', KEYS[1], KEYS[2])
return 1
`)

// sendLimitScript 原子检查冷却和窗口计数，并在允许时写入验证码及计数。
var sendLimitScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[2]) == 1 then return -1 end
local hour = tonumber(redis.call('GET', KEYS[3]) or '0')
local day = tonumber(redis.call('GET', KEYS[4]) or '0')
if (tonumber(ARGV[3]) > 0 and hour >= tonumber(ARGV[3])) or (tonumber(ARGV[4]) > 0 and day >= tonumber(ARGV[4])) then return -2 end
redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[5])
redis.call('SET', KEYS[2], '1', 'EX', ARGV[6])
redis.call('INCR', KEYS[3]); redis.call('EXPIRE', KEYS[3], 3600)
redis.call('INCR', KEYS[4]); redis.call('EXPIRE', KEYS[4], 86400)
return 1
`)

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
	code, err := generateSMSCode()
	if err != nil {
		return 0, err
	}
	// 先把验证码写入 Redis，再调用短信通道，避免用户收到一个服务端无法校验的验证码。
	result, err := sendLimitScript.Run(ctx, s.client, []string{smsCodeKey(phone), smsCooldownKey(phone), smsHourCountKey(phone), smsDayCountKey(phone)}, code, "1", policy.HourLimit, policy.DayLimit, int(smsCodeTTL/time.Second), int(smsCodeCooldown/time.Second)).Int()
	if err != nil {
		return 0, err
	}
	if result == -1 || result == -2 {
		return 0, ErrSMSCodeSendTooFrequent
	}
	if s.sender != nil {
		if err := s.sender.Send(ctx, phone, code); err != nil {
			// 短信通道失败时尽力清理状态；清理失败必须记录，交由补偿任务处理。
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			cleanupErr := s.client.Del(cleanupCtx, smsCodeKey(phone), smsCooldownKey(phone)).Err()
			cancel()
			if cleanupErr != nil {
				// 不打印验证码，避免凭证进入生产日志。
				log.Printf("usersvc sms state cleanup failed phone=%s err=%v", MaskPhone(phone), cleanupErr)
			}
			return 0, err
		}
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
	result, err := verifyScript.Run(ctx, s.client, []string{smsCodeKey(phone), smsCooldownKey(phone)}, code).Int()
	if err != nil {
		return false, err
	}
	if result < 0 {
		return false, ErrSMSCodeExpired
	}
	return result == 1, nil
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
