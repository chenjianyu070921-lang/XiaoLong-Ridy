// 全局枚举 -> 中文文案映射。
// 数值来源：scripts/sql/migrate/*.sql 注释与 rpc/*/proto 定义。

export const orderStatusText = (s) => ({ 1: '待接单', 2: '已接单', 3: '行程中', 4: '待支付', 5: '已完成', 6: '已取消' }[s] || `未知(${s})`)
export const orderStatusTag = (s) => ({ 1: 'info', 2: 'primary', 3: 'warning', 4: 'warning', 5: 'success', 6: 'danger' }[s] || 'info')

export const userStatusText = (s) => ({ 1: '正常', 2: '冻结' }[s] || `未知(${s})`)
export const userStatusTag = (s) => ({ 1: 'success', 2: 'danger' }[s] || 'info')

export const driverStatusText = (s) => ({ 1: '待审核', 2: '正常', 3: '冻结', 4: '注销' }[s] || `未知(${s})`)
export const vehicleStatusText = (s) => ({ 1: '待审核', 2: '正常', 3: '禁用' }[s] || `未知(${s})`)
export const auditStatusText = (s) => ({ 1: '待审核', 2: '通过', 3: '驳回' }[s] || `未知(${s})`)
export const auditStatusTag = (s) => ({ 1: 'warning', 2: 'success', 3: 'danger' }[s] || 'info')

export const genderText = (g) => ({ 0: '未知', 1: '男', 2: '女' }[g] || '未知')
export const carTypeText = (t) => ({ 1: '特惠快车', 2: '快车', 3: '拼车' }[t] || `未知(${t})`)
export const roleText = (r) => ({ 1: '超级管理员', 2: '运营', 3: '客服' }[r] || `未知(${r})`)

// 通用：空值展示占位
export const orDash = (v) => (v === null || v === undefined || v === '' ? '-' : v)
