package logic

import (
	"context"

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

// POISearch POI 搜索
func (l *POISearchLogic) POISearch(in *locationsvc.POISearchReq) (*locationsvc.POISearchResp, error) {
	l.Infof("POISearch: keyword=%s, lat=%f, lng=%f, radius=%d", in.Keyword, in.Lat, in.Lng, in.Radius)

	return &locationsvc.POISearchResp{
		Items: []*locationsvc.POIItem{
			{
				Name:     "深圳湾科技生态园",
				Address:  "深圳市南山区沙河西路",
				Lat:      22.5431,
				Lng:      113.9517,
				Category: "商务写字楼",
				Distance: 200,
			},
			{
				Name:     "万象天地",
				Address:  "深圳市南山区深南大道",
				Lat:      22.5362,
				Lng:      113.9556,
				Category: "购物中心",
				Distance: 800,
			},
		},
		Total: 2,
	}, nil
}
