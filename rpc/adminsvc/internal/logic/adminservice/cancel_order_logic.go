package adminservicelogic

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

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

// CancelOrder 调用 ordersvc.CancelOrder 完成后台取消订单。
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
	idempotencyKey := cancelOrderIdempotencyKey(in, reason)
	acquired, err := acquireCancelOrderIdempotency(l.ctx, l.svcCtx, idempotencyKey)
	if err != nil {
		return nil, err
	}
	if !acquired {
		return &adminsvc.CommonResponse{Message: "ok"}, nil
	}
	_, err = l.svcCtx.OrdersSvc.CancelOrder(l.ctx, &orderproto.CancelOrderRequest{
		OrderId:      in.GetOrderId(),
		OperatorType: "admin",
		OperatorId:   in.GetAdminId(),
		Reason:       reason,
	})
	if err != nil {
		releaseCancelOrderIdempotency(l.ctx, l.svcCtx, idempotencyKey)
		// ordersvc 目前部分业务错误仍以普通 error 返回，经 gRPC 透传后会变成 Unknown。
		// 后台 HTTP 层依赖 gRPC code 做统一响应映射，因此这里把明确可识别的订单不存在错误转换为 NotFound。
		if strings.Contains(err.Error(), "orderclient not found") {
			return nil, status.Error(codes.NotFound, "订单不存在")
		}
		return nil, err
	}
	// ordersvc 已完成跨服务订单状态变更，adminsvc 在本地补写后台审计日志。
	// 审计日志写入失败时由公共 helper 创建 outbox 补偿任务，避免跨服务成功但审计事件丢失。
	if err := writeAuditAfterCommitted(
		l.ctx,
		l.svcCtx,
		in.GetAdminId(),
		"order",
		"cancel",
		"ride_order",
		in.GetOrderId(),
		fmt.Sprintf("后台取消订单：%s", reason),
		in.GetIp(),
	); err != nil {
		releaseCancelOrderIdempotency(l.ctx, l.svcCtx, idempotencyKey)
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// cancelOrderIdempotencyKey 生成后台取消订单幂等键。
// 优先使用调用方传入的 request_id；旧调用方未传时用订单、管理员和原因生成兼容键，避免重复点击造成二次调用。
func cancelOrderIdempotencyKey(in *adminsvc.AdminCancelOrderRequest, reason string) string {
	raw := strings.TrimSpace(in.GetRequestId())
	if raw == "" {
		raw = fmt.Sprintf("order:%d:admin:%d:reason:%s", in.GetOrderId(), in.GetAdminId(), reason)
	}
	return adminOrderActionIdempotencyKey("cancel_order", in.GetAdminId(), raw)
}

// adminOrderActionIdempotencyKey 生成后台订单写操作的统一幂等键。
// action 用于隔离取消、改派、退款等不同动作；requestID 由 HTTP 层生成或调用方传入，避免重复点击重复触发下游 RPC。
func adminOrderActionIdempotencyKey(action string, adminID int64, requestID string) string {
	raw := strings.TrimSpace(requestID)
	sum := sha1.Sum([]byte(raw))
	return fmt.Sprintf("admin:idem:%s:%d:%s", action, adminID, hex.EncodeToString(sum[:]))
}

// acquireCancelOrderIdempotency 尝试登记取消订单幂等请求。
// 返回 false 表示同一个 request_id 已经处理或正在处理，上游可直接返回幂等成功。
func acquireCancelOrderIdempotency(ctx context.Context, svcCtx *svc.ServiceContext, key string) (bool, error) {
	if svcCtx == nil || svcCtx.Redis == nil {
		return true, nil
	}
	ok, err := svcCtx.Redis.SetNX(ctx, key, "processing", 24*time.Hour).Result()
	if err != nil {
		return false, status.Error(codes.FailedPrecondition, "幂等校验失败")
	}
	return ok, nil
}

// releaseCancelOrderIdempotency 在下游失败时释放幂等键，允许管理员按同一 request_id 重试。
func releaseCancelOrderIdempotency(ctx context.Context, svcCtx *svc.ServiceContext, key string) {
	if svcCtx == nil || svcCtx.Redis == nil || key == "" {
		return
	}
	_ = svcCtx.Redis.Del(ctx, key).Err()
}
