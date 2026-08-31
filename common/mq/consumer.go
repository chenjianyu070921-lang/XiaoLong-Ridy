package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// Consumer Kafka 消息消费者抽象。
type Consumer interface {
	// Consume 订阅指定 topics 并阻塞消费。
	// handler 返回错误时消息进入死信队列（<topic>.dlq），避免无限重试阻塞分区。
	Consume(ctx context.Context, group string, handler func(ctx context.Context, topic string, payload []byte) error, topics ...string) error
	// Close 关闭消费者。
	Close() error
}

// KafkaConsumer 基于 sarama 简单消费者（PartitionConsumer）的实现。
//
// 说明：当前部署环境中的 Kafka 无法完成 ConsumerGroup 协议协商
// （FindCoordinator/JoinGroup 返回未知错误码，sarama / segmentio 均复现），
// 因此本实现不再使用 ConsumerGroup，而是对每个 topic 的全部分区直接消费，
// offset 持久化到 Redis（SetOffsetStore 注入），实现单实例下的可靠消费：
//   - 首次启动且无 offset 记录时从最新消息（OffsetNewest）开始，避免重放历史事件；
//   - 重启后从 Redis 记录的最后 offset 续读，避免漏消费。
type KafkaConsumer struct {
	brokers []string
	dlq     Producer // 死信生产者（可选，nil 时不写 DLQ 仅记录错误）
	redis   *redis.Client
}

// NewKafkaConsumer 创建消费者。dlq 用于失败消息的死信队列（可传 NoopProducer 关闭）。
func NewKafkaConsumer(brokers []string, dlq Producer) *KafkaConsumer {
	return &KafkaConsumer{brokers: brokers, dlq: dlq}
}

// SetOffsetStore 注入 offset 持久化存储（Redis）。
func (c *KafkaConsumer) SetOffsetStore(rc *redis.Client) *KafkaConsumer {
	c.redis = rc
	return c
}

// Consume 消费指定 topics 的全部分区并阻塞。
func (c *KafkaConsumer) Consume(ctx context.Context, group string, handler func(ctx context.Context, topic string, payload []byte) error, topics ...string) error {
	if len(topics) == 0 {
		return errors.New("no topics to consume")
	}
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_1_0_0
	consumer, err := sarama.NewConsumer(c.brokers, cfg)
	if err != nil {
		return err
	}
	defer consumer.Close()

	var wg sync.WaitGroup
	for _, topic := range topics {
		parts, err := consumer.Partitions(topic)
		if err != nil {
			logx.Errorf("get partitions for topic %s: %v", topic, err)
			continue
		}
		for _, p := range parts {
			wg.Add(1)
			go func(topic string, p int32) {
				defer wg.Done()
				c.consumePartition(ctx, consumer, topic, p, handler)
			}(topic, p)
		}
	}
	wg.Wait()
	return ctx.Err()
}

// consumePartition 消费单个分区，offset 从 Redis 记录续读，无记录则从最新开始。
func (c *KafkaConsumer) consumePartition(ctx context.Context, consumer sarama.Consumer, topic string, partition int32, handler func(ctx context.Context, topic string, payload []byte) error) {
	offset := c.loadOffset(ctx, topic, partition)
	pc, err := consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		logx.Errorf("consume partition failed, topic=%s p=%d offset=%d: %v", topic, partition, offset, err)
		return
	}
	defer pc.Close()
	logx.Infof("consumer start, topic=%s p=%d from offset=%d", topic, partition, offset)
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-pc.Messages():
			if !ok {
				return
			}
			c.processMessage(ctx, msg, handler)
			c.saveOffset(ctx, topic, partition, msg.Offset+1)
		case perr, ok := <-pc.Errors():
			if !ok {
				return
			}
			logx.Errorf("partition consumer error, topic=%s p=%d: %v", topic, partition, perr)
		}
	}
}

// offsetKey 返回某 topic 分区的 offset 持久化 key。
func (c *KafkaConsumer) offsetKey(topic string, partition int32) string {
	return fmt.Sprintf("kafka:consumer:offset:%s:%d", topic, partition)
}

// loadOffset 读取上次消费 offset；无记录时返回 OffsetNewest（从最新开始，避免重放历史）。
func (c *KafkaConsumer) loadOffset(ctx context.Context, topic string, partition int32) int64 {
	if c.redis == nil {
		return sarama.OffsetNewest
	}
	v, err := c.redis.Get(ctx, c.offsetKey(topic, partition)).Result()
	if err != nil {
		return sarama.OffsetNewest
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n < 0 {
		return sarama.OffsetNewest
	}
	return n
}

// saveOffset 持久化消费进度（下一条待消费 offset）。
func (c *KafkaConsumer) saveOffset(ctx context.Context, topic string, partition int32, next int64) {
	if c.redis == nil {
		return
	}
	_ = c.redis.Set(ctx, c.offsetKey(topic, partition), next, 0).Err()
}

// processMessage 调用业务处理函数；失败时写入死信队列（<topic>.dlq）。
func (c *KafkaConsumer) processMessage(ctx context.Context, msg *sarama.ConsumerMessage, handle func(ctx context.Context, topic string, payload []byte) error) {
	if err := handle(ctx, msg.Topic, msg.Value); err != nil {
		// 处理失败：写死信队列并记录错误，防止单条坏消息阻塞分区。
		c.writeDLQ(msg, err)
		return
	}
	logx.Infof("consume ok, topic=%s partition=%d offset=%d", msg.Topic, msg.Partition, msg.Offset)
}

// writeDLQ 将失败消息写入 <topic>.dlq 死信队列。
func (c *KafkaConsumer) writeDLQ(msg *sarama.ConsumerMessage, cause error) {
	if c.dlq == nil {
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
	if err := c.dlq.Send(msg.Topic+".dlq", "", dlqPayload); err != nil {
		logx.Errorf("write dlq failed, topic=%s: %v", msg.Topic, err)
		return
	}
	logx.Errorf("handle message failed, moved to dlq %s, topic=%s partition=%d offset=%d: %v",
		msg.Topic+".dlq", msg.Topic, msg.Partition, msg.Offset, cause)
}

// Close 关闭消费者。消费 goroutine 随 ctx 退出，此处仅占位实现接口。
func (c *KafkaConsumer) Close() error { return nil }
