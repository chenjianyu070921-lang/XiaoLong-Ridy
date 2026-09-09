import request from './request'

// 查询当前乘客保存的常用地址列表。
export function listAddresses() {
  return request.post('/addresses/list', {})
}

// 新增常用地址，保存联系人、地址坐标和默认地址设置。
export function createAddress(data) {
  return request.post('/addresses/create', data)
}

// 更新指定常用地址，后端会校验地址是否属于当前乘客。
export function updateAddress(data) {
  return request.post('/addresses/update', data)
}

// 删除指定常用地址，后端按当前登录乘客隔离数据。
export function deleteAddress(id) {
  return request.post('/addresses/delete', { id })
}
