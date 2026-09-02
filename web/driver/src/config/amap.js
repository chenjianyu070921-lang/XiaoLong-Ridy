const publicAmapConfig = {
  key: 'fe003b9e1b112c532e653d35fe9dbc89',
  securityCode: '89b3eb6651cdbe10674ae29da44af88d'
}

export function getAmapConfig() {
  return {
    key: import.meta.env.VITE_AMAP_KEY || publicAmapConfig.key,
    securityCode: import.meta.env.VITE_AMAP_SECURITY_CODE || publicAmapConfig.securityCode
  }
}

export default publicAmapConfig
