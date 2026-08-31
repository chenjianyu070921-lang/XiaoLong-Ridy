package adminservicelogic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// refundRetryTask 是 ordersvc 写入 Redis 的退款事件补偿任务载荷。
type refundRetryTask struct {
	OrderID      int64  `json:"order_id"`
	OrderNo      string `json:"order_no"`
	RefundNo     string `json:"refund_no"`
	RefundCents  int64  `json:"refund_cents"`
	OperatorType string `json:"operator_type"`
	OperatorID   int64  `json:"operator_id"`
	Attempt      int    `json:"attempt"`
}

// RefundRetryLogic 提供退款补偿任务查询和人工触发能力。
type RefundRetryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRefundRetryLogic 创建退款补偿逻辑对象。
func NewRefundRetryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundRetryLogic {
	return &RefundRetryLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListRefundRetryTasks 查询 Redis 中的退款补偿任务，按下一次重试时间升序返回。
func (l *RefundRetryLogic) ListRefundRetryTasks(in *adminsvc.RefundRetryTaskListRequest) (*adminsvc.RefundRetryTaskListResponse, error) {
	if l.svcCtx.Redis == nil {
		return nil, status.Error(codes.FailedPrecondition, "退款补偿队列未配置")
	}
	page, pageSize := normalizeRefundPage(in.GetPage(), in.GetPageSize())
	members, err := l.svcCtx.Redis.ZRangeWithScores(l.ctx, constants.RefundRetryQueueKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("查询退款补偿队列失败: %w", err)
	}
	start := int((page - 1) * pageSize)
	end := start + int(pageSize)
	if start > len(members) {
		start = len(members)
	}
	if end > len(members) {
		end = len(members)
	}
	items := make([]*adminsvc.RefundRetryTask, 0, end-start)
	for _, member := range members[start:end] {
		var task refundRetryTask
		payload, ok := member.Member.(string)
		if !ok {
			continue
		}
		if err := json.Unmarshal([]byte(payload), &task); err != nil {
			continue
		}
		items = append(items, &adminsvc.RefundRetryTask{
			OrderId: task.OrderID, OrderNo: task.OrderNo, RefundNo: task.RefundNo,
			RefundCents: task.RefundCents, OperatorType: task.OperatorType,
			OperatorId: task.OperatorID, Attempt: int32(task.Attempt),
			NextRetryAt: int64(member.Score),
		})
	}
	return &adminsvc.RefundRetryTaskListResponse{
		List: items, Total: int64(len(members)), Page: page, PageSize: pageSize,
	}, nil
}

// RetryRefundTask 将指定退款补偿任务提前到当前时间，由 job 下一轮立即投递。
func (l *RefundRetryLogic) RetryRefundTask(in *adminsvc.RefundRetryTaskRequest) (*adminsvc.CommonResponse, error) {
	refundNo := strings.TrimSpace(in.GetRefundNo())
	if refundNo == "" {
		return nil, status.Error(codes.InvalidArgument, "退款单号不能为空")
	}
	if l.svcCtx.Redis == nil {
		return nil, status.Error(codes.FailedPrecondition, "退款补偿队列未配置")
	}
	members, err := l.svcCtx.Redis.ZRangeWithScores(l.ctx, constants.RefundRetryQueueKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("查询退款补偿任务失败: %w", err)
	}
	for _, member := range members {
		payload, ok := member.Member.(string)
		if !ok {
			continue
		}
		var task refundRetryTask
		if json.Unmarshal([]byte(payload), &task) == nil && task.RefundNo == refundNo {
			if err := l.svcCtx.Redis.ZAdd(l.ctx, constants.RefundRetryQueueKey, redis.Z{
				Score: float64(time.Now().Unix()), Member: payload,
			}).Err(); err != nil {
				return nil, fmt.Errorf("触发退款补偿重试失败: %w", err)
			}
			return &adminsvc.CommonResponse{Message: "ok"}, nil
		}
	}
	return nil, status.Error(codes.NotFound, "退款补偿任务不存在")
}

// normalizeRefundPage 统一限制分页参数，避免一次读取过多补偿任务。
func normalizeRefundPage(page, pageSize int32) (int32, int32) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
