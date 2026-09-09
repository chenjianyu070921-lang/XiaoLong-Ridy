import request from './request'

// 查询当前登录用户的钱包余额与流水，数据来源于 usersvc MySQL。
export function getWallet() { return request.post('/wallet', {}) }

// 充值由后端事务更新余额并写入充值流水。
export function rechargeWallet(amount) { return request.post('/wallet/recharge', { amount: Number(amount) }) }

// 提现由后端校验余额后事务扣减并写入提现流水。
export function withdrawWallet(amount) { return request.post('/wallet/withdraw', { amount: Number(amount) }) }

export function getWalletLedger() { return getWallet() }
