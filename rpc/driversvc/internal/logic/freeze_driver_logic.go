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

// FreezeDriverLogic 封装后台冻结司机的司机域业务逻辑。
type FreezeDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewFreezeDriverLogic 创建后台冻结司机逻辑处理器。
func NewFreezeDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FreezeDriverLogic {
	return &FreezeDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FreezeDriver 将司机状态置为冻结，并同步 Redis 在线状态为离线。
// 入参 driver_id 为司机 ID，reason 为冻结原因，operator_id/ip 用于跨服务审计追踪。
// 返回 CommonResponse 表示司机域状态已落库；后台操作日志由 adminsvc 负责记录。
func (l *FreezeDriverLogic) FreezeDriver(in *proto.FreezeDriverRequest) (*proto.CommonResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver id is required")
	}
	if in.GetOperatorId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "operator id is required")
	}
	if strings.TrimSpace(in.GetReason()) == "" {
		return nil, status.Error(codes.InvalidArgument, "freeze reason is required")
	}
	if _, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.GetDriverId())); err != nil {
		return nil, err
	}
	if err := l.svcCtx.DriverRepository.Update(l.ctx, uint64(in.GetDriverId()), map[string]interface{}{
		"status":     proto.DriverStatus_DRIVER_STATUS_FROZEN,
		"updated_at": time.Now(),
	}); err != nil {
		return nil, err
	}
	if l.svcCtx.OnlineStore != nil {
		if err := l.svcCtx.OnlineStore.SetOffline(l.ctx, in.GetDriverId()); err != nil {
			return nil, err
		}
	}
	if err := l.svcCtx.DriverRepository.UpdateLocationStatus(l.ctx, uint64(in.GetDriverId()), 0); err != nil {
		return nil, err
	}
	return &proto.CommonResponse{Message: "ok"}, nil
}
