// 高德地图公共配置：集中管理乘客端地图 SDK 的默认 Key 和安全密钥。
// 生产或个人环境可通过 .env.local 中的 VITE_AMAP_* 变量覆盖默认值。
const publicAmapConfig = {
  key: 'fe003b9e1b112c532e653d35fe9dbc89',
  securityCode: '89b3eb6651cdbe10674ae29da44af88d'
}

// 读取高德配置，环境变量优先，未配置时使用项目公用配置。
export function getAmapConfig() {
  return {
    key: import.meta.env.VITE_AMAP_KEY || publicAmapConfig.key,
    securityCode: import.meta.env.VITE_AMAP_SECURITY_CODE || publicAmapConfig.securityCode
  }
}

export default publicAmapConfig
