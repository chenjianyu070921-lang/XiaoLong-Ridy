package logic

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/redisx"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// orderLockTTL 接单/取消分布式锁过期时间，配合看门狗续期防止持有锁的服务异常退出导致锁永久占用。
const orderLockTTL = 10 * time.Second

// acquireOrderLock 为订单加分布式锁，避免同一订单被并发接单/取消竞态。
//
// 修复说明（司机侧缺陷#2）：Redis 不可用（含未配置）直接拒绝接单/取消，不再降级为单进程内存锁。
// 多实例部署下本地锁互不生效，会导致一单多派；Redis 必须作为订单并发安全的硬依赖。
func acquireOrderLock(ctx context.Context, rdb *redis.Client, orderID uint64) (func(), error) {
	noop := func() {}
	if rdb == nil {
		return noop, fmt.Errorf("redis unavailable: order %d lock rejected to avoid multi-instance race", orderID)
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
