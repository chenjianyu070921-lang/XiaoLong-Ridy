package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/svc"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestHandleDispatchNewWritesAvailableAndPublishesPush(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	consumer := &OrderConsumer{svcCtx: &svc.ServiceContext{Redis: rdb}}
	ctx := context.Background()
	channel := fmt.Sprintf(constants.RedisDriverPush, 25)
	pubsub := rdb.Subscribe(ctx, channel)
	defer pubsub.Close()
	if _, err := pubsub.ReceiveTimeout(ctx, time.Second); err != nil {
		t.Fatalf("subscribe driver push channel error = %v", err)
	}

	payload, err := json.Marshal(DispatchNewEvent{
		OrderId:      1001,
		DriverIds:    []int64{25},
		DispatchedAt: 123,
	})
	if err != nil {
		t.Fatalf("marshal event error = %v", err)
	}
	if err := consumer.handleDispatchNew(ctx, payload); err != nil {
		t.Fatalf("handleDispatchNew() error = %v", err)
	}

	member, err := rdb.SIsMember(ctx, fmt.Sprintf(constants.RedisDriverAvailable, 25), int64(1001)).Result()
	if err != nil {
		t.Fatalf("SIsMember() error = %v", err)
	}
	if !member {
		t.Fatalf("order should be in driver available set")
	}

	msg, err := pubsub.ReceiveMessage(ctx)
	if err != nil {
		t.Fatalf("ReceiveMessage() error = %v", err)
	}
	var push DriverPushDispatchMessage
	if err := json.Unmarshal([]byte(msg.Payload), &push); err != nil {
		t.Fatalf("unmarshal push payload error = %v", err)
	}
	if push.Type != constants.TopicDispatchNew || push.OrderId != 1001 || push.DriverId != 25 {
		t.Fatalf("push payload = %+v", push)
	}
}
