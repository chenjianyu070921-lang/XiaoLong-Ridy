import AMapLoader from '@amap/amap-jsapi-loader'

// 注意：这里的 key 必须是 Web端(JS API) 平台类型；后端 locationsvc 用独立的 Web服务 key。
// 司机端与乘客端共用同一 JS API key（按域名区分，不按 app 区分），与 web/user 保持一致。
const publicAmapConfig = {
  key: '8b282c23710212a9761f53fed184c2aa',
  securityCode: '2d360911e1fb4f65b9b2954d01c5d094'
}

let amapLoadPromise = null

export function getAmapConfig() {
  return {
    key: import.meta.env.VITE_AMAP_KEY || publicAmapConfig.key,
    securityCode: import.meta.env.VITE_AMAP_SECURITY_CODE || publicAmapConfig.securityCode
  }
}

export async function loadDriverAmap(plugins = []) {
  const pluginList = normalizeAmapPlugins(plugins)
  if (window.AMap) {
    await ensureAmapPlugins(window.AMap, pluginList)
    return window.AMap
  }

  const { key, securityCode } = getAmapConfig()
  if (!key) throw new Error('未配置高德地图Key')
  if (securityCode) window._AMapSecurityConfig = { securityJsCode: securityCode }

  if (!amapLoadPromise) {
    amapLoadPromise = AMapLoader.load({ key, version: '2.0', plugins: pluginList })
      .then(async (sdk) => {
        window.AMap = sdk
        await ensureAmapPlugins(sdk, pluginList)
        return sdk
      })
      .catch((error) => {
        amapLoadPromise = null
        throw error
      })
  }

  const sdk = await amapLoadPromise
  await ensureAmapPlugins(sdk, pluginList)
  return sdk
}

function normalizeAmapPlugins(plugins) {
  return Array.from(new Set((Array.isArray(plugins) ? plugins : [plugins]).filter(Boolean))).sort()
}

function ensureAmapPlugins(sdk, plugins) {
  if (!plugins.length || typeof sdk?.plugin !== 'function') return Promise.resolve()
  return new Promise((resolve) => {
    sdk.plugin(plugins, resolve)
  })
}

export default publicAmapConfig
