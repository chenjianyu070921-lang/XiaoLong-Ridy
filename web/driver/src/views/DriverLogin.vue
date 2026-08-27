<template>
  <div class="driver-login-page">
    <div class="login-hero">
      <img src="/logo.png" alt="花小龙" />
      <h1>司机端</h1>
      <p>移动接单工作台</p>
    </div>

    <van-tabs v-model:active="activeTab" shrink class="login-tabs">
      <van-tab title="密码登录">
        <van-form @submit="handlePasswordLogin">
          <van-field v-model="passwordForm.phone" name="phone" type="tel" label="手机号" placeholder="请输入手机号" />
          <van-field v-model="passwordForm.password" name="password" type="password" label="密码" placeholder="请输入密码" />
          <button class="primary-action" :disabled="loading" type="submit">
            {{ loading ? '登录中...' : '密码登录' }}
          </button>
        </van-form>
      </van-tab>

      <van-tab title="验证码登录">
        <van-form @submit="handleSMSLogin">
          <van-field v-model="smsForm.phone" name="phone" type="tel" label="手机号" placeholder="请输入手机号" />
          <van-field v-model="smsForm.code" name="code" type="tel" label="验证码" placeholder="请输入验证码">
            <template #button>
              <van-button size="small" type="primary" :disabled="smsCountdown > 0" @click="sendCode">
                {{ smsCountdown > 0 ? `${smsCountdown}s` : '发送' }}
              </van-button>
            </template>
          </van-field>
          <button class="primary-action" :disabled="loading" type="submit">
            {{ loading ? '登录中...' : '验证码登录' }}
          </button>
        </van-form>
      </van-tab>

      <van-tab title="注册">
        <van-form @submit="handleRegister">
          <van-field v-model="registerForm.phone" name="phone" type="tel" label="手机号" placeholder="请输入手机号" />
          <van-field v-model="registerForm.realName" name="realName" label="姓名" placeholder="请输入真实姓名" />
          <van-field v-model="registerForm.idCardNo" name="idCardNo" label="身份证号" placeholder="请输入身份证号" />
          <van-field v-model="registerForm.driverLicenseNo" name="driverLicenseNo" label="驾驶证号" placeholder="请输入驾驶证号" />
          <van-field v-model="registerForm.avatarUrl" name="avatarUrl" label="头像地址" placeholder="可选" />
          <van-field v-model="registerForm.password" name="password" type="password" label="密码" placeholder="设置登录密码" />
          <button class="primary-action" :disabled="loading" type="submit">
            {{ loading ? '提交中...' : '提交注册' }}
          </button>
        </van-form>
      </van-tab>
    </van-tabs>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { closeToast, showLoadingToast, showToast } from 'vant'
import { sendDriverSMSCode } from '@/api/driver'
import { useDriverStore } from '@/stores/driver'

const router = useRouter()
const driverStore = useDriverStore()

const activeTab = ref(0)
const loading = ref(false)
const smsCountdown = ref(0)
let smsTimer = null

const passwordForm = reactive({ phone: '', password: '' })
const smsForm = reactive({ phone: '', code: '' })
const registerForm = reactive({
  phone: '',
  realName: '',
  idCardNo: '',
  driverLicenseNo: '',
  avatarUrl: '',
  password: ''
})

function validatePhone(phone) {
  return /^1[3-9]\d{9}$/.test(phone)
}

async function sendCode() {
  if (!validatePhone(smsForm.phone)) {
    showToast('请输入正确的手机号')
    return
  }
  try {
    await sendDriverSMSCode(smsForm.phone, { silentError: true })
    showToast('验证码已发送')
  } catch (error) {
    showToast(apiErrorMessage(error, '验证码发送失败'))
    return
  }
  smsCountdown.value = 60
  smsTimer = window.setInterval(() => {
    smsCountdown.value -= 1
    if (smsCountdown.value <= 0) {
      window.clearInterval(smsTimer)
      smsTimer = null
    }
  }, 1000)
}

async function handlePasswordLogin() {
  if (!validatePhone(passwordForm.phone) || !passwordForm.password) {
    showToast('请输入手机号和密码')
    return
  }
  await submitLogin(() => driverStore.loginPassword(passwordForm.phone, passwordForm.password, { silentError: true }))
}

async function handleSMSLogin() {
  if (!validatePhone(smsForm.phone) || !smsForm.code) {
    showToast('请输入手机号和验证码')
    return
  }
  await submitLogin(() => driverStore.loginSMS(smsForm.phone, smsForm.code, { silentError: true }))
}

async function submitLogin(action) {
  try {
    loading.value = true
    showLoadingToast({ message: '登录中...', forbidClick: true, duration: 0 })
    await action()
    closeToast()
    showToast('登录成功')
    router.replace('/home')
  } catch (error) {
    closeToast()
    showToast(apiErrorMessage(error, '登录失败'))
  } finally {
    loading.value = false
  }
}

async function handleRegister() {
  if (!validatePhone(registerForm.phone) || !registerForm.realName || !registerForm.idCardNo || !registerForm.driverLicenseNo || !registerForm.password) {
    showToast('请填写完整注册信息')
    return
  }
  try {
    loading.value = true
    showLoadingToast({ message: '提交中...', forbidClick: true, duration: 0 })
    await driverStore.register({ ...registerForm }, { silentError: true })
    closeToast()
    showToast('注册成功，请等待资质审核')
    activeTab.value = 0
  } catch (error) {
    closeToast()
    showToast(apiErrorMessage(error, '注册失败'))
  } finally {
    loading.value = false
  }
}

function apiErrorMessage(error, fallbackMessage) {
  return error?.response?.data?.message || error?.message || fallbackMessage
}
</script>

<style scoped>
.driver-login-page {
  min-height: 100vh;
  background: #f5f7fb;
  padding-bottom: 20px;
}

.login-hero {
  padding: 16px 14px 14px;
  background: #6d4aff;
  color: #fff;
}

.login-hero img {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  display: block;
  margin: 6px auto 8px;
}

.login-hero h1 {
  text-align: center;
  font-size: 20px;
  line-height: 1.2;
}

.login-hero p {
  margin-top: 8px;
  text-align: center;
  color: rgba(255, 255, 255, 0.78);
}

.login-tabs {
  margin: 10px;
  overflow: hidden;
  border-radius: 12px;
  background: #fff;
}

.primary-action {
  width: calc(100% - 20px);
  height: 42px;
  margin: 12px 10px 14px;
  border: 0;
  border-radius: 12px;
  background: #6d4aff;
  color: #fff;
  font-size: 14px;
  font-weight: 700;
}

.primary-action:disabled {
  opacity: 0.65;
}
</style>
