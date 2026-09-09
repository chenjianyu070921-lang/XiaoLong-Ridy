package logic

import (
	"context"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnfreezeDriverLogic 封装后台解冻司机的司机域业务逻辑。
type UnfreezeDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewUnfreezeDriverLogic 创建后台解冻司机逻辑处理器。
func NewUnfreezeDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfreezeDriverLogic {
	return &UnfreezeDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UnfreezeDriver 将冻结状态的司机恢复为正常，仅允许冻结态司机解冻。
// 入参 driver_id 为司机 ID，reason 为解冻原因，operator_id/ip 用于跨服务审计追踪。
// 返回 CommonResponse 表示司机域状态已落库；后台操作日志由 adminsvc 负责记录。
func (l *UnfreezeDriverLogic) UnfreezeDriver(in *proto.UnfreezeDriverRequest) (*proto.CommonResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver id is required")
	}
	if in.GetOperatorId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "operator id is required")
	}
	if strings.TrimSpace(in.GetReason()) == "" {
		return nil, status.Error(codes.InvalidArgument, "unfreeze reason is required")
	}
	driver, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.GetDriverId()))
	if err != nil {
		return nil, err
	}
	if driver == nil || driver.Status != int8(proto.DriverStatus_DRIVER_STATUS_FROZEN) {
		return nil, status.Error(codes.FailedPrecondition, "driver is not frozen")
	}
	if err := l.svcCtx.DriverRepository.Update(l.ctx, uint64(in.GetDriverId()), map[string]interface{}{
		"status":     proto.DriverStatus_DRIVER_STATUS_NORMAL,
		"updated_at": time.Now(),
	}); err != nil {
		return nil, err
	}
	return &proto.CommonResponse{Message: "ok"}, nil
}
