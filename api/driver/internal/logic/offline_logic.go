package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

// OfflineLogic 封装司机下线逻辑，持有请求上下文与下游 driversvc 客户端。
type OfflineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewOfflineLogic 构造司机下线逻辑处理器，注入请求上下文与服务上下文。
func NewOfflineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OfflineLogic {
	return &OfflineLogic{ctx: ctx, svcCtx: svcCtx}
}

// SetOffline 将当前登录司机置为离线。driverID 由鉴权中间件从 JWT 解析得到，
// deviceID 为当前设备标识（用于多端互踢），longitude/latitude 为上报位置（可选），透传给 driversvc。
func (l *OfflineLogic) SetOffline(driverID int64, deviceID string, longitude, latitude float64) (*types.SetOfflineResponse, error) {
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.SetDriverOffline(l.ctx, &driversproto.SetDriverOfflineRequest{
		DriverId:  driverID,
		DeviceId:  deviceID,
		Longitude: longitude,
		Latitude:  latitude,
	})
	if err != nil {
		return nil, err
	}
	return &types.SetOfflineResponse{
		DriverID:     resp.GetDriverId(),
		OnlineStatus: int(resp.GetOnlineStatus()),
		Kicked:       resp.GetKicked(),
	}, nil
}

// driverClient 从服务上下文中安全取出 driversvc 客户端。
func (l *OfflineLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
