package events

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/common/mq"
)

// fakeBusProducer 记录发布调用，验证 KafkaBus 无信封直发。
type fakeBusProducer struct {
	topic string
	key   string
	value []byte
	err   error
}

func (f *fakeBusProducer) Send(topic string, key string, value []byte) error {
	f.topic = topic
	f.key = key
	f.value = value
	return f.err
}

func (f *fakeBusProducer) Close() error { return nil }

// fakeBusConsumer 记录消费订阅，验证 KafkaBus 转发。
type fakeBusConsumer struct {
	group   string
	topics  []string
	handler func(ctx context.Context, topic string, payload []byte) error
	err     error
}

func (f *fakeBusConsumer) Consume(ctx context.Context, group string, handler func(ctx context.Context, topic string, payload []byte) error, topics ...string) error {
	f.group = group
	f.handler = handler
	f.topics = topics
	return f.err
}

func (f *fakeBusConsumer) Close() error { return nil }

// TestKafkaBus_Publish_NoEnvelope Publish 应无信封直发：topic/key/value 原样传给 Producer。
func TestKafkaBus_Publish_NoEnvelope(t *testing.T) {
	p := &fakeBusProducer{}
	b := NewKafkaBus(p, nil)
	payload := []byte(`{"order_id":1}`)
	if err := b.Publish(context.Background(), "order.created", payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.topic != "order.created" {
		t.Fatalf("topic = %q, want order.created", p.topic)
	}
	if p.key != "" {
		t.Fatalf("key = %q, want empty (no envelope key)", p.key)
	}
	if string(p.value) != string(payload) {
		t.Fatalf("value = %s, want %s (raw payload, no envelope)", p.value, payload)
	}
}

// TestKafkaBus_Publish_NoProducer 未配置生产者时 Publish 应返回错误。
func TestKafkaBus_Publish_NoProducer(t *testing.T) {
	b := NewKafkaBus(nil, nil)
	if err := b.Publish(context.Background(), "order.created", []byte(`{}`)); err == nil {
		t.Fatal("expected error when producer is nil")
	}
}

// TestKafkaBus_Publish_PropagatesError 生产者失败时 Publish 应透传错误。
func TestKafkaBus_Publish_PropagatesError(t *testing.T) {
	p := &fakeBusProducer{err: errors.New("kafka down")}
	b := NewKafkaBus(p, nil)
	if err := b.Publish(context.Background(), "order.created", []byte(`{}`)); err == nil {
		t.Fatal("expected propagated error")
	}
}

// TestKafkaBus_Consume_ForwardsToConsumer Consume 应原样转发 group/handler/topics。
func TestKafkaBus_Consume_ForwardsToConsumer(t *testing.T) {
	c := &fakeBusConsumer{}
	b := NewKafkaBus(nil, c)
	handler := func(ctx context.Context, topic string, payload []byte) error { return nil }
	topics := []string{"order.created", "dispatch.new", "order.paid"}
	if err := b.Consume(context.Background(), "orderclient", handler, topics...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.group != "orderclient" {
		t.Fatalf("group = %q, want orderclient", c.group)
	}
	if len(c.topics) != 3 || c.topics[0] != "order.created" || c.topics[1] != "dispatch.new" || c.topics[2] != "order.paid" {
		t.Fatalf("topics = %v, want [order.created dispatch.new order.paid]", c.topics)
	}
	if c.handler == nil {
		t.Fatal("handler should be forwarded")
	}
	// 转发后的 handler 应能回调：模拟一次消费。
	if err := c.handler(context.Background(), "order.paid", []byte(`{"order_id":1}`)); err != nil {
		t.Fatalf("handler should be callable: %v", err)
	}
}

// TestKafkaBus_Consume_NoConsumer 未配置消费者时 Consume 应返回错误。
func TestKafkaBus_Consume_NoConsumer(t *testing.T) {
	b := NewKafkaBus(nil, nil)
	if err := b.Consume(context.Background(), "g", func(ctx context.Context, topic string, payload []byte) error { return nil }, "t"); err == nil {
		t.Fatal("expected error when consumer is nil")
	}
}

// TestKafkaBus_ImplementsBus KafkaBus 应满足 Bus 接口。
var _ Bus = (*KafkaBus)(nil)

// TestKafkaBus_Consume_PropagatesError 消费者失败时 Consume 应透传错误。
func TestKafkaBus_Consume_PropagatesError(t *testing.T) {
	c := &fakeBusConsumer{err: errors.New("kafka down")}
	b := NewKafkaBus(nil, c)
	if err := b.Consume(context.Background(), "g", func(ctx context.Context, topic string, payload []byte) error { return nil }, "t"); err == nil {
		t.Fatal("expected propagated error")
	}
}

// TestKafkaBus_NoopProducerRoundTrip KafkaBus 配合 NoopProducer 应不报错（Kafka 降级场景）。
func TestKafkaBus_NoopProducerRoundTrip(t *testing.T) {
	b := NewKafkaBus(&mq.NoopProducer{}, nil)
	if err := b.Publish(context.Background(), "order.created", []byte(`{"order_id":1}`)); err != nil {
		t.Fatalf("noop producer should not error: %v", err)
	}
}
