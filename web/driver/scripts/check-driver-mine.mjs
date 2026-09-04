import assert from 'node:assert/strict'
import { buildDriverMineMenuItems, buildDriverMineServices } from '../src/components/driver-home/driver-mine-data.js'

const menuItems = buildDriverMineMenuItems({
  driverName: '测试司机',
  phone: '13800138000',
  onlineStatus: 1,
  serviceScore: 4.9,
  todayIncomeCents: 12850,
  vehicleStatus: 'VEHICLE_STATUS_NORMAL',
  certificationStatus: 2
})

assert.deepEqual(
  menuItems.map((item) => item.label),
  ['个人资料', '车辆信息', '资质认证', '钱包提现', '收益明细', '订单记录', '评价与服务']
)

assert.ok(menuItems.every((item) => !['接单设置', '轨迹', '评价', '历史轨迹'].includes(item.label)), 'menu should not contain old mine sections')

const services = buildDriverMineServices()
assert.deepEqual(
  services.map((item) => item.label),
  ['联系客服', '安全中心', '帮助反馈', '邀请有礼']
)

console.log('driver mine config checks passed')

