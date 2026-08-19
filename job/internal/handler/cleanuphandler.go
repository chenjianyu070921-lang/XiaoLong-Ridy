package handler

import (
	"context"
	"errors"

	"XiaoLong-Ridy/job/internal/svc"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	order "XiaoLong-Ridy/rpc/ordersvc/orderclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// CleanupHandler 定时清理任务
type CleanupHandler struct {
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCleanupHandler(svcCtx *svc.ServiceContext) *CleanupHandler {
	return &CleanupHandler{
		svcCtx: svcCtx,
		Logger: logx.WithContext(context.Background()),
	}
}

// CleanExpiredLocation 清理过期司机位置数据
func (h *CleanupHandler) CleanExpiredLocation() error {
	h.Info("执行定时任务: 清理过期位置数据")
	return nil
}

// SyncOrderStatus 同步异常订单状态
func (h *CleanupHandler) SyncOrderStatus() error {
	h.Info("执行定时任务: 同步异常订单状态")
	return nil
}

// DailyReport 生成每日统计报表
func (h *CleanupHandler) DailyReport() error {
	h.Info("执行定时任务: 生成每日报表")
	return nil
}

// TimeoutCancelOrders 扫描超时未接单订单并自动取消。
// 分页拉取 ListTimeoutOrders，逐个调用 TimeoutCancel；单条取消失败不阻断，记录日志后继续。
func (h *CleanupHandler) TimeoutCancelOrders() error {
	timeoutSeconds := h.svcCtx.Config.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 300
	}
	page, pageSize := 1, 50
	cancelled := 0
	for {
		resp, err := h.svcCtx.OrderClient.ListTimeoutOrders(context.Background(), &order.ListTimeoutOrdersRequest{
			TimeoutSeconds: int32(timeoutSeconds),
			Page:           int32(page),
			PageSize:       int32(pageSize),
		})
		if err != nil {
			return err
		}
		if resp == nil || len(resp.List) == 0 {
			break
		}
		for _, item := range resp.List {
			if _, err := h.svcCtx.OrderClient.TimeoutCancel(context.Background(), &order.TimeoutCancelRequest{
				OrderId: item.OrderId,
				Reason:  "超时未接单，系统自动取消",
			}); err != nil {
				h.Errorf("取消超时订单失败 orderId=%d: %v", item.OrderId, err)
				continue
			}
			cancelled++
			h.Infof("已取消超时订单 orderId=%d orderNo=%s", item.OrderId, item.OrderNo)
		}
		if len(resp.List) < pageSize {
			break
		}
		page++
	}
	h.Infof("超时取消任务完成: 本轮取消 %d 单", cancelled)
	return nil
}

// RescheduleExpiredDispatches 扫描派单超时的订单并触发重派。
// 分页拉取 ListTimeoutPendingOrders，逐个查订单详情并调用 DispatchOrder
// （其内部幂等：将超时记录置 Timeout 后重新匹配派单）；单单失败不阻断，记日志后继续。
func (h *CleanupHandler) RescheduleExpiredDispatches() error {
	timeoutSeconds := h.svcCtx.Config.DispatchTimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60
	}
	page, pageSize := 1, 50
	redispatched := 0
	for {
		resp, err := h.svcCtx.DispatchClient.ListTimeoutPendingOrders(context.Background(), &dispatch.ListTimeoutPendingOrdersRequest{
			TimeoutSeconds: timeoutSeconds,
			Page:           int32(page),
			PageSize:       int32(pageSize),
		})
		if err != nil {
			return err
		}
		if resp == nil || len(resp.OrderIds) == 0 {
			break
		}
		for _, orderID := range resp.OrderIds {
			if err := h.redispatchOrder(orderID); err != nil {
				h.Errorf("重派订单失败 orderId=%d: %v", orderID, err)
				continue
			}
			redispatched++
			h.Infof("已触发重派 orderId=%d", orderID)
		}
		if int64(len(resp.OrderIds)) < int64(resp.PageSize) {
			break
		}
		page++
	}
	h.Infof("派单超时重派任务完成: 本轮重派 %d 单", redispatched)
	return nil
}

// redispatchOrder 拉取订单详情并触发派单（幂等重派）。
func (h *CleanupHandler) redispatchOrder(orderID int64) error {
	orderInfo, err := h.svcCtx.OrderClient.GetOrder(context.Background(), &order.GetOrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		return err
	}
	if orderInfo == nil || orderInfo.OrderId <= 0 {
		return errors.New("order not found")
	}
	_, err = h.svcCtx.DispatchClient.DispatchOrder(context.Background(), &dispatch.DispatchOrderRequest{
		OrderId:       orderInfo.OrderId,
		FromLongitude: orderInfo.FromLongitude,
		FromLatitude:  orderInfo.FromLatitude,
		CarType:       orderInfo.CarType,
		CityCode:      "", // GetOrder 未返回城市编码，交由派单引擎兜底
	})
	return err
}
