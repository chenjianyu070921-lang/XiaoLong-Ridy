// Package geo 提供与派单/听单相关的地理位置计算工具。
package geo

import "math"

const earthRadiusMeters = 6371000.0

// HaversineMeters 计算两个经纬度之间的球面距离（米）。
func HaversineMeters(lng1, lat1, lng2, lat2 float64) float64 {
	dLat := degreesToRadians(lat2 - lat1)
	dLng := degreesToRadians(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(degreesToRadians(lat1))*math.Cos(degreesToRadians(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return earthRadiusMeters * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// IsOnRoute 判定"司机 → 订单终点 → 回家目的地"是否顺路（回家模式过滤）。
// 判定口径：绕路代价 = (经订单终点的路程 - 直达回家的路程) / 直达回家的路程，
// 不超过最大绕路比例即视为顺路；反向单会导致绕路代价远大于阈值而被过滤。
// 司机已在目的地附近（直达距离过小）时不判定，直接放行。
func IsOnRoute(driverLng, driverLat, orderToLng, orderToLat, homeLng, homeLat, maxDetourRatio float64) bool {
	if !Valid(driverLng, driverLat) || !Valid(homeLng, homeLat) || !Valid(orderToLng, orderToLat) {
		return true // 坐标缺失不做过滤，避免误杀订单
	}
	if maxDetourRatio <= 0 {
		maxDetourRatio = 0.2
	}
	direct := HaversineMeters(driverLng, driverLat, homeLng, homeLat)
	if direct <= 1 {
		return true
	}
	via := HaversineMeters(driverLng, driverLat, orderToLng, orderToLat) +
		HaversineMeters(orderToLng, orderToLat, homeLng, homeLat)
	detourRatio := (via - direct) / direct
	return detourRatio <= maxDetourRatio
}

// Valid 判断经纬度是否为有效值（非零且在合理区间）。
func Valid(lng, lat float64) bool {
	if lng == 0 && lat == 0 {
		return false
	}
	return lng >= -180 && lng <= 180 && lat >= -90 && lat <= 90
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}
