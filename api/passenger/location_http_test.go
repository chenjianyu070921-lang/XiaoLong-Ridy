package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"XiaoLong-Ridy/api/passenger/internal/router"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	locationproto "XiaoLong-Ridy/rpc/locationsvc/locationsvc"
	"XiaoLong-Ridy/rpc/usersvc/client"

	"google.golang.org/grpc"
)

type fakeLocationClient struct{}

func (fakeLocationClient) ReverseGeocode(_ context.Context, req *locationproto.ReverseGeocodeReq, _ ...grpc.CallOption) (*locationproto.ReverseGeocodeResp, error) {
	return &locationproto.ReverseGeocodeResp{
		Province: "江苏省",
		City:     "宿迁市",
		District: "宿城区",
		Address:  "江苏省宿迁市宿城区人民大道1号",
		PoiName:  "市政府",
	}, nil
}

func (fakeLocationClient) POISearch(_ context.Context, req *locationproto.POISearchReq, _ ...grpc.CallOption) (*locationproto.POISearchResp, error) {
	return &locationproto.POISearchResp{
		Items: []*locationproto.POIItem{{
			Name:     "宿迁宝龙广场",
			Address:  "江苏省宿迁市宿城区西湖路",
			Lat:      33.9636,
			Lng:      118.2752,
			Category: "购物服务",
			Distance: 1200,
		}},
		Total: 1,
	}, nil
}

func (fakeLocationClient) GetDriverLocation(context.Context, *locationproto.GetDriverLocationReq, ...grpc.CallOption) (*locationproto.GetDriverLocationResp, error) {
	return &locationproto.GetDriverLocationResp{}, nil
}

func (fakeLocationClient) RoutePlan(context.Context, *locationproto.RoutePlanReq, ...grpc.CallOption) (*locationproto.RoutePlanResp, error) {
	return &locationproto.RoutePlanResp{}, nil
}

// TestPassengerLocationHTTPFlow 验证乘客端定位回显和目的地地址解析都通过 passenger API 转发到位置服务。
func TestPassengerLocationHTTPFlow(t *testing.T) {
	server := httptest.NewServer(router.NewRouter(svc.NewServiceContext(
		client.NewLocalClient("test-signing-key", nil),
		svc.WithLocationClient(fakeLocationClient{}),
	)))
	defer server.Close()

	reverseResponse := callJSON(t, http.MethodPost, server.URL+"/api/passenger/v1/location/reverse-geocode", map[string]float64{
		"longitude": 118.2752,
		"latitude":  33.9636,
	}, "")
	if reverseResponse.Code != 0 {
		t.Fatalf("reverse geocode response = %+v", reverseResponse)
	}
	reverseData := decodeData[types.ReverseGeocodeResponse](t, reverseResponse.Data)
	if reverseData.CityName != "宿迁市" || reverseData.Address == "" {
		t.Fatalf("reverse geocode data = %+v", reverseData)
	}

	geocodeResponse := callJSON(t, http.MethodPost, server.URL+"/api/passenger/v1/location/geocode", map[string]any{
		"address":   "宿迁宝龙广场",
		"cityCode":  "321300",
		"longitude": 118.2752,
		"latitude":  33.9636,
	}, "")
	if geocodeResponse.Code != 0 {
		t.Fatalf("geocode response = %+v", geocodeResponse)
	}
	geocodeData := decodeData[types.GeocodeResponse](t, geocodeResponse.Data)
	if geocodeData.Name != "宿迁宝龙广场" || geocodeData.Longitude == 0 || geocodeData.Latitude == 0 {
		t.Fatalf("geocode data = %+v", geocodeData)
	}
}
