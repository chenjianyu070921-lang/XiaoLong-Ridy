package logic

import (
	"context"
	"fmt"
	"sort"
	"strconv"

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

	// 司机端当前统一写入 default 城市 GEO key；同时查询配置城市 key，兼容历史数据和多城市配置。
	keys := []string{svc.GeoKey("default")}
	configuredKey := svc.GeoKey(l.svcCtx.GetConfig().DefaultCityCode)
	if configuredKey != keys[0] {
		keys = append(keys, configuredKey)
	}
	query := &redis.GeoRadiusQuery{
		Radius:    in.Radius,
		Unit:      "m",
		WithDist:  true,
		WithCoord: true,
		Sort:      "ASC",
		Count:     limit,
	}
	// 汇总多个 key 后按距离排序，避免同一司机因兼容 key 重复出现在结果中。
	seen := make(map[string]bool)
	res := make([]redis.GeoLocation, 0, limit)
	for _, key := range keys {
		items, queryErr := l.svcCtx.Redis.GeoRadius(l.ctx, key, in.Lng, in.Lat, query).Result()
		if queryErr != nil {
			l.Errorf("Redis GEO 查询失败: key=%s err=%v", key, queryErr)
			continue
		}
		for _, item := range items {
			if !seen[item.Name] {
				seen[item.Name] = true
				res = append(res, item)
			}
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Dist < res[j].Dist })
	if len(res) > limit {
		res = res[:limit]
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

	l.Infof("附近司机查询: center(%.6f,%.6f) radius=%.0fm 命中 %d 个", in.Lng, in.Lat, in.Radius, len(resp.Drivers))
	return resp, nil
}
