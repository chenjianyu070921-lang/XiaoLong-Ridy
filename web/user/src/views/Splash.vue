<template>
  <div class="splash-page">
    <div class="logo-container animate-fadeIn">
      <img src="/logo.png" alt="花小龙打车" class="logo-img" />
      <h1 class="app-name">花小龙</h1>
      <p class="app-slogan">打车更省钱</p>
    </div>
    
    <div class="action-buttons safe-area-bottom">
      <button class="btn-primary btn-login" @click="goToLogin">
        手机号登录
      </button>
      <button class="btn-success btn-register" @click="goToRegister">
        新用户注册
      </button>
    </div>

    <div class="agreement">
      <van-checkbox v-model="agreed" shape="square" icon-size="14px">
        <span class="agreement-text">
          我已阅读并同意
          <span class="link" @click.stop="showAgreement('user')">《用户协议》</span>、
          <span class="link" @click.stop="showAgreement('privacy')">《隐私政策》</span>
        </span>
      </van-checkbox>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const agreed = ref(true)

onMounted(() => {
  // 2秒后自动跳转到登录页（如果未登录）
  setTimeout(() => {
    // 可以在这里检查是否已登录
  }, 2000)
})

const goToLogin = () => {
  if (!agreed.value) {
    // 提示同意协议
    return
  }
  router.push('/login/phone')
}

const goToRegister = () => {
  if (!agreed.value) {
    return
  }
  router.push('/login/register')
}

const showAgreement = (type) => {
  console.log('Show agreement:', type)
}
</script>

<style scoped>
.splash-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #F5F3FF 0%, #FFFFFF 50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 24px;
}

.logo-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: auto;
}

.logo-img {
  width: 200px;
  height: 200px;
  object-fit: contain;
  margin-bottom: 20px;
}

.app-name {
  font-size: 36px;
  font-weight: bold;
  color: var(--primary-color);
  margin-bottom: 8px;
}

.app-slogan {
  font-size: 18px;
  color: var(--text-secondary);
}

.action-buttons {
  width: 100%;
  max-width: 360px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding-bottom: 40px;
}

.btn-login,
.btn-register {
  width: 100%;
  height: 48px;
  border-radius: 25px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  border: none;
}

.btn-login {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-light) 100%);
  color: white;
  box-shadow: 0 4px 15px rgba(124, 58, 237, 0.3);
}

.btn-register {
  background: linear-gradient(135deg, #10B981 0%, #059669 100%);
  color: white;
  box-shadow: 0 4px 15px rgba(16, 185, 129, 0.3);
}

.btn-login:active,
.btn-register:active {
  transform: scale(0.98);
  opacity: 0.9;
}

.agreement {
  width: 100%;
  max-width: 360px;
  text-align: center;
  margin-top: 24px;
  padding-bottom: 20px;
}

.agreement-text {
  font-size: 12px;
  color: var(--text-light);
  line-height: 1.5;
}

.link {
  color: var(--primary-color);
}
</style>
