// 高德地图公共配置：集中管理乘客端地图 SDK 的默认 Key 和安全密钥。
// 生产或个人环境可通过 .env.local 中的 VITE_AMAP_* 变量覆盖默认值。
// 注意：这里的 key 必须是 Web端(JS API) 平台类型；后端 locationsvc 用独立的 Web服务 key。
const publicAmapConfig = {
  key: '8b282c23710212a9761f53fed184c2aa',
  securityCode: '2d360911e1fb4f65b9b2954d01c5d094'
}

// 读取高德配置，环境变量优先，未配置时使用项目公用配置。
export function getAmapConfig() {
  return {
    key: import.meta.env.VITE_AMAP_KEY || publicAmapConfig.key,
    securityCode: import.meta.env.VITE_AMAP_SECURITY_CODE || publicAmapConfig.securityCode
  }
}

export default publicAmapConfig