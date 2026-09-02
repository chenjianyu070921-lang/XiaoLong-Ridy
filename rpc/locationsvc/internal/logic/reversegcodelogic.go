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

// ReverseGeocode 逆地址解析（经纬度 → 地址），调用高德 regeo 接口返回真实数据
func (l *ReverseGeocodeLogic) ReverseGeocode(in *locationsvc.ReverseGeocodeReq) (*locationsvc.ReverseGeocodeResp, error) {
	l.Infof("ReverseGeocode: lat=%f, lng=%f", in.Lat, in.Lng)

	regeo, err := l.svcCtx.GetGeo().ReverseGeocode(in.Lat, in.Lng)
	if err != nil {
		l.Errorf("调用高德逆地理编码失败: %v", err)
		// 地图服务属于外部依赖，失败时返回坐标本身，避免乘客定位流程整体中断。
		return &locationsvc.ReverseGeocodeResp{Address: fmt.Sprintf("%.6f,%.6f", in.Lng, in.Lat)}, nil
	}

	comp := regeo.Regeocode.AddressComponent
	address := regeo.AddressStr()

	// 高德只支持中国境内坐标；境外坐标 formatted_address 会是空数组
	if address == "" {
		// 地址为空属于可预期情况（境外坐标、服务降级或数据缺失），返回坐标兜底。
		return &locationsvc.ReverseGeocodeResp{Address: fmt.Sprintf("%.6f,%.6f", in.Lng, in.Lat)}, nil
	}

	// 兜底：直辖市的高德 city 字段是空数组，用省代替（如"北京市"）
	city := regeo.CityStr()
	if city == "" {
		city = comp.Province
	}

	// 附近 POI 名称（base 模式下可能为空，取第一个）
	poiName := ""
	if len(regeo.Regeocode.Pois) > 0 {
		poiName = regeo.Regeocode.Pois[0].Name
	}

	return &locationsvc.ReverseGeocodeResp{
		Province: comp.Province,
		City:     city,
		District: comp.District,
		Address:  address,
		PoiName:  poiName,
	}, nil
}
