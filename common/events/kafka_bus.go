package events

import (
	"context"
	"errors"

	"XiaoLong-Ridy/common/mq"
)

// KafkaBus 基于 Kafka 的事件总线，发布与消费方式与支付模块（paysvc）完全对齐：
//   - 发布：无信封直发，调用 mq.Producer.Send(topic, "", payload)，消息体即业务 payload；
//   - 消费：基于 Kafka ConsumerGroup 按 topic 订阅，handler 收到 (topic, payload)。
type KafkaBus struct {
	producer mq.Producer
	consumer mq.Consumer
}

// NewKafkaBus 创建 Kafka 事件总线。producer/consumer 至少需要一个，按用途传入。
// 仅发布：consumer 传 nil；仅消费：producer 传 nil。
func NewKafkaBus(producer mq.Producer, consumer mq.Consumer) *KafkaBus {
	return &KafkaBus{producer: producer, consumer: consumer}
}

// Publish 发布事件：无信封直发，与 paysvc 的 Producer.Send 完全一致。
func (b *KafkaBus) Publish(ctx context.Context, topic string, payload []byte) error {
	if b.producer == nil {
		return errors.New("kafka producer not configured")
	}
	return b.producer.Send(topic, "", payload)
}

// Consume 以消费者组消费指定 topics，handler 收到 (ctx, topic, payload)。
func (b *KafkaBus) Consume(ctx context.Context, group string, handler func(ctx context.Context, topic string, payload []byte) error, topics ...string) error {
	if b.consumer == nil {
		return errors.New("kafka consumer not configured")
	}
	return b.consumer.Consume(ctx, group, handler, topics...)
}
