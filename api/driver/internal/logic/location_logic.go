package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
	locationproto "XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"github.com/zeromicro/go-zero/core/logx"
)

// LocationLogic 封装司机位置上报逻辑，负责校验坐标并调用 driversvc 刷新在线状态与位置。
type LocationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLocationLogic 构造司机位置上报逻辑处理器。
func NewLocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LocationLogic {
	return &LocationLogic{ctx: ctx, svcCtx: svcCtx}
}

// ReportLocation 处理司机实时位置上报。
// driverID 来自 JWT，deviceID 与经纬度来自请求体；下游 driversvc 会执行在线保活、多端互踢判定和位置落库。
func (l *LocationLogic) ReportLocation(driverID int64, req *types.ReportLocationRequest) (*types.ReportLocationResponse, error) {
	if driverID <= 0 || req == nil || strings.TrimSpace(req.DeviceID) == "" ||
		!validLocation(req.Longitude, req.Latitude) {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.ReportLocation(l.ctx, &driversproto.ReportLocationRequest{
		DriverId:  driverID,
		DeviceId:  strings.TrimSpace(req.DeviceID),
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	})
	if err != nil {
		return nil, err
	}
	// 辅助操作失败不阻断主流程：driversvc 已成功更新位置和在线状态，
	// locationsvc / 轨迹记录失败仅告警，避免司机端收到"上报失败"但实际已成功的状态错乱。
	if l.svcCtx != nil && l.svcCtx.LocationClient != nil {
		if _, locErr := l.svcCtx.LocationClient.ReportLocation(l.ctx, &locationproto.ReportLocationReq{
			DriverId:     driverID,
			Lng:          req.Longitude,
			Lat:          req.Latitude,
			Heading:      req.Heading,
			SpeedKmh:     req.SpeedKmh,
			OnlineStatus: resp.GetOnlineStatus(),
			OrderId:      req.OrderID,
		}); locErr != nil {
			logx.WithContext(l.ctx).Errorf("location svc report failed (driver location already updated): %v", locErr)
		}
	}
	if req.OrderID > 0 {
		if l.svcCtx != nil && l.svcCtx.TrajectoryRepository != nil {
			if trajErr := l.svcCtx.TrajectoryRepository.RecordPoint(l.ctx, &svc.TrajectoryRecord{
				OrderID:   req.OrderID,
				DriverID:  driverID,
				Longitude: req.Longitude,
				Latitude:  req.Latitude,
				SpeedKmh:  req.SpeedKmh,
				Heading:   req.Heading,
			}); trajErr != nil {
				logx.WithContext(l.ctx).Errorf("trajectory record failed (driver location already updated): %v", trajErr)
			}
		}
	}
	return &types.ReportLocationResponse{
		DriverID:     resp.GetDriverId(),
		OnlineStatus: int(resp.GetOnlineStatus()),
		Kicked:       resp.GetKicked(),
		ReportTime:   resp.GetReportTime(),
	}, nil
}

// driverClient 从服务上下文中安全取出 driversvc 客户端。
func (l *LocationLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
