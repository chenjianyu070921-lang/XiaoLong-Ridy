package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CancelOrderLogic 处理后台取消订单 RPC。
type CancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewCancelOrderLogic 创建后台取消订单逻辑对象。
func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CancelOrder 通过 ordersvc 取消订单，并返回统一结果。
func (l *CancelOrderLogic) CancelOrder(in *adminsvc.AdminCancelOrderRequest) (*adminsvc.CommonResponse, error) {
	if in.GetOrderId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "订单ID不能为空")
	}
	if in.GetAdminId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "管理员ID不能为空")
	}
	reason := strings.TrimSpace(in.GetReason())
	if reason == "" {
		reason = "后台取消订单"
	}
	if l.svcCtx.OrdersSvc == nil {
		return nil, status.Error(codes.FailedPrecondition, "ordersvc client not ready")
	}
	_, err := l.svcCtx.OrdersSvc.CancelOrder(l.ctx, &orderproto.CancelOrderRequest{
		OrderId:      in.GetOrderId(),
		OperatorType: "admin",
		OperatorId:   in.GetAdminId(),
		Reason:       reason,
	})
	if err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}
