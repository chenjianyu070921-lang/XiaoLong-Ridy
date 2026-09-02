package logic

import (
	"context"
	"google.golang.org/grpc"
	"math"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	locationproto "XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultGeocodeRadius = 50000
	defaultGeocodeSize   = 10
)

// LocationLogic 提供乘客端的位置能力：目的地 POI 关键词搜索（后端代理）、坐标/地址互转
// 与附近司机查询。统一由后端调用 locationsvc 的高德 key，避开浏览器暴露 key 与类型不匹配的问题。

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
	logx.Logger
}

// NewLocationLogic 创建位置业务逻辑。
func NewLocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LocationLogic {
	return &LocationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// POISearch 转发乘客端目的地搜索请求到 locationsvc，
// 由 locationsvc 统一使用后端 AMap key 完成真实 POI 检索并写入本地缓存。
func (l *LocationLogic) POISearch(req *types.POISearchRequest) (*types.POISearchResponse, error) {
	if req.Keyword == "" {
		return nil, ErrInvalidRequest
	}
	client, err := l.locationClient()
	if err != nil {
		return nil, err
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 20
	}
	if size > 50 {
		size = 50
	}
	radius := req.Radius
	if radius <= 0 {
		radius = 3000
	}
	resp, err := client.POISearch(l.ctx, &locationproto.POISearchReq{
		Keyword: req.Keyword,
		City:    req.City,
		Lat:     req.Lat,
		Lng:     req.Lng,
		Radius:  radius,
		Page:    page,
		Size:    size,
	})
	if err != nil {
		l.Errorf("POISearch 调用 locationsvc 失败: %v", err)
		return nil, err
	}
	items := make([]types.POIDTO, 0, len(resp.GetItems()))
	for _, p := range resp.GetItems() {
		items = append(items, types.POIDTO{
			Name:     p.GetName(),
			Address:  p.GetAddress(),
			Lat:      p.GetLat(),
			Lng:      p.GetLng(),
			Category: p.GetCategory(),
			Distance: p.GetDistance(),
		})
	}
	return &types.POISearchResponse{Items: items, Total: resp.GetTotal()}, nil
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
