import request from './request'

// POISearch 调用乘客端后端的位置搜索代理接口（passenger-api → locationsvc → AMap）。
// 不再直接调用 AMap JS API，避免浏览器端 key 类型不匹配与泄露问题。
export function poiSearch(params) {
  return request.post('/location/poi-search', params)
}