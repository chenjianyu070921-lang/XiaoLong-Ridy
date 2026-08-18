package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/locationsvc/internal/model"
	"XiaoLong-Ridy/rpc/locationsvc/internal/svc"
	"XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type POISearchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPOISearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *POISearchLogic {
	return &POISearchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// POISearch POI 搜索：先查本地缓存库，没有再调高德 API，结果写回缓存
func (l *POISearchLogic) POISearch(in *locationsvc.POISearchReq) (*locationsvc.POISearchResp, error) {
	l.Infof("POISearch: keyword=%s, lat=%f, lng=%f, radius=%d", in.Keyword, in.Lat, in.Lng, in.Radius)

	// 1. 先查本地缓存库
	cached, err := l.svcCtx.PoiModel.SearchByName(in.Keyword, int(in.Size))
	if err == nil && len(cached) > 0 {
		l.Infof("命中本地缓存，共 %d 条", len(cached))
		return toPoiSearchResp(cached, int32(len(cached))), nil
	}
	if err != nil {
		l.Errorf("查询本地POI缓存失败: %v", err)
	}

	// 2. 本地没有，调高德 API
	amapResp, err := l.svcCtx.GetGeo().SearchPoi(in.Keyword, in.Lat, in.Lng, in.Radius, in.Page, in.Size)
	if err != nil {
		l.Errorf("调用高德POI搜索失败: %v", err)
		return nil, err
	}

	// 3. 高德结果写回本地库（下次搜索直接命中缓存，不再调高德）
	pois := amapResp.ToPoiModels()
	if len(pois) > 0 {
		if err := l.svcCtx.PoiModel.BatchInsert(pois); err != nil {
			l.Errorf("写入POI缓存失败: %v", err)
		}
	}

	l.Infof("高德返回 %d 条结果", len(pois))
	return toPoiSearchResp(pois, amapResp.Total()), nil
}

// toPoiSearchResp 把本地 POI 模型转成 proto 响应
func toPoiSearchResp(pois []model.Poi, total int32) *locationsvc.POISearchResp {
	resp := &locationsvc.POISearchResp{
		Items: make([]*locationsvc.POIItem, 0, len(pois)),
		Total: total,
	}
	for _, p := range pois {
		resp.Items = append(resp.Items, &locationsvc.POIItem{
			Name:     p.Name,
			Address:  p.Address,
			Lat:      p.Latitude,
			Lng:      p.Longitude,
			Category: p.Category,
			Distance: int32(p.Distance),
		})
	}
	return resp
}
