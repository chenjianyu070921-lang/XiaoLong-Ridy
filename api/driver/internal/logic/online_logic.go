package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

// OnlineLogic 封装司机上线逻辑。
type OnlineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewOnlineLogic 构造司机上线逻辑处理器。
func NewOnlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OnlineLogic {
	return &OnlineLogic{ctx: ctx, svcCtx: svcCtx}
}

// SetOnline 将当前登录司机置为在线。driverID 由鉴权中间件从 JWT 解析得到。
func (l *OnlineLogic) SetOnline(driverID int64) (*types.SetOnlineResponse, error) {
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.SetDriverOnline(l.ctx, &driversproto.SetDriverOnlineRequest{DriverId: driverID})
	if err != nil {
		return nil, err
	}
	return &types.SetOnlineResponse{
		DriverID:     resp.GetDriverId(),
		OnlineStatus: int(resp.GetOnlineStatus()),
	}, nil
}

// driverClient 从服务上下文中安全取出 driversvc 客户端。
func (l *OnlineLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
