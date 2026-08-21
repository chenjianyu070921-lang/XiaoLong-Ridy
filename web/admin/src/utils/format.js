// 前端展示辅助函数：不改变后端数据类型，只负责中文标签和安全展示。
export const text = (v) => (v === null || v === undefined || v === '' ? '-' : String(v))
export const statusText = (v, map) => map[v] || `未知(${v ?? '-'})`
export const statusTag = (v, map) => map[v] || 'info'
export const pageData = (data) => ({ list: data?.list || [], total: data?.total || 0 })
