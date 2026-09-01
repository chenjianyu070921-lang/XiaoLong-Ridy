package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	locationproto "XiaoLong-Ridy/rpc/locationsvc/locationsvc"
)

const (
	defaultGeocodeRadius = 50000
	defaultGeocodeSize   = 10
)

// LocationLogic 封装乘客端定位地址解析流程。
type LocationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLocationLogic 创建乘客端定位解析逻辑。
func NewLocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LocationLogic {
	return &LocationLogic{ctx: ctx, svcCtx: svcCtx}
}

// ReverseGeocode 将乘客 GPS 坐标转换为可回显的文字地址。
func (l *LocationLogic) ReverseGeocode(req *types.ReverseGeocodeRequest) (*types.ReverseGeocodeResponse, error) {
	if req == nil || !isValidLongitudeLatitude(req.Longitude, req.Latitude) {
		return nil, ErrInvalidRequest
	}
	client, err := l.locationClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.ReverseGeocode(l.ctx, &locationproto.ReverseGeocodeReq{
		Lng: req.Longitude,
		Lat: req.Latitude,
	})
	if err != nil {
		return nil, err
	}
	return &types.ReverseGeocodeResponse{
		Address:   strings.TrimSpace(resp.GetAddress()),
		Province:  strings.TrimSpace(resp.GetProvince()),
		CityName:  strings.TrimSpace(resp.GetCity()),
		District:  strings.TrimSpace(resp.GetDistrict()),
		PoiName:   strings.TrimSpace(resp.GetPoiName()),
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	}, nil
}

// Geocode 将乘客输入的目的地文字解析为下单坐标。
func (l *LocationLogic) Geocode(req *types.GeocodeRequest) (*types.GeocodeResponse, error) {
	if req == nil || strings.TrimSpace(req.Address) == "" {
		return nil, ErrInvalidRequest
	}
	client, err := l.locationClient()
	if err != nil {
		return nil, err
	}
	radius := req.Radius
	if radius <= 0 {
		radius = defaultGeocodeRadius
	}
	resp, err := client.POISearch(l.ctx, &locationproto.POISearchReq{
		Keyword: strings.TrimSpace(req.Address),
		City:    strings.TrimSpace(req.CityCode),
		Lng:     req.Longitude,
		Lat:     req.Latitude,
		Radius:  radius,
		Page:    1,
		Size:    defaultGeocodeSize,
	})
	if err != nil {
		return nil, err
	}
	for _, item := range resp.GetItems() {
		if item == nil || !isValidLongitudeLatitude(item.GetLng(), item.GetLat()) {
			continue
		}
		return &types.GeocodeResponse{
			Name:      strings.TrimSpace(item.GetName()),
			Address:   strings.TrimSpace(item.GetAddress()),
			Longitude: item.GetLng(),
			Latitude:  item.GetLat(),
			Category:  strings.TrimSpace(item.GetCategory()),
			Distance:  item.GetDistance(),
			CityCode:  strings.TrimSpace(req.CityCode),
		}, nil
	}
	return nil, ErrInvalidRequest
}

func (l *LocationLogic) locationClient() (svc.LocationClient, error) {
	if l.svcCtx == nil || l.svcCtx.LocationClient == nil {
		return nil, ErrLocationClientNotConfigured
	}
	return l.svcCtx.LocationClient, nil
}
