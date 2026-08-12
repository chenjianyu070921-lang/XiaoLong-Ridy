package logic

import (
	"context"
	"fmt"

	"XiaoLong-Ridy/rpc/locationsvc/internal/svc"
	"XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReverseGeocodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReverseGeocodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReverseGeocodeLogic {
	return &ReverseGeocodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReverseGeocode 逆地址解析（经纬度 → 地址）
func (l *ReverseGeocodeLogic) ReverseGeocode(in *locationsvc.ReverseGeocodeReq) (*locationsvc.ReverseGeocodeResp, error) {
	l.Infof("ReverseGeocode: lat=%f, lng=%f", in.Lat, in.Lng)

	return &locationsvc.ReverseGeocodeResp{
		Province: "广东省",
		City:     "深圳市",
		District: "南山区",
		Address:  fmt.Sprintf("深圳市南山区科技园(%.4f, %.4f)", in.Lat, in.Lng),
		PoiName:  "科技园",
	}, nil
}
