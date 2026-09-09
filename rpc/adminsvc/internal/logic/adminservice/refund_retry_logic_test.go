package adminservicelogic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// TestRefundRetryLogic_ListAndRetry 验证补偿任务可查询，并可被人工提前调度。
func TestRefundRetryLogic_ListAndRetry(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run: %v", err)
	}
	defer mini.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	defer rdb.Close()

	payload := `{"order_id":1001,"order_no":"ORD1001","refund_no":"RF1001","refund_cents":2500,"operator_type":"admin","operator_id":9,"attempt":2}`
	if err := rdb.ZAdd(context.Background(), constants.RefundRetryQueueKey, redis.Z{
		Score: float64(time.Now().Add(time.Hour).Unix()), Member: payload,
	}).Err(); err != nil {
		t.Fatalf("seed queue: %v", err)
	}
	logic := NewRefundRetryLogic(context.Background(), &svc.ServiceContext{Redis: rdb})
	resp, err := logic.ListRefundRetryTasks(&adminsvc.RefundRetryTaskListRequest{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListRefundRetryTasks: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].RefundNo != "RF1001" {
		t.Fatalf("list response = %+v", resp)
	}
	if _, err := logic.RetryRefundTask(&adminsvc.RefundRetryTaskRequest{RefundNo: "RF1001"}); err != nil {
		t.Fatalf("RetryRefundTask: %v", err)
	}
	score, err := rdb.ZScore(context.Background(), constants.RefundRetryQueueKey, payload).Result()
	if err != nil {
		t.Fatalf("read retry score: %v", err)
	}
	if score > float64(time.Now().Unix()+1) {
		t.Fatalf("retry score = %v, want immediate retry", score)
	}
}
