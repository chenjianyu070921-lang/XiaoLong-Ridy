package mq

import (
	"testing"

	"github.com/IBM/sarama/mocks"
)

func TestKafkaProducer_Send(t *testing.T) {
	mp := mocks.NewSyncProducer(t, nil)
	mp.ExpectSendMessageAndSucceed()

	p := &KafkaProducer{producer: mp}
	if err := p.Send("test-topic", "key1", []byte("hello")); err != nil {
		t.Fatalf("Send error: %v", err)
	}
	// mocks 库会在 Close 时校验所有期望的消息已发送
	if err := mp.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

func TestKafkaProducer_SendEmptyKey(t *testing.T) {
	mp := mocks.NewSyncProducer(t, nil)
	mp.ExpectSendMessageAndSucceed()

	p := &KafkaProducer{producer: mp}
	if err := p.Send("test-topic", "", []byte("hello")); err != nil {
		t.Fatalf("Send error: %v", err)
	}
	if err := mp.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}
