export function buildDriverMineMenuItems() {
  return [
    { label: '个人资料', icon: 'user-o', color: '#2563eb', action: 'edit-profile' },
    { label: '车辆信息', icon: 'orders-o', color: '#f59e0b', route: '/mine/vehicle' },
    { label: '资质认证', icon: 'shield-o', color: '#10b981', route: '/mine/certification' },
    { label: '钱包提现', icon: 'balance-o', color: '#06b6d4', route: '/mine/wallet' },
    { label: '收益明细', icon: 'todo-list-o', color: '#ef4444', route: '/mine/income' },
    { label: '订单记录', icon: 'orders-o', color: '#4b5563', route: '/mine/orders' },
    { label: '评价与服务', icon: 'service-o', color: '#f97316', action: 'show-service' }
  ]
}

export function buildDriverMineServices() {
  return [
    { label: '联系客服', icon: 'service-o', color: '#2563eb', action: 'contact-service' },
    { label: '安全中心', icon: 'shield-o', color: '#10b981', action: 'safety-center' },
    { label: '帮助反馈', icon: 'question-o', color: '#f59e0b', action: 'help-center' },
    { label: '邀请有礼', icon: 'gift-o', color: '#ef4444', action: 'invite-friends' }
  ]
}
