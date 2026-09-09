package logic

import (
	"context"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// acceptedTimeoutThreshold 司机接单后未开始行程的超时阈值；超过则自动取消并释放运力。
	acceptedTimeoutThreshold = 5 * time.Minute
	// acceptedTimeoutScanPageSize 已接单超时扫描每页条数。
	acceptedTimeoutScanPageSize = 50
)

// TimeoutAcceptLogic 回收“已接单但未开始行程”的超时订单，释放被长期占用的司机运力。
type TimeoutAcceptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewTimeoutAcceptLogic 创建已接单超时回收逻辑对象。
func NewTimeoutAcceptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TimeoutAcceptLogic {
	return &TimeoutAcceptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Run 扫描并回收超时未开始行程的已接单订单；单条失败不阻断整轮扫描。
func (l *TimeoutAcceptLogic) Run() {
	before := time.Now().Add(-acceptedTimeoutThreshold)
	page := 1
	processed := 0
	for {
		orders, _, err := l.svcCtx.OrderRepository.ListAcceptedTimeoutOrders(l.ctx, before, int32(page), acceptedTimeoutScanPageSize)
		if err != nil {
			l.Errorf("扫描已接单超时订单失败: %v", err)
			return
		}
		if len(orders) == 0 {
			break
		}
		for _, o := range orders {
			if err := l.processOne(o); err != nil {
				l.Errorf("回收已接单超时订单失败 orderId=%d: %v", o.Id, err)
				continue
			}
			processed++
		}
		if len(orders) < acceptedTimeoutScanPageSize {
			break
		}
		page++
	}
	if processed > 0 {
		l.Infof("已接单超时回收完成: 本轮回收 %d 单", processed)
	}
}

// processOne 对单笔已接单订单加锁后执行超时取消，并释放司机忙碌状态、失效待派单记录。
func (l *TimeoutAcceptLogic) processOne(o model.RideOrder) error {
	release, err := acquireOrderLock(l.ctx, l.svcCtx.Redis, o.Id)
	if err != nil {
		return err
	}
	defer release()

	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, o.Id)
	if err != nil {
		return err
	}
	if order.Status != constants.OrderStatusAccepted || order.DriverId == 0 {
		return nil
	}

	reason := "司机接单后超时未开始行程，系统自动取消"
	statusLog := &model.OrderStatusLog{
		FromStatus:   order.Status,
		ToStatus:     constants.OrderStatusCancelled,
		OperatorType: constants.OperatorSystem,
		OperatorId:   0,
		Remark:       reason,
	}
	ok, err := l.svcCtx.OrderRepository.TimeoutAccept(l.ctx, order.Id, time.Now().Add(-acceptedTimeoutThreshold), reason, statusLog)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	// 释放司机忙碌状态，恢复可接单（失败不阻断主流程）。
	if order.DriverId > 0 {
		unmarkDriverBusy(l.ctx, l.svcCtx, order.DriverId)
	}
	// 同步失效该订单的待派单记录，避免被重派任务重复处理。
	syncCancelDispatch(l.ctx, l.svcCtx.DispatchClient, order.Id, reason)
	return nil
}
