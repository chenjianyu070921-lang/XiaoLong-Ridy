import axios from 'axios'
import request from './request'

// 获取乘客头像的七牛云客户端直传凭证。
export function getAvatarUploadToken(extension) {
  return request.post('/upload/avatar-token', { extension })
}

// 使用后端签发的临时凭证将图片直接上传到七牛云。
export function uploadToQiniu(uploadURL, token, key, file) {
  const formData = new FormData()
  formData.append('token', token)
  formData.append('key', key)
  formData.append('file', file)
  return axios.post(uploadURL, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 30000
  })
}
