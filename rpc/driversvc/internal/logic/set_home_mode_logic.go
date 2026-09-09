package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SetHomeModeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetHomeModeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetHomeModeLogic {
	return &SetHomeModeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetHomeMode 仅切换回家顺路模式开关：
// 开启 = 派单与大厅推单做顺路过滤；关闭 = 恢复全域听单。
// 未设置过目的地的司机不允许开启（避免无方向可判顺路）。
func (l *SetHomeModeLogic) SetHomeMode(in *proto.SetHomeModeRequest) (*proto.HomeDestinationResponse, error) {
	if in == nil || in.DriverId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver_id required")
	}
	if l.svcCtx == nil || l.svcCtx.DriverHomeDestinationRepository == nil {
		return nil, status.Error(codes.Internal, "home destination repository not ready")
	}
	repo := l.svcCtx.DriverHomeDestinationRepository
	current, err := repo.GetByDriver(l.ctx, uint64(in.DriverId))
	if err != nil {
		l.Errorf("load home destination failed: driverId=%d err=%v", in.DriverId, err)
		return nil, status.Error(codes.Internal, "load home destination failed")
	}
	if current == nil {
		return nil, status.Error(codes.FailedPrecondition, "home destination not set")
	}
	if in.Open && !validLongitudeLatitude(current.HomeLng, current.HomeLat) {
		return nil, status.Error(codes.FailedPrecondition, "home destination coordinate invalid")
	}
	updated, err := repo.SetModeOpen(l.ctx, uint64(in.DriverId), in.Open)
	if err != nil {
		l.Errorf("set home mode failed: driverId=%d open=%v err=%v", in.DriverId, in.Open, err)
		return nil, status.Error(codes.Internal, "set home mode failed")
	}
	if !updated {
		return nil, status.Error(codes.FailedPrecondition, "home destination not set")
	}
	current.IsHomeModeOpen = boolToInt8(in.Open)
	syncHomeDestinationCache(l.ctx, l.svcCtx, current)
	return homeDestinationResponse(current), nil
}

func boolToInt8(open bool) int8 {
	if open {
		return 1
	}
	return 0
}

// validHomeModeTransition 保留给上层校验使用：关闭模式永远允许，开启要求坐标有效。
func validHomeModeTransition(setting *model.DriverHomeDestination, open bool) bool {
	if !open {
		return true
	}
	return setting != nil && validLongitudeLatitude(setting.HomeLng, setting.HomeLat)
}
