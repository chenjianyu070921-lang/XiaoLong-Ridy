package logic

import (
	"context"
	"fmt"
	"strconv"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/locationsvc/internal/svc"
	"XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type NearbyDriversLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewNearbyDriversLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NearbyDriversLogic {
	return &NearbyDriversLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// NearbyDrivers 附近司机查询：Redis GEO 半径搜索，按距离升序返回
func (l *NearbyDriversLogic) NearbyDrivers(in *locationsvc.NearbyDriversReq) (*locationsvc.NearbyDriversResp, error) {
	if in.Lat < -90 || in.Lat > 90 || in.Lng < -180 || in.Lng > 180 {
		return nil, fmt.Errorf("经纬度非法: lat=%f lng=%f", in.Lat, in.Lng)
	}
	if in.Radius <= 0 {
		return nil, fmt.Errorf("radius 必须大于 0")
	}
	limit := int(in.Limit)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	res, err := l.svcCtx.Redis.GeoRadius(l.ctx, constants.DriverGeoKeyOf(in.City), in.Lng, in.Lat, &redis.GeoRadiusQuery{
		Radius:    in.Radius,
		Unit:      "m",
		WithDist:  true,
		WithCoord: true,
		Sort:      "ASC",
		Count:     limit,
	}).Result()
	if err != nil {
		l.Errorf("Redis GEO 查询失败: %v", err)
		return nil, err
	}

	resp := &locationsvc.NearbyDriversResp{
		Drivers: make([]*locationsvc.NearbyDriver, 0, len(res)),
	}
	for _, g := range res {
		driverID, err := strconv.ParseInt(g.Name, 10, 64)
		if err != nil {
			l.Errorf("司机ID解析失败: %s", g.Name)
			continue
		}
		resp.Drivers = append(resp.Drivers, &locationsvc.NearbyDriver{
			DriverId: driverID,
			Lng:      g.Longitude,
			Lat:      g.Latitude,
			Distance: g.Dist,
		})
	}

	l.Infof("附近司机查询: city=%s center(%.6f,%.6f) radius=%.0fm 命中 %d 个", in.City, in.Lng, in.Lat, in.Radius, len(resp.Drivers))
	return resp, nil
}
