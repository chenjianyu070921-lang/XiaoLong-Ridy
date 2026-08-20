package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

// OnlineLogic 封装司机上线逻辑，持有请求上下文与下游 driversvc 客户端。
type OnlineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewOnlineLogic 构造司机上线逻辑处理器，注入请求上下文与服务上下文。
func NewOnlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OnlineLogic {
	return &OnlineLogic{ctx: ctx, svcCtx: svcCtx}
}

// SetOnline 将当前登录司机置为在线。driverID 由鉴权中间件从 JWT 解析得到，
// deviceID 为当前设备标识（用于多端互踢），longitude/latitude 为上报位置（可选），透传给 driversvc。
func (l *OnlineLogic) SetOnline(driverID int64, deviceID string, longitude, latitude float64) (*types.SetOnlineResponse, error) {
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.SetDriverOnline(l.ctx, &driversproto.SetDriverOnlineRequest{
		DriverId:  driverID,
		DeviceId:  deviceID,
		Longitude: longitude,
		Latitude:  latitude,
	})
	if err != nil {
		return nil, err
	}
	return &types.SetOnlineResponse{
		DriverID:     resp.GetDriverId(),
		OnlineStatus: int(resp.GetOnlineStatus()),
		Kicked:       resp.GetKicked(),
	}, nil
}

// driverClient 从服务上下文中安全取出 driversvc 客户端。
func (l *OnlineLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
