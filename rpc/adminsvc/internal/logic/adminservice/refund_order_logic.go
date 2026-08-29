package adminservicelogic

import (
	"context"
	"fmt"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RefundOrderLogic 处理管理后台订单退款请求。
// 资金类动作仅暴露后台受控入口，实际订单状态、退款金额边界和优惠券回滚由 ordersvc 强制退款状态机处理。
type RefundOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewRefundOrderLogic 创建后台订单退款逻辑对象。
func NewRefundOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundOrderLogic {
	return &RefundOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RefundOrder 调用 ordersvc.ForceRefundOrder 发起后台退款。
// request_id 同时作为后台幂等号和 ordersvc 的 refund_no，保证审计、订单状态和后续补偿事件可按同一业务号追踪。
func (l *RefundOrderLogic) RefundOrder(in *adminsvc.AdminRefundOrderRequest) (*adminsvc.AdminRefundOrderResponse, error) {
	if in.GetOrderId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "订单ID不能为空")
	}
	if in.GetAdminId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "管理员ID不能为空")
	}
	requestID := strings.TrimSpace(in.GetRequestId())
	if requestID == "" {
		return nil, status.Error(codes.InvalidArgument, "request_id不能为空")
	}
	reason := strings.TrimSpace(in.GetReason())
	if reason == "" {
		return nil, status.Error(codes.InvalidArgument, "退款原因不能为空")
	}
	if l.svcCtx.OrdersSvc == nil {
		return nil, status.Error(codes.FailedPrecondition, "ordersvc client not ready")
	}
	resp, err := l.svcCtx.OrdersSvc.ForceRefundOrder(l.ctx, &orderproto.ForceRefundOrderRequest{
		OrderId:           in.GetOrderId(),
		OperatorId:        in.GetAdminId(),
		RefundNo:          requestID,
		RefundAmountCents: in.GetRefundAmountCents(),
		Reason:            reason,
	})
	if err != nil {
		return nil, err
	}
	if err := writeAuditAfterCommitted(
		l.ctx,
		l.svcCtx,
		in.GetAdminId(),
		"order",
		"refund",
		"ride_order",
		in.GetOrderId(),
		fmt.Sprintf("后台订单退款：refund_no=%s，refund_amount_cents=%d，reason=%s", requestID, in.GetRefundAmountCents(), reason),
		in.GetIp(),
	); err != nil {
		return nil, err
	}
	return &adminsvc.AdminRefundOrderResponse{
		OrderId:     resp.GetOrderId(),
		Status:      int32(resp.GetStatus()),
		RefundCents: resp.GetRefundCents(),
		RefundNo:    requestID,
		Message:     "ok",
	}, nil
}
