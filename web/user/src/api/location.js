import request from './request'

// 将 GPS 坐标交给乘客端后端反查文字地址，用于首页定位回显。
export function reverseGeocodeLocation(data) {
  return request.post('/location/reverse-geocode', data)
}

// 将乘客输入的地址交给乘客端后端解析成经纬度，用于下单前保存终点坐标。
export function geocodeAddress(data) {
  return request.post('/location/geocode', data)
}

// 查询乘客当前位置 5 公里内的在线司机，供首页地图绘制车辆图标。
export function getNearbyDrivers(data) {
  return request.post('/location/nearby-drivers', data)
}
