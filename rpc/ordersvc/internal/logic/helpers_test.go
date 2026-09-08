package logic

import (
	"testing"

	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newTestSvcCtx 构造带可用 Redis 的测试 ServiceContext。
//
// 背景：司机侧缺陷#2 修复后，acquireOrderLock 在 Redis 不可用（含未配置）时直接拒绝接单/取消，
// 不再降级为单进程内存锁。因此接单/取消/超时类单测必须注入可用的 Redis 客户端，
// 否则会因 rdb==nil 而拒绝，无法覆盖并发安全主路径。这里用 miniredis 提供进程内 Redis。
func newTestSvcCtx(t *testing.T, repo repository.OrderRepository) *svc.ServiceContext {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return &svc.ServiceContext{OrderRepository: repo, Redis: rdb}
}
