package logic

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/rpc/locationsvc/internal/svc"
	"XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// defaultOrderTrackLimit 是后台轨迹查询默认返回点数，避免一次性拉取过多轨迹点。
	defaultOrderTrackLimit = 1000
	// maxOrderTrackLimit 是单次轨迹查询点数上限，保护数据库和 RPC 响应体大小。
	maxOrderTrackLimit = 5000
)

// GetOrderTrackLogic 处理订单轨迹查询。
type GetOrderTrackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetOrderTrackLogic 创建订单轨迹查询逻辑对象。
func NewGetOrderTrackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderTrackLogic {
	return &GetOrderTrackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetOrderTrack 查询指定订单的轨迹点。
// 该接口只读取 ride_track_point 既有表，不做任何状态修改，适合管理后台客服查看轨迹证据。
func (l *GetOrderTrackLogic) GetOrderTrack(in *locationsvc.GetOrderTrackReq) (*locationsvc.GetOrderTrackResp, error) {
	if in.GetOrderId() <= 0 {
		return nil, fmt.Errorf("order_id 非法: %d", in.GetOrderId())
	}
	startAt, endAt := unixRange(in.GetStartTime(), in.GetEndTime())
	limit := normalizeOrderTrackLimit(in.GetLimit())
	points, err := l.svcCtx.RideTrackPointModel.ListByOrder(in.GetOrderId(), startAt, endAt, limit)
	if err != nil {
		l.Errorf("查询订单轨迹失败: orderId=%d err=%v", in.GetOrderId(), err)
		return nil, err
	}
	resp := &locationsvc.GetOrderTrackResp{Points: make([]*locationsvc.TrackPoint, 0, len(points))}
	for _, point := range points {
		resp.Points = append(resp.Points, &locationsvc.TrackPoint{
			Id:         int64(point.ID),
			OrderId:    int64(point.OrderID),
			DriverId:   int64(point.DriverID),
			Lng:        point.Longitude,
			Lat:        point.Latitude,
			SpeedKmh:   point.SpeedKmh,
			Direction:  int32(point.Direction),
			RecordedAt: point.RecordedAt.Unix(),
		})
	}
	return resp, nil
}

// normalizeOrderTrackLimit 归一化轨迹点查询上限。
func normalizeOrderTrackLimit(limit int32) int {
	if limit <= 0 {
		return defaultOrderTrackLimit
	}
	if limit > maxOrderTrackLimit {
		return maxOrderTrackLimit
	}
	return int(limit)
}

// unixRange 将可选 Unix 秒时间范围转换为 time.Time，非法或缺省值按不限处理。
func unixRange(start, end int64) (time.Time, time.Time) {
	var startAt, endAt time.Time
	if start > 0 {
		startAt = time.Unix(start, 0)
	}
	if end > 0 {
		endAt = time.Unix(end, 0)
	}
	return startAt, endAt
}
