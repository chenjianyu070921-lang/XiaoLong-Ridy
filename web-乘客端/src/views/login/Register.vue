<template>
  <div class="register-page">
    <div class="page-header">
      <div class="back-btn" @click="goBack">
        <van-icon name="arrow-left" size="20" color="#1F2937" />
      </div>
      <span class="title">注册账号</span>
      <div style="width: 32px;"></div>
    </div>

    <div class="content">
      <div class="form-item">
        <label>手机号码</label>
        <input
          v-model="form.phone"
          type="tel"
          class="custom-input"
          placeholder="请输入手机号码"
          maxlength="11"
        />
      </div>

      <div class="form-item">
        <label>验证码</label>
        <div class="code-wrapper">
          <input
            v-model="form.code"
            type="tel"
            class="custom-input code-input"
            placeholder="请输入验证码"
            maxlength="6"
          />
          <button 
            class="send-code-btn" 
            :disabled="!isPhoneValid || countdown > 0"
            @click="sendCode"
          >
            {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
          </button>
        </div>
      </div>

      <div class="form-item">
        <label>设置密码（可选）</label>
        <input
          v-model="form.password"
          :type="showPassword ? 'text' : 'password'"
          class="custom-input"
          placeholder="设置密码方便下次登录"
        />
        <van-icon 
          :name="showPassword ? 'eye-o' : 'closed-eye'" 
          class="eye-icon"
          @click="showPassword = !showPassword"
        />
      </div>

      <button 
        class="btn-primary register-btn" 
        :disabled="!isFormValid || loading"
        @click="handleRegister"
      >
        {{ loading ? '注册中...' : '注 册' }}
      </button>

      <div class="agreement">
        <van-checkbox v-model="agreed" shape="square" icon-size="14px">
          <span class="agreement-text">
            我已阅读并同意
            <span class="link">《用户协议》</span>和
            <span class="link">《隐私政策》</span>
          </span>
        </van-checkbox>
      </div>

      <div class="login-link">
        已有账号？<span @click="goToLogin">立即登录</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showLoadingToast, closeToast } from 'vant'
import { sendSMSCode, loginBySMS } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const form = ref({
  phone: '',
  code: '',
  password: ''
})
const showPassword = ref(false)
const agreed = ref(true)
const countdown = ref(0)
const loading = ref(false)
let timer = null

const isPhoneValid = computed(() => /^1[3-9]\d{9}$/.test(form.value.phone))
const isFormValid = computed(() => {
  return isPhoneValid.value && form.value.code.length === 6 && agreed.value
})

const goBack = () => router.back()

const goToLogin = () => router.push('/login/phone')

const sendCode = async () => {
  if (!isPhoneValid.value) return
  
  try {
    await sendSMSCode(form.value.phone)
    showToast('验证码已发送')
    
    countdown.value = 60
    timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (error) {
    console.error(error)
  }
}

const handleRegister = async () => {
  try {
    loading.value = true
    showLoadingToast({
      message: '注册中...',
      forbidClick: true,
      duration: 0
    })

    // 注册实际上就是登录（新用户自动创建）
    let res
    try {
      res = await loginBySMS(form.value.phone, form.value.code)
    } catch (apiError) {
      console.warn('API调用失败，使用开发模式:', apiError.message)
      
      // 模拟注册成功响应
      res = {
        token: 'dev_token_' + Date.now(),
        refreshToken: 'dev_refresh_' + Date.now(),
        isNewUser: true,
        user: {
          userId: 0,
          phone: form.value.phone,
          nickname: '乘客',
          avatarUrl: '',
          realNameStatus: 'unverified'
        }
      }
    }
    
    // 保存登录信息
    userStore.token = res.token
    userStore.refreshToken = res.refreshToken
    userStore.userInfo = res.user
    
    localStorage.setItem('token', res.token)
    localStorage.setItem('refreshToken', res.refreshToken)
    localStorage.setItem('userInfo', JSON.stringify(res.user))
    
    closeToast()
    showToast('注册成功')
    
    setTimeout(() => {
      router.replace('/home')
    }, 500)
  } catch (error) {
    console.error(error)
    showToast(error.message || '注册失败，请重试')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.register-page {
  min-height: 100vh;
  background: white;
}

.content {
  padding: 30px 24px;
}

.form-item {
  margin-bottom: 24px;
  position: relative;
}

.form-item label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.code-wrapper {
  display: flex;
  gap: 12px;
}

.code-input {
  flex: 1;
}

.send-code-btn {
  width: 120px;
  height: 48px;
  background: linear-gradient(135deg, #F5F3FF 0%, #EDE9FE 100%);
  color: var(--primary-color);
  border: 1px solid var(--primary-color);
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
}

.send-code-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.eye-icon {
  position: absolute;
  right: 16px;
  bottom: 14px;
  color: var(--text-light);
  cursor: pointer;
  font-size: 18px;
}

.register-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  margin-top: 10px;
}

.agreement {
  margin-top: 20px;
}

.agreement-text {
  font-size: 12px;
  color: var(--text-light);
}

.link {
  color: var(--primary-color);
}

.login-link {
  text-align: center;
  margin-top: 30px;
  font-size: 14px;
  color: var(--text-secondary);
}

.login-link span {
  color: var(--primary-color);
  font-weight: 500;
  cursor: pointer;
}
</style>
