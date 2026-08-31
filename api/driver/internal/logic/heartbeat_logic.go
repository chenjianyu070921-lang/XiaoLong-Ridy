package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

// HeartbeatLogic 封装司机心跳上报逻辑，持有请求上下文与下游 driversvc 客户端。
type HeartbeatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewHeartbeatLogic 构造司机心跳上报逻辑处理器，注入请求上下文与服务上下文。
func NewHeartbeatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HeartbeatLogic {
	return &HeartbeatLogic{ctx: ctx, svcCtx: svcCtx}
}

// Heartbeat 上报司机心跳：透传 deviceID 与位置到 driversvc，刷新在线状态保活并判定多端互踢。
// driverID 由鉴权中间件从 JWT 解析得到；deviceID 取自请求体，用于互踢判定。
func (l *HeartbeatLogic) Heartbeat(driverID int64, req *types.HeartbeatRequest) (*types.HeartbeatResponse, error) {
	if driverID <= 0 || req == nil || req.DeviceID == "" || !validLocation(req.Longitude, req.Latitude) {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.Heartbeat(l.ctx, &driversproto.HeartbeatRequest{
		DriverId:  driverID,
		DeviceId:  req.DeviceID,
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	})
	if err != nil {
		return nil, err
	}
	return &types.HeartbeatResponse{
		OnlineStatus: int(resp.GetOnlineStatus()),
		Kicked:       resp.GetKicked(),
		ServerTime:   resp.GetServerTime(),
	}, nil
}

// driverClient 从服务上下文中安全取出 driversvc 客户端。
func (l *HeartbeatLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
