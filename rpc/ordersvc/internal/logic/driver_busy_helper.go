package logic

import (
	"context"
	"strconv"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// markDriverBusy 将司机写入忙碌集合（driver:busy），派单引擎据此过滤不可接单司机（P1-M4-8）。
// 在司机接单成功后调用；Redis 未配置或写入失败仅记录错误，不阻断订单主流程。
func markDriverBusy(ctx context.Context, svcCtx *svc.ServiceContext, driverID uint64) {
	if svcCtx == nil || svcCtx.Redis == nil || driverID == 0 {
		return
	}
	if err := svcCtx.Redis.SAdd(ctx, constants.RedisDriverBusy, strconv.FormatUint(driverID, 10)).Err(); err != nil {
		logx.WithContext(ctx).Errorf("mark driver %d busy failed: %v", driverID, err)
	}
}

// unmarkDriverBusy 将司机移出忙碌集合（driver:busy），行程结束或订单取消时调用（P1-M4-8）。
// 幂等操作；Redis 不可用时仅记录错误，下一笔订单接单后会被重新置忙，不会造成永久脏数据。
func unmarkDriverBusy(ctx context.Context, svcCtx *svc.ServiceContext, driverID uint64) {
	if svcCtx == nil || svcCtx.Redis == nil || driverID == 0 {
		return
	}
	if err := svcCtx.Redis.SRem(ctx, constants.RedisDriverBusy, strconv.FormatUint(driverID, 10)).Err(); err != nil {
		logx.WithContext(ctx).Errorf("unmark driver %d busy failed: %v", driverID, err)
	}
}
