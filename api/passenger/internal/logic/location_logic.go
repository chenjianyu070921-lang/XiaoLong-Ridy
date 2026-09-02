package logic

import (
	"context"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	locationproto "XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"github.com/zeromicro/go-zero/core/logx"
)

// LocationLogic 提供乘客端的位置/POI 检索能力，
// 通过后端代理避开浏览器暴露 AMap key 与服务端密钥类型不匹配的问题。
type LocationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewLocationLogic 创建位置检索逻辑。
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
	if l.svcCtx == nil || l.svcCtx.LocationClient == nil {
		return nil, ErrLocationClientNotConfigured
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
	resp, err := l.svcCtx.LocationClient.POISearch(l.ctx, &locationproto.POISearchReq{
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