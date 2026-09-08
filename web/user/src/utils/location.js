// 乘客端统一地点数据结构与转换工具。
// 地图选点、POI 搜索、逆地理编码均产出统一的 Location 对象，
// 避免各处各自拼装地址字符串，导致下单、估价与展示字段不一致。

// 地点类型：上车点 / 目的地。
export const LOCATION_TYPE = {
  PICKUP: 'pickup',
  DESTINATION: 'destination'
}

// 判断经纬度是否为有效坐标（排除 0 与 NaN/Infinity）。
export function hasValidCoordinate(latitude, longitude) {
  return Number.isFinite(Number(latitude)) && Number.isFinite(Number(longitude)) && Number(latitude) !== 0 && Number(longitude) !== 0
}

/**
 * Location 统一结构：
 * {
 *   name: string;       // 地点简短名（POI 名或定位点），用于订单展示
 *   address: string;    // 完整详细地址，用于叫车页地址栏
 *   latitude: number;
 *   longitude: number;
 *   poiId?: string;     // 高德 POI id，非定位点才有
 *   province?: string;
 *   city?: string;      // 城市名，例如「宿迁市」
 *   district?: string;
 *   street?: string;    // 后端逆地理暂未单独返回，先以 district 兜底
 * }
 *
 * 兼容历史 lng/lat、cityName、poiName 等命名，统一规整字段。
 */
export function createLocation(data = {}) {
  const latitude = Number(data.latitude ?? data.lat)
  const longitude = Number(data.longitude ?? data.lng)
  return {
    name: data.name || data.poiName || '',
    address: data.address || '',
    latitude,
    longitude,
    poiId: data.poiId || data.id || '',
    province: data.province || '',
    city: data.city || data.cityName || '',
    district: data.district || '',
    street: data.street || data.district || ''
  }
}

// 将 Location 映射为 orderParams 的下单字段（from*/to*），供既有下单与估价流程继续使用。
export function toOrderParams(location, type) {
  const loc = createLocation(location)
  const prefix = type === LOCATION_TYPE.PICKUP ? 'from' : 'to'
  return {
    [`${prefix}Address`]: loc.name,
    [`${prefix}Lat`]: loc.latitude,
    [`${prefix}Lng`]: loc.longitude
  }
}

// 将 orderParams 中的起终点字段反向构造为 Location，供展示与 estimatePrice 使用。
export function locationFromOrderParams(params, type) {
  const prefix = type === LOCATION_TYPE.PICKUP ? 'from' : 'to'
  return createLocation({
    name: params[`${prefix}Address`],
    latitude: params[`${prefix}Lat`],
    longitude: params[`${prefix}Lng`]
  })
}
