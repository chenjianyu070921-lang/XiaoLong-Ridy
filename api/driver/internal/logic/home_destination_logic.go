package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

// HomeDestinationLogic 封装司机回家目的地与回家顺路模式的读写。
type HomeDestinationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewHomeDestinationLogic 创建回家目的地逻辑处理器。
func NewHomeDestinationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HomeDestinationLogic {
	return &HomeDestinationLogic{ctx: ctx, svcCtx: svcCtx}
}

// Get 查询当前司机的回家目的地与模式状态。
func (l *HomeDestinationLogic) Get(driverID int64) (*types.HomeDestinationResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetHomeDestination(l.ctx, &driversproto.GetHomeDestinationRequest{DriverId: driverID})
	if err != nil {
		return nil, err
	}
	return &types.HomeDestinationResponse{
		HasSetting:     resp.GetHasSetting(),
		HomeAddr:       resp.GetHomeAddr(),
		HomeLng:        resp.GetHomeLng(),
		HomeLat:        resp.GetHomeLat(),
		IsHomeModeOpen: resp.GetIsHomeModeOpen(),
		MaxDetourRatio: resp.GetMaxDetourRatio(),
	}, nil
}

// Set 保存回家目的地；open=true 时同时开启回家顺路模式（只改过滤规则，不影响听单状态）。
func (l *HomeDestinationLogic) Set(driverID int64, req *types.SetHomeDestinationRequest) (*types.HomeDestinationResponse, error) {
	if driverID <= 0 || req == nil {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.SetHomeDestination(l.ctx, &driversproto.SetHomeDestinationRequest{
		DriverId:       driverID,
		HomeAddr:       req.HomeAddr,
		HomeLng:        req.HomeLng,
		HomeLat:        req.HomeLat,
		Open:           req.Open,
		MaxDetourRatio: req.MaxDetourRatio,
	})
	if err != nil {
		return nil, err
	}
	return &types.HomeDestinationResponse{
		HasSetting:     resp.GetHasSetting(),
		HomeAddr:       resp.GetHomeAddr(),
		HomeLng:        resp.GetHomeLng(),
		HomeLat:        resp.GetHomeLat(),
		IsHomeModeOpen: resp.GetIsHomeModeOpen(),
		MaxDetourRatio: resp.GetMaxDetourRatio(),
	}, nil
}

// SetMode 仅切换回家顺路模式开关；关闭即恢复全域听单。
func (l *HomeDestinationLogic) SetMode(driverID int64, req *types.SetHomeModeRequest) (*types.HomeDestinationResponse, error) {
	if driverID <= 0 || req == nil {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.SetHomeMode(l.ctx, &driversproto.SetHomeModeRequest{DriverId: driverID, Open: req.Open})
	if err != nil {
		return nil, err
	}
	return &types.HomeDestinationResponse{
		HasSetting:     resp.GetHasSetting(),
		HomeAddr:       resp.GetHomeAddr(),
		HomeLng:        resp.GetHomeLng(),
		HomeLat:        resp.GetHomeLat(),
		IsHomeModeOpen: resp.GetIsHomeModeOpen(),
		MaxDetourRatio: resp.GetMaxDetourRatio(),
	}, nil
}

func (l *HomeDestinationLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
