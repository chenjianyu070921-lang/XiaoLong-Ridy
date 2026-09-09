package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetHomeDestinationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHomeDestinationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHomeDestinationLogic {
	return &GetHomeDestinationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetHomeDestination 查询司机回家目的地与回家模式状态；未设置过返回 has_setting=false。
func (l *GetHomeDestinationLogic) GetHomeDestination(in *proto.GetHomeDestinationRequest) (*proto.HomeDestinationResponse, error) {
	if in == nil || in.DriverId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver_id required")
	}
	if l.svcCtx == nil || l.svcCtx.DriverHomeDestinationRepository == nil {
		return nil, status.Error(codes.Internal, "home destination repository not ready")
	}
	setting, err := l.svcCtx.DriverHomeDestinationRepository.GetByDriver(l.ctx, uint64(in.DriverId))
	if err != nil {
		l.Errorf("get home destination failed: driverId=%d err=%v", in.DriverId, err)
		return nil, status.Error(codes.Internal, "get home destination failed")
	}
	return homeDestinationResponse(setting), nil
}
