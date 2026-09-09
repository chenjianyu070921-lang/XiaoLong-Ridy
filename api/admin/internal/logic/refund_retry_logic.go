package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// RefundRetryLogic 负责退款补偿任务的 HTTP 到 RPC 参数转换。
type RefundRetryLogic struct {
	ctx *svc.ServiceContext
}

// NewRefundRetryLogic 创建退款补偿逻辑对象。
func NewRefundRetryLogic(ctx *svc.ServiceContext) *RefundRetryLogic {
	return &RefundRetryLogic{ctx: ctx}
}

// List 查询退款补偿任务。
func (l *RefundRetryLogic) List(ctx context.Context, page, pageSize int32) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListRefundRetryTasks(ctx, &adminclient.RefundRetryTaskListRequest{Page: page, PageSize: pageSize})
	if err != nil {
		return nil, err
	}
	items := make([]types.RefundRetryTask, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, types.RefundRetryTask{
			OrderID: item.OrderId, OrderNo: item.OrderNo, RefundNo: item.RefundNo,
			RefundCents: item.RefundCents, OperatorType: item.OperatorType,
			OperatorID: item.OperatorId, Attempt: item.Attempt, NextRetryAt: item.NextRetryAt,
		})
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// Retry 触发指定退款补偿任务。
func (l *RefundRetryLogic) Retry(ctx context.Context, refundNo string, adminID int64, ip string) error {
	_, err := l.ctx.AdminSvc.RetryRefundTask(ctx, &adminclient.RefundRetryTaskRequest{
		RefundNo: refundNo, AdminId: adminID, Ip: ip,
	})
	return err
}
