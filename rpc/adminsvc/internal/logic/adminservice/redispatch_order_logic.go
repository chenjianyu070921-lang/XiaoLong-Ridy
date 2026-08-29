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

// RedispatchOrderLogic 处理管理后台人工改派订单请求。
// 该逻辑只负责后台入口校验、幂等保护、审计记录和 RPC 转发，订单状态机仍由 ordersvc 统一裁决。
type RedispatchOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewRedispatchOrderLogic 创建后台人工改派逻辑对象。
func NewRedispatchOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RedispatchOrderLogic {
	return &RedispatchOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RedispatchOrder 调用 ordersvc.RedispatchOrder 完成订单人工改派。
// request_id 在 adminsvc 层用于防重复提交；ordersvc 继续负责释放原司机、回到派单池和同步触发派单。
func (l *RedispatchOrderLogic) RedispatchOrder(in *adminsvc.AdminRedispatchOrderRequest) (*adminsvc.AdminRedispatchOrderResponse, error) {
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
		return nil, status.Error(codes.InvalidArgument, "改派原因不能为空")
	}
	if l.svcCtx.OrdersSvc == nil {
		return nil, status.Error(codes.FailedPrecondition, "ordersvc client not ready")
	}
	idempotencyKey := adminOrderActionIdempotencyKey("redispatch_order", in.GetAdminId(), requestID)
	acquired, err := acquireCancelOrderIdempotency(l.ctx, l.svcCtx, idempotencyKey)
	if err != nil {
		return nil, err
	}
	if !acquired {
		return &adminsvc.AdminRedispatchOrderResponse{OrderId: in.GetOrderId(), Message: "ok"}, nil
	}
	resp, err := l.svcCtx.OrdersSvc.RedispatchOrder(l.ctx, &orderproto.RedispatchOrderRequest{
		OrderId:     in.GetOrderId(),
		NewDriverId: in.GetNewDriverId(),
		OperatorId:  in.GetAdminId(),
		Reason:      reason,
	})
	if err != nil {
		releaseCancelOrderIdempotency(l.ctx, l.svcCtx, idempotencyKey)
		return nil, err
	}
	if err := writeAuditAfterCommitted(
		l.ctx,
		l.svcCtx,
		in.GetAdminId(),
		"order",
		"redispatch",
		"ride_order",
		in.GetOrderId(),
		fmt.Sprintf("后台人工改派订单：new_driver_id=%d，reason=%s，request_id=%s", in.GetNewDriverId(), reason, requestID),
		in.GetIp(),
	); err != nil {
		releaseCancelOrderIdempotency(l.ctx, l.svcCtx, idempotencyKey)
		return nil, err
	}
	return &adminsvc.AdminRedispatchOrderResponse{
		OrderId:  resp.GetOrderId(),
		Status:   int32(resp.GetStatus()),
		DriverId: resp.GetDriverId(),
		Message:  "ok",
	}, nil
}
