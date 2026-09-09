// 管理后台地图 SDK 配置，允许通过 VITE_AMAP_* 环境变量覆盖默认值。
const config = { key: '8b282c23710212a9761f53fed184c2aa', securityCode: '2d360911e1fb4f65b9b2954d01c5d094' }
export function getAmapConfig() {
  return { key: import.meta.env.VITE_AMAP_KEY || config.key, securityCode: import.meta.env.VITE_AMAP_SECURITY_CODE || config.securityCode }
}
