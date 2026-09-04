package mq

import (
	"context"
	"errors"
	"testing"

	"github.com/IBM/sarama"
)

// fakeDLQProducer 记录写入死信队列调用的测试生产者。
type fakeDLQProducer struct {
	topic string
	key   string
	value []byte
	err   error
}

func (f *fakeDLQProducer) Send(topic string, key string, value []byte) error {
	f.topic = topic
	f.key = key
	f.value = value
	return f.err
}

func (f *fakeDLQProducer) Close() error { return nil }

func msg(topic string, payload []byte) *sarama.ConsumerMessage {
	return &sarama.ConsumerMessage{Topic: topic, Partition: 0, Offset: 1, Value: payload}
}

// TestProcessMessage_Success 处理成功应不写 DLQ。
func TestProcessMessage_Success(t *testing.T) {
	dlq := &fakeDLQProducer{}
	c := &KafkaConsumer{dlq: dlq}
	c.processMessage(context.Background(), msg("order.created", []byte(`{"order_id":1}`)),
		func(ctx context.Context, topic string, payload []byte) error { return nil })
	if dlq.topic != "" {
		t.Fatalf("dlq should not be written on success, got topic=%s", dlq.topic)
	}
}

// TestProcessMessage_Failure_WritesDLQ 处理失败应写入 <topic>.dlq 死信队列。
func TestProcessMessage_Failure_WritesDLQ(t *testing.T) {
	dlq := &fakeDLQProducer{}
	c := &KafkaConsumer{dlq: dlq}
	c.processMessage(context.Background(), msg("order.paid", []byte(`{"order_id":1}`)),
		func(ctx context.Context, topic string, payload []byte) error { return errors.New("boom") })
	if dlq.topic != "order.paid.dlq" {
		t.Fatalf("dlq topic = %q, want %q", dlq.topic, "order.paid.dlq")
	}
	if len(dlq.value) == 0 {
		t.Fatal("dlq payload should not be empty")
	}
}

// TestProcessMessage_Failure_NoDLQ 未配置 DLQ 时失败仅记日志，不 panic。
func TestProcessMessage_Failure_NoDLQ(t *testing.T) {
	c := &KafkaConsumer{dlq: nil}
	c.processMessage(context.Background(), msg("order.created", []byte(`{}`)),
		func(ctx context.Context, topic string, payload []byte) error { return errors.New("boom") })
}

// TestNewKafkaConsumer_NilBrokers 空 brokers 不应 panic，返回可用的消费者对象。
func TestNewKafkaConsumer_NilBrokers(t *testing.T) {
	c := NewKafkaConsumer(nil, nil)
	if c == nil {
		t.Fatal("consumer should not be nil")
	}
	if err := c.Close(); err != nil {
		t.Fatalf("close should not error: %v", err)
	}
}
