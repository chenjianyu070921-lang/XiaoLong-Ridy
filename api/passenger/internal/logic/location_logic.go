package logic

import (
	"context"
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

// LocationLogic 提供乘客端的位置能力：目的地 POI 关键词搜索（后端代理）与坐标/地址互转。
// 统一由后端调用 locationsvc 的高德 key，避开浏览器暴露 key 与类型不匹配的问题。
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
