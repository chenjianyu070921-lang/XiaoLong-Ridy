package geo

import "testing"

// 北京天安门 → 国贸 → 望京 的坐标样本。
const (
	tiananmenLng = 116.397
	tiananmenLat = 39.908
	guomaoLng    = 116.461
	guomaoLat    = 39.909
	wangjingLng  = 116.470
	wangjingLat  = 40.000
)

func TestHaversineMetersKnownDistance(t *testing.T) {
	// 天安门→国贸 直线约 5.5km，允许 10% 误差。
	got := HaversineMeters(tiananmenLng, tiananmenLat, guomaoLng, guomaoLat)
	if got < 5000 || got > 6200 {
		t.Fatalf("distance = %.0f, want ~5500", got)
	}
}

// 顺路：司机在国贸，订单终点望京与回家方向（望京）一致 → 通过。
func TestIsOnRouteAcceptsSameDirection(t *testing.T) {
	if !IsOnRoute(guomaoLng, guomaoLat, wangjingLng, wangjingLat, wangjingLng, wangjingLat, 0.2) {
		t.Fatal("同向订单应判定为顺路")
	}
}

// 反向：司机在国贸，家在东北的望京，订单终点却在西南的天安门 → 绕路远超阈值，应过滤。
func TestIsOnRouteRejectsOppositeDirection(t *testing.T) {
	if IsOnRoute(guomaoLng, guomaoLat, tiananmenLng, tiananmenLat, wangjingLng, wangjingLat, 0.2) {
		t.Fatal("反向订单应被过滤")
	}
}

// 坐标缺失时不误杀：一律放行。
func TestIsOnRouteSkipsWhenCoordinateMissing(t *testing.T) {
	if !IsOnRoute(0, 0, tiananmenLng, tiananmenLat, wangjingLng, wangjingLat, 0.2) {
		t.Fatal("坐标缺失应放行")
	}
}

func TestValid(t *testing.T) {
	if Valid(0, 0) || Valid(200, 40) {
		t.Fatal("非法坐标应判为无效")
	}
	if !Valid(116.397, 39.908) {
		t.Fatal("合法坐标应判为有效")
	}
}
