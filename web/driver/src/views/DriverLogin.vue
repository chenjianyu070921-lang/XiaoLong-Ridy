<template>
  <main class="driver-login-page">
    <section class="login-hero">
      <img src="/logo.png" alt="花小龙打车" />
      <div>
        <p>花小龙打车</p>
        <h1>接单工作台</h1>
      </div>
    </section>

    <section class="login-card">
      <div class="auth-mode-tabs">
        <button
          v-for="item in authTabs"
          :key="item.value"
          type="button"
          :class="{ active: activeTab === item.value }"
          @click="activeTab = item.value"
        >
          {{ item.label }}
        </button>
      </div>

      <van-form v-if="activeTab === 0" class="auth-form" @submit="handlePasswordLogin">
        <van-field v-model="passwordForm.phone" name="phone" type="tel" label="手机号" placeholder="请输入手机号" clearable />
        <van-field v-model="passwordForm.password" name="password" type="password" label="密码" placeholder="请输入密码" clearable />
        <button class="primary-action" :disabled="loading" type="submit">
          {{ loading ? '登录中...' : '账号登录' }}
        </button>
      </van-form>

      <van-form v-else-if="activeTab === 1" class="auth-form" @submit="handleSMSLogin">
        <van-field v-model="smsForm.phone" name="phone" type="tel" label="手机号" placeholder="请输入手机号" clearable>
          <template #button>
            <van-button size="small" type="primary" native-type="button" :disabled="smsCountdown > 0" @click="sendCode">
              {{ smsCountdown > 0 ? smsCountdown + 's' : '获取' }}
            </van-button>
          </template>
        </van-field>
        <van-field v-model="smsForm.code" name="code" type="tel" label="验证码" placeholder="请输入验证码" clearable />
        <button class="primary-action" :disabled="loading" type="submit">
          {{ loading ? '登录中...' : '验证码登录' }}
        </button>
      </van-form>

      <van-form v-else class="auth-form" @submit="handleRegister">
        <van-field v-model="registerForm.phone" name="phone" type="tel" label="手机号" placeholder="请输入手机号" clearable />
        <van-field v-model="registerForm.realName" name="realName" label="姓名" placeholder="请输入真实姓名" clearable />
        <van-field v-model="registerForm.idCardNo" name="idCardNo" label="身份证号" placeholder="请输入身份证号" clearable />
        <van-field v-model="registerForm.driverLicenseNo" name="driverLicenseNo" label="驾驶证号" placeholder="请输入驾驶证号" clearable />
        <van-field v-model="registerForm.password" name="password" type="password" label="密码" placeholder="请设置登录密码" clearable />
        <button class="primary-action" :disabled="loading" type="submit">
          {{ loading ? '提交中...' : '提交注册' }}
        </button>
      </van-form>
    </section>
  </main>
</template>

<script setup>
import { onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { closeToast, showLoadingToast, showToast } from 'vant'
import { sendDriverSMSCode } from '@/api/driver'
import { useDriverStore } from '@/stores/driver'
import { apiErrorMessage } from '@/utils/safe-request'

const router = useRouter()
const driverStore = useDriverStore()

const authTabs = [
  { label: '账号', value: 0 },
  { label: '验证码', value: 1 },
  { label: '注册', value: 2 }
]

const activeTab = ref(0)
const loading = ref(false)
const smsCountdown = ref(0)
let smsTimer = null

const driverPhoneRegexp = /^(?:1[3-9]\d{9}|\d{12,15})$/

const passwordForm = reactive({ phone: '', password: '' })
const smsForm = reactive({ phone: '', code: '' })
const registerForm = reactive({
  phone: '',
  realName: '',
  idCardNo: '',
  driverLicenseNo: '',
  password: ''
})

onUnmounted(() => {
  if (smsTimer) window.clearInterval(smsTimer)
})

function normalizePhone(phone) {
  return String(phone || '').trim()
}

function validatePhone(phone) {
  return driverPhoneRegexp.test(normalizePhone(phone))
}

function validatePassword(password) {
  const value = String(password || '')
  return value.length >= 8 && value.length <= 72
}

async function sendCode() {
  const phone = normalizePhone(smsForm.phone)
  if (!validatePhone(phone)) {
    showToast('请输入正确的手机号')
    return
  }
  try {
    await sendDriverSMSCode(phone, { silentError: true })
    showToast('验证码已发送（联调验证码见服务端日志）')
  } catch (error) {
    showToast(apiErrorMessage(error, '验证码发送失败'))
    return
  }

  smsCountdown.value = 60
  if (smsTimer) window.clearInterval(smsTimer)
  smsTimer = window.setInterval(() => {
    smsCountdown.value -= 1
    if (smsCountdown.value <= 0) {
      window.clearInterval(smsTimer)
      smsTimer = null
    }
  }, 1000)
}

async function handlePasswordLogin() {
  const phone = normalizePhone(passwordForm.phone)
  if (!validatePhone(phone) || !passwordForm.password) {
    showToast('请输入手机号和密码')
    return
  }
  await submitLogin(() => driverStore.loginPassword(phone, passwordForm.password, { silentError: true }))
}

async function handleSMSLogin() {
  const phone = normalizePhone(smsForm.phone)
  if (!validatePhone(phone) || !smsForm.code) {
    showToast('请输入手机号和验证码')
    return
  }
  await submitLogin(() => driverStore.loginSMS(phone, smsForm.code, { silentError: true }))
}

async function submitLogin(action) {
  try {
    loading.value = true
    showLoadingToast({ message: '登录中...', forbidClick: true, duration: 0 })
    const res = await action()
    closeToast()
    if (String(res?.driver?.status || '').toUpperCase().includes('PENDING')) {
      showToast('账号待审核，可补充车辆和资质信息，审核通过后可听单')
    } else {
      showToast('登录成功')
    }
    router.replace('/home')
  } catch (error) {
    closeToast()
    showToast(apiErrorMessage(error, '登录失败'))
  } finally {
    loading.value = false
  }
}

async function handleRegister() {
  const phone = normalizePhone(registerForm.phone)
  if (!validatePhone(phone) || !registerForm.realName || !registerForm.idCardNo || !registerForm.driverLicenseNo || !registerForm.password) {
    showToast('请填写完整注册信息')
    return
  }
  if (!validatePassword(registerForm.password)) {
    showToast('密码长度需为 8-72 位')
    return
  }

  try {
    loading.value = true
    showLoadingToast({ message: '提交中...', forbidClick: true, duration: 0 })
    await driverStore.register({ ...registerForm, phone }, { silentError: true })
    closeToast()
    showToast('注册成功，请等待管理员审核')
    activeTab.value = 0
  } catch (error) {
    closeToast()
    showToast(apiErrorMessage(error, '注册失败'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.driver-login-page {
  min-height: 100vh;
  padding: 18px 12px 24px;
  background: #eef2f7;
}

.login-hero {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 148px;
  padding: 22px 16px;
  border-radius: 0 0 22px 22px;
  background: linear-gradient(135deg, #6d4aff, #4b2bc5);
  color: #fff;
}

.login-hero img {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  background: #fff;
  object-fit: cover;
}

.login-hero p {
  margin: 0 0 5px;
  color: rgba(255, 255, 255, .78);
  font-size: 13px;
}

.login-hero h1 {
  margin: 0;
  font-size: 25px;
  line-height: 1.15;
  letter-spacing: 0;
}

.login-card {
  margin-top: -22px;
  padding: 12px;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 14px 34px rgba(77, 48, 160, .16);
}

.auth-mode-tabs {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
  padding: 4px;
  border-radius: 10px;
  background: #f1f5fb;
}

.auth-mode-tabs button {
  min-height: 38px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: #667085;
  font-size: 14px;
  font-weight: 800;
}

.auth-mode-tabs button.active {
  background: #6d4aff;
  color: #fff;
  box-shadow: 0 6px 14px rgba(109, 74, 255, .22);
}

.auth-form {
  padding-top: 12px;
}

.driver-login-page :deep(.van-cell) {
  padding: 12px 4px;
}

.driver-login-page :deep(.van-field__label) {
  width: 72px;
  color: #667085;
  font-size: 13px;
}

.primary-action {
  width: 100%;
  min-height: 48px;
  margin-top: 16px;
  border: 0;
  border-radius: 10px;
  background: #6d4aff;
  color: #fff;
  font-size: 16px;
  font-weight: 800;
}

.primary-action:disabled {
  opacity: .62;
}
</style>
