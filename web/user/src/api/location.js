import request from './request'

// POISearch 调用乘客端后端的位置搜索代理接口（passenger-api → locationsvc → AMap）。
// 不再直接调用 AMap JS API，避免浏览器端 key 类型不匹配与泄露问题。
export function poiSearch(params) {
  return request.post('/location/poi-search', params)
}

// 将 GPS 坐标交给乘客端后端反查文字地址，用于首页定位回显。
export function reverseGeocodeLocation(data) {
  return request.post('/location/reverse-geocode', data)
}

// 将乘客输入的地址交给乘客端后端解析成经纬度，用于下单前保存终点坐标。
export function geocodeAddress(data) {
  return request.post('/location/geocode', data)
}
