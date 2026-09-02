package logic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/redisx"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// orderLockTTL 接单/取消分布式锁过期时间，配合看门狗续期防止持有锁的服务异常退出导致锁永久占用。
const orderLockTTL = 10 * time.Second

// localOrderLocks 是 Redis 不可用时的本进程订单锁。
// 该降级锁只覆盖单实例并发，跨实例一致性仍由 Redis 分布式锁负责。
var localOrderLocks sync.Map

// acquireOrderLock 为订单加分布式锁，避免同一订单被并发接单/取消竞态。
//
// 修复说明（P1-M4-1/M4-2）：
//   - 旧实现用 SetNX(key,1,TTL)+Del 释放，无 owner 校验，TTL 截断时会误删新锁；
//   - 改用 redisx.Lock（带唯一 token + Lua 原子释放 + 看门狗续期），杜绝误删；
//   - 始终返回非 nil 的 release 函数，即使获取失败也可安全 defer 调用，避免 nil panic。
//
// Redis 未配置（如单元测试）时返回空释放函数，退化为依赖 DB 条件更新保证并发安全。
func acquireOrderLock(ctx context.Context, rdb *redis.Client, orderID uint64) (func(), error) {
	noop := func() {}
	if rdb == nil {
		lock := &sync.Mutex{}
		actual, _ := localOrderLocks.LoadOrStore(orderID, lock)
		actual.(*sync.Mutex).Lock()
		return func() {
			actual.(*sync.Mutex).Unlock()
		}, nil
	}
	key := fmt.Sprintf(constants.RedisOrderLock, orderID)
	lk, err := redisx.TryLock(ctx, rdb, key, orderLockTTL)
	if err != nil {
		if err == redisx.ErrLockHeld {
			return noop, fmt.Errorf("order %d is being processed by another request", orderID)
		}
		return noop, err
	}
	return lk.Release, nil
}

// syncCancelDispatch 同步失效该订单的全部待派单记录。
// 客户端未配置或调用失败仅记日志，不阻断订单取消主流程。
func syncCancelDispatch(ctx context.Context, client dispatch.Dispatch, orderID uint64, reason string) {
	if client == nil {
		return
	}
	if _, err := client.CancelDispatch(ctx, &dispatch.CancelDispatchRequest{
		OrderId: int64(orderID),
		Reason:  reason,
	}); err != nil {
		logx.Errorf("sync cancel dispatch failed, orderId=%d: %v", orderID, err)
	}
}
