<template>
  <div class="login-page">
    <div class="login-shell">
      <button class="back-btn" type="button" @click="goBack">
        <van-icon name="arrow-left" size="20" color="#111827" />
      </button>

      <div class="brand-block">
        <img class="brand-logo" src="/logo.png" alt="花小龙" />
        <h1>花小龙</h1>
        <p>打车更省钱</p>
      </div>

      <div class="form-card">
        <div class="input-row">
          <button class="country-code" type="button" @click="showCountryPicker = true">
            <span>{{ countryCode }}</span>
            <van-icon name="arrow-down" size="12" />
          </button>
          <input
            v-model="phone"
            type="tel"
            class="phone-input"
            placeholder="请输入手机号码"
            maxlength="11"
          />
          <button v-if="phone.length === 11" class="clear-btn" type="button" @click="phone = ''">
            <van-icon name="clear" size="16" />
          </button>
        </div>

        <p class="hint">未注册手机号验证后自动创建账号</p>

        <button class="primary-btn" type="button" :disabled="!isPhoneValid || loading" @click="sendCode">
          {{ loading ? '发送中...' : '获取验证码' }}
        </button>
      </div>

      <p class="agreement">
        登录即代表同意 <span>《用户协议》</span> 与 <span>《隐私政策》</span>
      </p>
    </div>

    <van-popup v-model:show="showCountryPicker" position="bottom" round>
      <van-picker
        :columns="countryCodes"
        @confirm="onCountryConfirm"
        @cancel="showCountryPicker = false"
      />
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showLoadingToast, closeToast } from 'vant'
import { sendSMSCode } from '@/api/auth'

const router = useRouter()
const phone = ref('')
const countryCode = ref('+86')
const showCountryPicker = ref(false)
const loading = ref(false)

const countryCodes = [
  { text: '+86 中国大陆', value: '+86' },
  { text: '+852 中国香港', value: '+852' },
  { text: '+886 中国台湾', value: '+886' }
]

const isPhoneValid = computed(() => /^1[3-9]\d{9}$/.test(phone.value))

const goBack = () => router.back()

const onCountryConfirm = ({ selectedOptions }) => {
  countryCode.value = selectedOptions[0]?.value || '+86'
  showCountryPicker.value = false
}

const sendCode = async () => {
  if (!isPhoneValid.value) {
    showToast('请输入正确的手机号')
    return
  }

  try {
    loading.value = true
    showLoadingToast({
      message: '发送中...',
      forbidClick: true,
      duration: 0
    })

    await sendSMSCode(phone.value)
    closeToast()
    showToast('验证码已发送')

    router.push({
      path: '/login/code',
      query: { phone: phone.value }
    })
  } catch (error) {
    console.error('Send code error:', error)
    closeToast()
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #ffffff 0%, #faf7ff 100%);
}

.login-shell {
  min-height: 100vh;
  padding: 18px 20px 28px;
  display: flex;
  flex-direction: column;
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

.brand-block {
  flex: 1;
  min-height: 250px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  text-align: center;
}

.brand-logo {
  width: 112px;
  height: 112px;
  object-fit: cover;
}

.brand-block h1 {
  font-size: 32px;
  line-height: 1;
  font-weight: 800;
  color: #111827;
}

.brand-block p {
  font-size: 16px;
  color: #6b7280;
}

.form-card {
  padding: 22px 18px 18px;
  border-radius: 24px;
  background: #ffffff;
  box-shadow: 0 12px 40px rgba(124, 58, 237, 0.12);
}

.input-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 14px;
  border-bottom: 1px solid #f3f4f6;
}

.country-code {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  border: 0;
  background: transparent;
  color: #111827;
  font-size: 16px;
  font-weight: 600;
}

.phone-input {
  flex: 1;
  min-width: 0;
  border: 0;
  outline: none;
  font-size: 18px;
  color: #111827;
  background: transparent;
}

.phone-input::placeholder {
  color: #9ca3af;
}

.clear-btn {
  width: 28px;
  height: 28px;
  border: 0;
  border-radius: 50%;
  background: #f3f4f6;
  color: #9ca3af;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.hint {
  margin-top: 12px;
  font-size: 13px;
  color: #9ca3af;
}

.primary-btn {
  width: 100%;
  height: 48px;
  margin-top: 18px;
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

.agreement {
  margin-top: 16px;
  padding-bottom: env(safe-area-inset-bottom);
  text-align: center;
  font-size: 12px;
  color: #9ca3af;
}

.agreement span {
  color: #7c3aed;
}
</style>
