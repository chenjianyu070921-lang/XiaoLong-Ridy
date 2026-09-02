<template>
  <main class="driver-profile-edit-page">
    <header class="edit-header">
      <button type="button" class="back-button" aria-label="返回" @click="goBack">
        <van-icon name="arrow-left" />
      </button>
      <div>
        <span>司机资料</span>
        <h1>编辑个人信息</h1>
      </div>
    </header>

    <section class="avatar-edit-panel">
      <button type="button" class="avatar-picker" aria-label="更换头像" @click="chooseAvatar">
        <img v-if="avatarPreview" :src="avatarPreview" alt="司机头像" />
        <span v-else class="avatar-fallback">{{ driverStore.displayName.slice(0, 1) || '--' }}</span>
        <b><van-icon name="photograph" /></b>
      </button>
      <div>
        <strong>{{ driverStore.displayName || '司机' }}</strong>
        <p>{{ avatarChanged ? '已选择新头像，保存后生效' : '点击头像从本地图库选择照片' }}</p>
        <button type="button" class="avatar-gallery-button" @click="chooseAvatar">
          <van-icon name="photograph" />
          <span>从本地图库选择</span>
        </button>
      </div>
      <input ref="avatarInput" class="avatar-file-input" type="file" accept="image/*" @change="readAvatarFile" />
    </section>

    <van-form class="profile-edit-form" @submit="submitProfile">
      <van-field v-model="profileForm.realName" name="realName" label="姓名" placeholder="请输入司机姓名" clearable />
      <van-field v-model="profileForm.phone" name="phone" label="手机号" :placeholder="phonePlaceholder" clearable />
      <van-field v-model="profileForm.idCardNo" name="idCardNo" label="身份证号" :placeholder="idCardPlaceholder" clearable />
      <van-field v-model="profileForm.driverLicenseNo" name="driverLicenseNo" label="驾驶证号" placeholder="请输入驾驶证号" clearable />
      <button class="primary-action" type="submit" :disabled="submitting">
        {{ submitting ? '保存中...' : '保存资料' }}
      </button>
    </van-form>

    <section class="profile-status-panel">
      <p><b>司机ID</b><span>{{ driverStore.driverId || '--' }}</span></p>
      <p><b>账号状态</b><span>{{ formatDriverStatus(driverStore.driver.status) }}</span></p>
    </section>
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { closeToast, showLoadingToast, showToast } from 'vant'
import { uploadDriverAvatar } from '@/api/driver'
import { useDriverStore } from '@/stores/driver'

const router = useRouter()
const driverStore = useDriverStore()
const avatarInput = ref(null)
const submitting = ref(false)
const avatarBase64 = ref('')

const profileForm = reactive({
  realName: '',
  phone: '',
  idCardNo: '',
  driverLicenseNo: '',
  avatarUrl: ''
})

const avatarPreview = computed(() => avatarBase64.value || profileForm.avatarUrl || driverStore.driver.avatarUrl || '')
const avatarChanged = computed(() => Boolean(avatarBase64.value))
const phonePlaceholder = computed(() => driverStore.driver.phone || '请输入新的手机号')
const idCardPlaceholder = computed(() => driverStore.driver.idCardNo || '请输入新的身份证号')

onMounted(async () => {
  await driverStore.refreshProfile({ silentError: true }).catch(() => null)
  syncProfileForm()
})

function syncProfileForm() {
  profileForm.realName = driverStore.driver.realName || ''
  profileForm.phone = ''
  profileForm.idCardNo = ''
  profileForm.driverLicenseNo = driverStore.driver.driverLicenseNo || ''
  profileForm.avatarUrl = driverStore.driver.avatarUrl || ''
}

function chooseAvatar() {
  avatarInput.value?.click()
}

async function readAvatarFile(event) {
  const file = event.target.files?.[0]
  if (!file) return
  if (!file.type.startsWith('image/')) {
    showToast('请选择图片文件')
    event.target.value = ''
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    showToast('头像图片不能超过5MB')
    event.target.value = ''
    return
  }
  try {
    avatarBase64.value = await fileToBase64(file)
  } catch (error) {
    showToast(error?.message || '头像读取失败')
  } finally {
    event.target.value = ''
  }
}

async function submitProfile() {
  if (submitting.value) return
  submitting.value = true
  showLoadingToast({ message: '保存中...', forbidClick: true, duration: 0 })
  try {
    if (avatarBase64.value) {
      const uploaded = await uploadDriverAvatar(avatarBase64.value)
      profileForm.avatarUrl = uploaded.avatarUrl || ''
      avatarBase64.value = ''
    }
    await driverStore.saveProfile(compact({
      realName: profileForm.realName,
      phone: profileForm.phone,
      idCardNo: profileForm.idCardNo,
      driverLicenseNo: profileForm.driverLicenseNo,
      avatarUrl: profileForm.avatarUrl
    }))
    closeToast()
    showToast('资料已保存')
    syncProfileForm()
  } catch (error) {
    closeToast()
    showToast(error?.response?.data?.message || error?.message || '资料保存失败')
  } finally {
    submitting.value = false
  }
}

function goBack() {
  if (window.history.length > 1) router.back()
  else router.replace('/home')
}

function compact(payload) {
  return Object.fromEntries(Object.entries(payload).filter(([, value]) => String(value ?? '').trim() !== ''))
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('头像读取失败'))
    reader.readAsDataURL(file)
  })
}

function formatDriverStatus(status) {
  return {
    DRIVER_STATUS_PENDING: '待审核',
    DRIVER_STATUS_NORMAL: '正常',
    DRIVER_STATUS_FROZEN: '冻结',
    DRIVER_STATUS_CANCELLED: '注销'
  }[status] || status || '--'
}
</script>

<style scoped>
.driver-profile-edit-page {
  min-height: 100vh;
  padding: 16px 12px calc(24px + env(safe-area-inset-bottom));
  background: #f6f7fb;
  color: #172033;
}

.edit-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 4px 2px 16px;
}

.back-button {
  display: grid;
  width: 38px;
  height: 38px;
  flex: 0 0 38px;
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: #fff;
  color: #172033;
  font-size: 20px;
  box-shadow: 0 4px 14px rgba(15, 23, 42, .08);
}

.edit-header span {
  color: #7a8496;
  font-size: 12px;
}

.edit-header h1 {
  margin: 2px 0 0;
  font-size: 22px;
  line-height: 1.2;
}

.avatar-edit-panel,
.profile-status-panel {
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(15, 23, 42, .06);
}

.avatar-edit-panel {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px;
}

.avatar-picker {
  position: relative;
  width: 76px;
  height: 76px;
  flex: 0 0 76px;
  padding: 0;
  border: 0;
  border-radius: 50%;
  background: #eef2ff;
}

.avatar-picker img,
.avatar-fallback {
  width: 76px;
  height: 76px;
  border-radius: 50%;
  object-fit: cover;
}

.avatar-fallback {
  display: grid;
  place-items: center;
  color: #5b5cff;
  font-size: 26px;
  font-weight: 800;
}

.avatar-picker b {
  position: absolute;
  right: -2px;
  bottom: -2px;
  display: grid;
  width: 28px;
  height: 28px;
  place-items: center;
  border: 2px solid #fff;
  border-radius: 50%;
  background: #ffb72c;
  color: #fff;
  font-size: 15px;
}

.avatar-edit-panel strong {
  display: block;
  font-size: 18px;
  line-height: 1.2;
}

.avatar-edit-panel p {
  margin: 6px 0 0;
  color: #7a8496;
  font-size: 12px;
  line-height: 1.45;
}

.avatar-gallery-button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 34px;
  margin-top: 10px;
  padding: 0 12px;
  border: 0;
  border-radius: 999px;
  background: #5b5cff;
  color: #fff;
  font-size: 13px;
  font-weight: 800;
}

.avatar-file-input {
  display: none;
}

.profile-edit-form {
  margin-top: 12px;
  overflow: hidden;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(15, 23, 42, .06);
}

.profile-edit-form :deep(.van-cell) {
  min-height: 52px;
}

.primary-action {
  width: calc(100% - 28px);
  min-height: 48px;
  margin: 14px;
  border: 0;
  border-radius: 999px;
  background: #5b5cff;
  color: #fff;
  font-size: 16px;
  font-weight: 800;
}

.profile-status-panel {
  display: grid;
  gap: 10px;
  margin-top: 12px;
  padding: 16px 18px;
}

.profile-status-panel p {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin: 0;
  color: #7a8496;
  font-size: 13px;
}

.profile-status-panel b {
  color: #172033;
}

.profile-status-panel span {
  min-width: 0;
  overflow: hidden;
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;
}

button:disabled {
  opacity: .48;
}
</style>
