import axios from 'axios'
import driverRequest from '@/api/request'

export function getAvatarUploadToken(extension) {
  return driverRequest.post('/upload/avatar-token', { extension })
}

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
