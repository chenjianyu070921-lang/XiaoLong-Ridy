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
const DefaultStream = "order:event:stream"

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
	Consume(ctx context.Context, group string, handler func(ctx context.Context, topic string, payload []byte) error) error
}

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

// Consume 以消费者组消费事件，成功后 ACK，失败不 ACK 由后续重试。
func (b *RedisStreamBus) Consume(ctx context.Context, group string, handler func(context.Context, string, []byte) error) error {
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
				raw, ok := msg.Values["event"].(string)
				if !ok {
					continue
				}
				var evt Event
				if err := json.Unmarshal([]byte(raw), &evt); err != nil {
					logx.Errorf("unmarshal event failed: %v", err)
					continue
				}
				if err := handler(ctx, evt.Topic, evt.Payload); err != nil {
					logx.Errorf("handle event %s failed: %v", evt.Topic, err)
					continue
				}
				_, _ = b.client.XAck(ctx, b.stream, group, msg.ID).Result()
			}
		}
	}
}

func isBusyGroup(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == "BUSYGROUP" || msg == "BUSYGROUP Consumer Group name already exists"
}
