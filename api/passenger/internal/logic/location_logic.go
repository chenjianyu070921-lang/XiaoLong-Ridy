package logic

import (
	"context"
	"google.golang.org/grpc"
	"math"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	locationproto "XiaoLong-Ridy/rpc/locationsvc/locationsvc"
)

const (
	defaultGeocodeRadius = 50000
	defaultGeocodeSize   = 10
)

// NearbyDrivers 查询指定半径内仍在听单的司机，默认限制 5 公里和 50 个结果。
func (l *LocationLogic) NearbyDrivers(req *types.NearbyDriversRequest) ([]*types.NearbyDriverResponse, error) {
	if req == nil || !isValidLongitudeLatitude(req.Longitude, req.Latitude) {
		return nil, ErrInvalidRequest
	}
	radius := req.Radius
	// 拒绝 NaN/Inf，避免非法半径进入位置服务；超出业务范围统一收敛到 5 公里。
	if math.IsNaN(radius) || math.IsInf(radius, 0) || radius <= 0 || radius > 5000 {
		radius = 5000
	}
	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 50
	}
	client, err := l.locationClient()
	if err != nil {
		// 位置服务未接入时按无附近司机处理，保证首页仍可正常使用。
		return []*types.NearbyDriverResponse{}, nil
	}
	// 附近司机是位置客户端的可选扩展，兼容尚未升级该 RPC 的旧 mock/本地客户端。
	nearbyClient, ok := client.(interface {
		NearbyDrivers(context.Context, *locationproto.NearbyDriversReq, ...grpc.CallOption) (*locationproto.NearbyDriversResp, error)
	})
	if !ok {
		return []*types.NearbyDriverResponse{}, nil
	}
	resp, err := nearbyClient.NearbyDrivers(l.ctx, &locationproto.NearbyDriversReq{Lng: req.Longitude, Lat: req.Latitude, Radius: radius, Limit: limit})
	if err != nil {
		// 附近司机仅用于地图增强展示；位置服务暂不可用时返回空列表，不能阻断乘客叫车主流程。
		return []*types.NearbyDriverResponse{}, nil
	}
	// RPC 可能在异常实现或 mock 中返回 nil 响应；将其视为空结果，避免网关 panic。
	if resp == nil {
		return []*types.NearbyDriverResponse{}, nil
	}
	items := make([]*types.NearbyDriverResponse, 0, len(resp.Drivers))
	for _, d := range resp.Drivers {
		if d == nil || d.DriverId <= 0 || !isValidLongitudeLatitude(d.Lng, d.Lat) || math.IsNaN(d.Distance) || math.IsInf(d.Distance, 0) || d.Distance < 0 {
			continue
		}
		items = append(items, &types.NearbyDriverResponse{DriverID: d.DriverId, Longitude: d.Lng, Latitude: d.Lat, Distance: d.Distance})
	}
	return items, nil
}

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
