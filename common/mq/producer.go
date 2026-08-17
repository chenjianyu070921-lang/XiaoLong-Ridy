package mq

import (
	"github.com/IBM/sarama"
)

// Producer Kafka 消息生产者抽象，便于业务层解耦与测试替换。
type Producer interface {
	// Send 发送一条消息。key 为空则按分区策略散列。
	Send(topic string, key string, value []byte) error
	// Close 关闭生产者。
	Close() error
}

// KafkaProducer 基于 sarama SyncProducer 的同步生产者实现。
type KafkaProducer struct {
	producer sarama.SyncProducer
}

// NewKafkaProducer 创建同步生产者。
func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll // 等所有副本确认，保证可靠投递
	cfg.Producer.Return.Successes = true          // 同步模式必须
	cfg.Producer.Return.Errors = true
	cfg.Producer.Partitioner = sarama.NewHashPartitioner

	p, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, err
	}
	return &KafkaProducer{producer: p}, nil
}

// Send 发送一条消息。
func (p *KafkaProducer) Send(topic string, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}
	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}
	_, _, err := p.producer.SendMessage(msg)
	return err
}

// Close 关闭生产者。
func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}

// NoopProducer 空实现生产者，用于 Kafka 未启动时的降级兜底。
type NoopProducer struct{}

// Send 空实现，直接丢弃消息。
func (n *NoopProducer) Send(topic string, key string, value []byte) error {
	return nil
}

// Close 空实现。
func (n *NoopProducer) Close() error {
	return nil
}
