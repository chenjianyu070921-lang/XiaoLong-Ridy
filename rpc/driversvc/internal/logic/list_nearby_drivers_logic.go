package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListNearbyDriversLogic 附近司机查询逻辑。
type ListNearbyDriversLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListNearbyDriversLogic 构造附近司机查询逻辑处理器。
func NewListNearbyDriversLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNearbyDriversLogic {
	return &ListNearbyDriversLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListNearbyDrivers 按经纬度 + 半径查找在线司机，返回按距离升序的列表。
func (l *ListNearbyDriversLogic) ListNearbyDrivers(in *proto.ListNearbyDriversRequest) (*proto.ListNearbyDriversResponse, error) {
	if in == nil || !validLongitudeLatitude(in.GetLongitude(), in.GetLatitude()) || in.GetRadiusMeters() < 0 || in.GetLimit() < 0 {
		return nil, errInvalidLongitudeLatitude
	}
	filter := repository.NearbyDriverFilter{
		Longitude:    in.GetLongitude(),
		Latitude:     in.GetLatitude(),
		RadiusMeters: in.GetRadiusMeters(),
		Limit:        int(in.GetLimit()),
	}

	locations, err := l.svcCtx.DriverRepository.ListNearbyDrivers(l.ctx, filter)
	if err != nil {
		return nil, err
	}

	list := make([]*proto.NearbyDriver, 0, len(locations))
	for _, loc := range locations {
		list = append(list, &proto.NearbyDriver{
			DriverId:       int64(loc.DriverID),
			Longitude:      loc.Longitude,
			Latitude:       loc.Latitude,
			DistanceMeters: int32(loc.DistanceMeters),
		})
	}
	return &proto.ListNearbyDriversResponse{Drivers: list}, nil
}
