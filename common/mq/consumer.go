package mq

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

// Consumer Kafka 消息消费者抽象。
type Consumer interface {
	// Consume 以消费者组订阅指定 topics 并阻塞消费。
	// handler 返回错误时消息进入死信队列（<topic>.dlq），避免无限重试阻塞分区。
	Consume(ctx context.Context, group string, handler func(ctx context.Context, topic string, payload []byte) error, topics ...string) error
	// Close 关闭消费者。
	Close() error
}

// KafkaConsumer 基于 sarama ConsumerGroup 的消费者实现。
type KafkaConsumer struct {
	brokers []string
	dlq     Producer // 死信生产者（可选，nil 时不写 DLQ 仅记录错误）
}

// NewKafkaConsumer 创建消费者。dlq 用于失败消息的死信队列（可传 NoopProducer 关闭）。
func NewKafkaConsumer(brokers []string, dlq Producer) *KafkaConsumer {
	return &KafkaConsumer{brokers: brokers, dlq: dlq}
}

// Consume 以消费者组消费消息。
func (c *KafkaConsumer) Consume(ctx context.Context, group string, handler func(ctx context.Context, topic string, payload []byte) error, topics ...string) error {
	if len(topics) == 0 {
		return errors.New("no topics to consume")
	}
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Return.Errors = true
	cg, err := sarama.NewConsumerGroup(c.brokers, group, cfg)
	if err != nil {
		return err
	}
	defer cg.Close()

	gh := &consumerGroupHandler{handle: handler, dlq: c.dlq}
	for {
		if err := cg.Consume(ctx, topics, gh); err != nil {
			if err == sarama.ErrClosedConsumerGroup {
				return nil
			}
			logx.Errorf("kafka consume error, group=%s topics=%v: %v", group, topics, err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
			}
			continue
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// Close 关闭消费者。ConsumerGroup 在 Consume 返回后自动关闭，此处仅占位实现接口。
func (c *KafkaConsumer) Close() error { return nil }

// consumerGroupHandler 实现 sarama.ConsumerGroupHandler。
type consumerGroupHandler struct {
	handle func(ctx context.Context, topic string, payload []byte) error
	dlq    Producer
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.processMessage(session.Context(), msg)
		session.MarkMessage(msg, "")
	}
	return nil
}

// processMessage 调用业务处理函数；失败时写入死信队列（<topic>.dlq）。
func (h *consumerGroupHandler) processMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	if err := h.handle(ctx, msg.Topic, msg.Value); err != nil {
		// 处理失败：写死信队列并记录错误，防止单条坏消息阻塞分区。
		h.writeDLQ(msg, err)
		return err
	}
	return nil
}

// writeDLQ 将失败消息写入 <topic>.dlq 死信队列。
func (h *consumerGroupHandler) writeDLQ(msg *sarama.ConsumerMessage, cause error) {
	if h.dlq == nil {
		logx.Errorf("handle message failed, topic=%s partition=%d offset=%d: %v", msg.Topic, msg.Partition, msg.Offset, cause)
		return
	}
	dlqPayload, _ := json.Marshal(map[string]interface{}{
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
		"value":     string(msg.Value),
		"error":     cause.Error(),
	})
	if err := h.dlq.Send(msg.Topic+".dlq", "", dlqPayload); err != nil {
		logx.Errorf("write dlq failed, topic=%s: %v", msg.Topic, err)
		return
	}
	logx.Errorf("handle message failed, moved to dlq %s, topic=%s partition=%d offset=%d: %v",
		msg.Topic+".dlq", msg.Topic, msg.Partition, msg.Offset, cause)
}
