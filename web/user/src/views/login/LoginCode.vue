<template>
  <div class="code-page">
    <div class="page-head">
      <button class="back-btn" type="button" @click="goBack">
        <van-icon name="arrow-left" size="20" color="#111827" />
      </button>
      <span class="title">输入验证码</span>
      <div class="head-spacer"></div>
    </div>

    <div class="code-body">
      <div class="intro-block">
        <p class="sub">验证码已发送至</p>
        <p class="phone">{{ maskedPhone }}</p>
      </div>

      <div class="code-grid" @click="focusInput">
        <div
          v-for="(item, index) in 6"
          :key="index"
          class="code-cell"
          :class="{ active: currentIndex === index, filled: code[index] }"
        >
          {{ code[index] || '' }}
        </div>
        <input
          ref="codeInput"
          v-model="code"
          type="tel"
          maxlength="6"
          class="hidden-input"
          inputmode="numeric"
          pattern="[0-9]*"
          @input="onInput"
        />
      </div>

      <div class="resend-area">
        <span v-if="countdown > 0" class="countdown">{{ countdown }}s 后重新获取</span>
        <button v-else class="resend-btn" type="button" @click="resendCode">重新获取验证码</button>
      </div>

      <button class="primary-btn" type="button" :disabled="code.length !== 6 || loading" @click="handleLogin">
        {{ loading ? '登录中...' : '登录' }}
      </button>

      <p class="service-link">
        遇到问题？<span @click="contactService">联系客服</span>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { showToast, showLoadingToast, closeToast } from 'vant'
import { sendSMSCode } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const phone = ref(String(route.query.phone || '').trim())
const code = ref('')
const codeInput = ref(null)
const countdown = ref(60)
const loading = ref(false)
let timer = null

const maskedPhone = computed(() => {
  if (!phone.value) return ''
  return phone.value.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
})

const currentIndex = computed(() => code.value.length)

const focusInput = () => {
  codeInput.value?.focus()
}

const onInput = (e) => {
  const value = e.target.value.replace(/\D/g, '').slice(0, 6)
  code.value = value
  if (value.length === 6) {
    setTimeout(() => {
      if (!loading.value) handleLogin()
    }, 250)
  }
}

const goBack = () => router.back()

const startCountdown = () => {
  countdown.value = 60
  timer = setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) clearInterval(timer)
  }, 1000)
}

const resendCode = async () => {
  try {
    await sendSMSCode(phone.value)
    showToast('验证码已重新发送')
    code.value = ''
    startCountdown()
  } catch (error) {
    console.error('Resend error:', error)
  }
}

const handleLogin = async () => {
  // 验证码输入满 6 位后会自动触发登录；此处防止用户同时点击按钮造成重复请求。
  if (loading.value) return
  if (code.value.length !== 6) {
    showToast('请输入6位验证码')
    return
  }

  try {
    loading.value = true
    showLoadingToast({
      message: '登录中...',
      forbidClick: true,
      duration: 0
    })

    // 统一走用户仓库登录，保证 token 和本地缓存同步，避免路由守卫拿不到登录态。
    await userStore.login(phone.value, code.value)

    closeToast()
    showToast('登录成功')
    await router.replace('/home')
  } catch (error) {
    console.error('登录失败:', error)
    closeToast()
    const message = error?.response?.data?.message || error?.message || '登录失败，请重试'
    showToast(message)
  } finally {
    loading.value = false
  }
}

const contactService = () => {
  showToast('客服电话：400-123-4567')
}

onMounted(() => {
  focusInput()
  startCountdown()
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.code-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #ffffff 0%, #faf7ff 100%);
  padding: 16px 20px 28px;
}

.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.back-btn {
  width: 36px;
  height: 36px;
  border: 0;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 4px 14px rgba(17, 24, 39, 0.08);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.title {
  font-size: 17px;
  font-weight: 700;
  color: #111827;
}

.head-spacer {
  width: 36px;
  height: 36px;
}

.code-body {
  min-height: calc(100vh - 68px);
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 22px;
}

.intro-block {
  text-align: center;
}

.sub {
  font-size: 13px;
  color: #9ca3af;
}

.phone {
  margin-top: 8px;
  font-size: 22px;
  font-weight: 800;
  color: #111827;
}

.code-grid {
  position: relative;
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 10px;
}

.code-cell {
  height: 54px;
  border: 1.5px solid #e5e7eb;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  font-weight: 700;
  color: #111827;
  background: #fff;
}

.code-cell.active {
  border-color: #7c3aed;
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.08);
}

.code-cell.filled {
  border-color: #7c3aed;
  background: #f5f3ff;
}

.hidden-input {
  position: absolute;
  inset: 0;
  opacity: 0;
}

.resend-area {
  min-height: 20px;
  text-align: center;
}

.countdown {
  font-size: 13px;
  color: #9ca3af;
}

.resend-btn {
  border: 0;
  background: transparent;
  color: #7c3aed;
  font-size: 13px;
  font-weight: 600;
}

.primary-btn {
  width: 100%;
  height: 48px;
  border: 0;
  border-radius: 16px;
  color: #ffffff;
  background: linear-gradient(135deg, #7c3aed 0%, #8b5cf6 100%);
  font-size: 16px;
  font-weight: 600;
  box-shadow: 0 10px 24px rgba(124, 58, 237, 0.24);
}

.primary-btn:disabled {
  background: #d1d5db;
  box-shadow: none;
}

.service-link {
  text-align: center;
  font-size: 13px;
  color: #9ca3af;
}

.service-link span {
  color: #7c3aed;
}
</style>
