package logic

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/constants"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// orderLockTTL 接单/取消分布式锁过期时间，防止持有锁的服务异常退出导致锁永久占用。
const orderLockTTL = 10 * time.Second

// acquireOrderLock 为订单加分布式锁，避免同一订单被并发接单/取消竞态。
// Redis 未配置（如单元测试）时返回空释放函数，退化为依赖 DB 条件更新保证并发安全。
func acquireOrderLock(ctx context.Context, rdb *redis.Client, orderID uint64) (func(), error) {
	if rdb == nil {
		return func() {}, nil
	}
	key := fmt.Sprintf(constants.RedisOrderLock, orderID)
	ok, err := rdb.SetNX(ctx, key, 1, orderLockTTL).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("order %d is being processed by another request", orderID)
	}
	return func() {
		releaseOrderLock(ctx, rdb, orderID)
	}, nil
}

// releaseOrderLock 释放订单分布式锁。
func releaseOrderLock(ctx context.Context, rdb *redis.Client, orderID uint64) {
	if rdb == nil {
		return
	}
	key := fmt.Sprintf(constants.RedisOrderLock, orderID)
	if err := rdb.Del(ctx, key).Err(); err != nil {
		logx.Errorf("release order lock failed, orderId=%d: %v", orderID, err)
	}
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
