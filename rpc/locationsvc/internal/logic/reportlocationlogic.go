package logic

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/locationsvc/internal/model"
	"XiaoLong-Ridy/rpc/locationsvc/internal/svc"
	"XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type ReportLocationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReportLocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportLocationLogic {
	return &ReportLocationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReportLocation 司机位置上报：落库 driver_location + 写 Redis GEO（供附近司机查询）
func (l *ReportLocationLogic) ReportLocation(in *locationsvc.ReportLocationReq) (*locationsvc.ReportLocationResp, error) {
	if in.DriverId <= 0 {
		return nil, fmt.Errorf("driver_id 非法: %d", in.DriverId)
	}
	if in.Lat < -90 || in.Lat > 90 || in.Lng < -180 || in.Lng > 180 {
		return nil, fmt.Errorf("经纬度非法: lat=%f lng=%f", in.Lat, in.Lng)
	}

	loc := &model.DriverLocation{
		DriverID:     uint64(in.DriverId),
		Longitude:    in.Lng,
		Latitude:     in.Lat,
		Heading:      int16(in.Heading),
		SpeedKmh:     in.SpeedKmh,
		OnlineStatus: int8(in.OnlineStatus),
		ReportTime:   time.Now(),
	}

	// 1. 落库（driver_id 唯一，冲突则更新最新位置）
	if err := l.svcCtx.DriverLocationModel.Upsert(loc); err != nil {
		l.Errorf("写入 driver_location 失败: %v", err)
		return nil, err
	}

	geoKey := constants.DriverGeoKeyOf(in.City)

	// 2. 写 Redis GEO（按城市分桶），供附近司机查询
	if err := l.svcCtx.Redis.GeoAdd(l.ctx, geoKey, &redis.GeoLocation{
		Name:      fmt.Sprintf("%d", in.DriverId),
		Longitude: in.Lng,
		Latitude:  in.Lat,
	}).Err(); err != nil {
		l.Errorf("写 Redis GEO 失败: %v", err)
		return nil, err
	}

	// 3. 发布位置事件到 Redis Stream，供 location-consumer 消费（在线状态维护、离线清理等）
	if err := l.svcCtx.Redis.XAdd(l.ctx, &redis.XAddArgs{
		Stream: constants.LocationStreamKey,
		Values: map[string]interface{}{
			"driver_id":     fmt.Sprintf("%d", in.DriverId),
			"lng":           in.Lng,
			"lat":           in.Lat,
			"online_status": in.OnlineStatus,
			"city":          in.City,
			"ts":            time.Now().Unix(),
		},
	}).Err(); err != nil {
		// 位置流失败不影响主链路（GEO 已写入），仅告警
		l.Errorf("发布位置事件到 Stream 失败: %v", err)
	}

	l.Infof("司机位置上报成功: driverId=%d city=%s lng=%.6f lat=%.6f", in.DriverId, in.City, in.Lng, in.Lat)
	return &locationsvc.ReportLocationResp{Success: true}, nil
}
