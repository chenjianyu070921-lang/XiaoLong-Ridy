package logic

import (
	"context"
	"fmt"
	"time"

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
	if in.OrderId > 0 {
		// 行程中位置带有 order_id 时同步写入轨迹点表，供乘客端轨迹回放和后台客服仲裁使用。
		// 这里只使用既有 ride_track_point 表，不创建或修改表结构。
		point := &model.RideTrackPoint{
			OrderID:    uint64(in.OrderId),
			DriverID:   uint64(in.DriverId),
			Longitude:  in.Lng,
			Latitude:   in.Lat,
			SpeedKmh:   in.SpeedKmh,
			Direction:  int16(in.Heading),
			RecordedAt: loc.ReportTime,
		}
		if err := l.svcCtx.RideTrackPointModel.Insert(point); err != nil {
			l.Errorf("写入 ride_track_point 失败: %v", err)
			return nil, err
		}
	}

	// 2. 写 Redis GEO，供附近司机查询
	if err := l.svcCtx.Redis.GeoAdd(l.ctx, svc.DriverGeoKey, &redis.GeoLocation{
		Name:      fmt.Sprintf("%d", in.DriverId),
		Longitude: in.Lng,
		Latitude:  in.Lat,
	}).Err(); err != nil {
		l.Errorf("写 Redis GEO 失败: %v", err)
		return nil, err
	}

	l.Infof("司机位置上报成功: driverId=%d lng=%.6f lat=%.6f", in.DriverId, in.Lng, in.Lat)
	return &locationsvc.ReportLocationResp{Success: true}, nil
}
