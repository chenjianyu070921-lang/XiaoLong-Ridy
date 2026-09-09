package events

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// DefaultStream 订单事件默认 Redis Stream。
const DefaultStream = "orderclient:event:stream"

// Event 事件信封，写入 Redis Stream 时整体序列化。
type Event struct {
	ID        string          `json:"id"`
	Topic     string          `json:"topic"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt int64           `json:"created_at"`
}

// Bus 定义事件发布与消费接口。
type Bus interface {
	Publish(ctx context.Context, topic string, payload []byte) error
	// Consume 以消费者组消费事件，成功后 ACK。
	// handler 返回错误时不 ACK，消息保留在 PEL 由后续投递重试；
	// 单条消息重试超过 maxEventRetries 次后写入死信队列（<stream>:dlq）并 ACK，防止永久滞留（P1-M4-10）。
	// topics 非空时只处理指定主题，其余主题的消息直接 ACK 跳过——用于同一消费者组按主题并行消费。
	Consume(ctx context.Context, group string, handler func(ctx context.Context, topic string, payload []byte) error, topics ...string) error
}

// maxEventRetries 单条事件的最大处理尝试次数，超过后进入死信队列。
const maxEventRetries = 3

// RedisStreamBus 基于 Redis Stream 的轻量事件总线。
type RedisStreamBus struct {
	client *redis.Client
	stream string
}

// NewRedisStreamBus 创建 Redis Stream 事件总线。
func NewRedisStreamBus(client *redis.Client, stream string) *RedisStreamBus {
	if stream == "" {
		stream = DefaultStream
	}
	return &RedisStreamBus{client: client, stream: stream}
}

// Publish 向指定主题发布一条事件。
func (b *RedisStreamBus) Publish(ctx context.Context, topic string, payload []byte) error {
	evt := Event{
		ID:        fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63()),
		Topic:     topic,
		Payload:   payload,
		CreatedAt: time.Now().Unix(),
	}
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return b.client.XAdd(ctx, &redis.XAddArgs{
		Stream: b.stream,
		Values: map[string]interface{}{"event": string(data)},
	}).Err()
}

// Consume 以消费者组消费事件：主题过滤 → 处理 → ACK。
// 失败消息通过 Redis Hash 记录重试次数，达到 maxEventRetries 后进死信队列（<stream>:dlq）。
func (b *RedisStreamBus) Consume(ctx context.Context, group string, handler func(context.Context, string, []byte) error, topics ...string) error {
	// 构造主题集合用于 O(1) 过滤；空 topics 表示不过滤（兼容旧调用）。
	topicSet := make(map[string]struct{}, len(topics))
	for _, tp := range topics {
		topicSet[tp] = struct{}{}
	}
	err := b.client.XGroupCreateMkStream(ctx, b.stream, group, "0").Err()
	if err != nil && !isBusyGroup(err) {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		streams, err := b.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: group,
			Streams:  []string{b.stream, ">"},
			Count:    10,
			Block:    2 * time.Second,
		}).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			logx.Errorf("read event stream %s failed: %v", b.stream, err)
			time.Sleep(time.Second)
			continue
		}
		for _, s := range streams {
			for _, msg := range s.Messages {
				b.consumeOne(ctx, group, handler, msg, topicSet)
			}
		}
	}
}

// consumeOne 处理单条消息：解析、主题过滤、重试计数、处理、ACK/死信（P1-M4-10）。
func (b *RedisStreamBus) consumeOne(ctx context.Context, group string, handler func(context.Context, string, []byte) error, msg redis.XMessage, topicSet map[string]struct{}) {
	// 重试计数 key：每个消息独立计数，成功后删除。
	retryKey := fmt.Sprintf("%s:retry:%s", b.stream, msg.ID)

	// 失败/脏数据必须收敛：ack 并跳过，避免消息无限滞留 Pending。
	failAck := func(reason string, a ...interface{}) {
		logx.Errorf(reason, a...)
		_, _ = b.client.XAck(ctx, b.stream, group, msg.ID).Result()
		_, _ = b.client.Del(ctx, retryKey).Result()
	}

	raw, ok := msg.Values["event"].(string)
	if !ok {
		failAck("event payload missing, msg id=%s, drop and ack", msg.ID)
		return
	}
	var evt Event
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		failAck("unmarshal event failed, msg id=%s err=%v, drop and ack", msg.ID, err)
		return
	}
	// 主题过滤：非本消费者关心的主题直接 ack，用于按主题并行消费。
	if len(topicSet) > 0 {
		if _, ok := topicSet[evt.Topic]; !ok {
			_, _ = b.client.XAck(ctx, b.stream, group, msg.ID).Result()
			return
		}
	}

	if err := handler(ctx, evt.Topic, evt.Payload); err != nil {
		retries, _ := b.client.HIncrBy(ctx, retryKey, "count", 1).Result()
		if retries >= maxEventRetries {
			// 重试耗尽：写入死信队列后 ACK，人工/告警介入，防止无限循环。
			dlqPayload, _ := json.Marshal(map[string]interface{}{
				"stream":  b.stream,
				"group":   group,
				"msg_id":  msg.ID,
				"event":   raw,
				"error":   err.Error(),
				"retries": retries,
			})
			if derr := b.client.XAdd(ctx, &redis.XAddArgs{
				Stream: b.stream + ":dlq",
				Values: map[string]interface{}{"event": string(dlqPayload)},
			}).Err(); derr != nil {
				logx.Errorf("write dlq failed, msg id=%s err=%v", msg.ID, derr)
			}
			_, _ = b.client.XAck(ctx, b.stream, group, msg.ID).Result()
			_, _ = b.client.Del(ctx, retryKey).Result()
			logx.Errorf("handle event %s failed after %d retries, moved to dlq, msg id=%s: %v", evt.Topic, retries, msg.ID, err)
			return
		}
		logx.Errorf("handle event %s failed (retry %d/%d), msg id=%s: %v", evt.Topic, retries, maxEventRetries, msg.ID, err)
		return
	}
	_, _ = b.client.XAck(ctx, b.stream, group, msg.ID).Result()
	_, _ = b.client.Del(ctx, retryKey).Result()
}

func isBusyGroup(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == "BUSYGROUP" || msg == "BUSYGROUP Consumer Group name already exists"
}
