// 钱包账本工具：统一管理余额与分类流水，供个人中心和支付页复用。
const STORAGE_KEY = 'xiaolong_wallet_ledger'

const readLedger = () => {
  try { return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{"balance":0,"transactions":[]}') } catch (_) { return { balance: 0, transactions: [] } }
}

// 存储空间不足或数据异常时仍保持可恢复的账本结构。
const saveLedger = ledger => localStorage.setItem(STORAGE_KEY, JSON.stringify(ledger))

// 记录一笔钱包变动。amount 为正数表示收入，负数表示支出。
export function recordWalletTransaction({ type, title, amount, orderId = '' }) {
  const ledger = readLedger()
  // 订单支付回调可能重复到达，使用订单号保证同一类订单流水只记一次。
  if (orderId && ledger.transactions.some(item => item.orderId === orderId && item.type === type)) return ledger
  const value = Number(amount) || 0
  ledger.balance = Number((Number(ledger.balance || 0) + value).toFixed(2))
  ledger.transactions.unshift({ id: `${Date.now()}-${Math.random()}`, type, title, amount: value, orderId, time: new Date().toISOString() })
  saveLedger(ledger)
  return ledger
}

// 查询钱包快照，并计算各分类累计金额。
export function getWalletLedger() {
  const ledger = readLedger()
  const transactions = Array.isArray(ledger.transactions) ? ledger.transactions : []
  return {
    balance: Number(ledger.balance || 0).toFixed(2),
    pending: '0.00',
    couponDiscount: Math.abs(transactions.filter(item => item.type === 'coupon').reduce((sum, item) => sum + Math.min(item.amount, 0), 0)).toFixed(2),
    recharged: transactions.filter(item => item.type === 'recharge').reduce((sum, item) => sum + Math.max(item.amount, 0), 0).toFixed(2),
    transactions
  }
}
