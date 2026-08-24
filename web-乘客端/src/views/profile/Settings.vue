<template>
  <div class="settings-page">
    <!-- 顶部导航 -->
    <div class="page-header">
      <van-icon name="arrow-left" size="20" @click="goBack" />
      <span class="title">设置</span>
      <div style="width: 20px;"></div>
    </div>

    <!-- 账号安全 -->
    <div class="section">
      <h3 class="section-title">账号与安全</h3>
      <div class="menu-list">
        <div class="menu-item" @click="goToPage('phone')">
          <div class="left">
            <van-icon name="phone-o" size="20" color="#7C3AED" />
            <span>账号与密码</span>
          </div>
          <div class="right">
            <span class="value">{{ maskedPhone }}</span>
            <van-icon name="arrow" size="14" color="#9CA3AF" />
          </div>
        </div>

        <div class="menu-item" @click="goToPage('realname')">
          <div class="left">
            <van-icon name="idcard" size="20" color="#10B981" />
            <span>实名认证</span>
          </div>
          <div class="right">
            <span class="status" :class="realNameStatus">{{ realNameText }}</span>
            <van-icon name="arrow" size="14" color="#9CA3AF" />
          </div>
        </div>

        <div class="menu-item" @click="showChangePhone = true">
          <div class="left">
            <van-icon name="exchange" size="20" color="#F59E0B" />
            <span>修改手机号</span>
          </div>
          <div class="right">
            <van-icon name="arrow" size="14" color="#9CA3AF" />
          </div>
        </div>
      </div>
    </div>

    <!-- 通知设置 -->
    <div class="section">
      <h3 class="section-title">通知设置</h3>
      <div class="menu-list">
        <div class="menu-item">
          <div class="left">
            <van-icon name="bell" size="20" color="#EF4444" />
            <span>推送通知</span>
          </div>
          <div class="right">
            <van-switch v-model="settings.pushNotification" size="22" />
          </div>
        </div>

        <div class="menu-item">
          <div class="left">
            <van-icon name="volume-o" size="20" color="#8B5CF6" />
            <span>声音提醒</span>
          </div>
          <div class="right">
            <van-switch v-model="settings.soundAlert" size="22" />
          </div>
        </div>

        <div class="menu-item">
          <div class="left">
            <van-icon name="envelop-o" size="20" color="#3B82F6" />
            <span>短信通知</span>
          </div>
          <div class="right">
            <van-switch v-model="settings.smsNotify" size="22" />
          </div>
        </div>
      </div>
    </div>

    <!-- 隐私设置 -->
    <div class="section">
      <h3 class="section-title">隐私设置</h3>
      <div class="menu-list">
        <div class="menu-item">
          <div class="left">
            <van-icon name="location-o" size="20" color="#DC2626" />
            <span>位置信息</span>
          </div>
          <div class="right">
            <span class="value">已开启</span>
            <van-icon name="arrow" size="14" color="#9CA3AF" />
          </div>
        </div>

        <div class="menu-item">
          <div class="left">
            <van-icon icon="shield-o" size="20" color="#059669" />
            <span>隐私政策</span>
          </div>
          <div class="right">
            <van-icon name="arrow" size="14" color="#9CA3AF" />
          </div>
        </div>

        <div class="menu-item">
          <div class="left">
            <van-icon name="description" size="20" color="#6366F1" />
            <span>用户协议</span>
          </div>
          <div class="right">
            <van-icon name="arrow" size="14" color="#9CA3AF" />
          </div>
        </div>
      </div>
    </div>

    <!-- 其他 -->
    <div class="section">
      <h3 class="section-title">其他</h3>
      <div class="menu-list">
        <div class="menu-item" @click="clearCache">
          <div class="left">
            <van-icon name="delete-o" size="20" color="#9CA3AF" />
            <span>清除缓存</span>
          </div>
          <div class="right">
            <span class="value">{{ cacheSize }}</span>
            <van-icon name="arrow" size="14" color="#9CA3AF" />
          </div>
        </div>

        <div class="menu-item" @click="checkUpdate">
          <div class="left">
            <van-icon name="upgrade" size="20" color="#0EA5E9" />
            <span>检查更新</span>
          </div>
          <div class="right">
            <span class="value">v1.0.0</span>
            <van-icon name="arrow" size="14" color="#9CA3AF" />
          </div>
        </div>

        <div class="menu-item" @click="aboutUs">
          <div class="left">
            <van-icon name="info-o" size="20" color="#7C3AED" />
            <span>关于花小龙</span>
          </div>
          <div class="right">
            <van-icon name="arrow" size="14" color="#9CA3AF" />
          </div>
        </div>
      </div>
    </div>

    <!-- 退出登录 -->
    <button class="logout-btn safe-area-bottom" @click="handleLogout">
      退出登录
    </button>

    <!-- 修改手机号弹窗 -->
    <van-dialog
      v-model:show="showChangePhone"
      title="修改手机号"
      show-cancel-button
      confirm-button-text="下一步"
      @confirm="verifyOldPhone"
    >
      <div class="dialog-content">
        <p>请输入原手机号验证码</p>
        <input 
          v-model="oldPhoneCode"
          type="tel"
          placeholder="请输入验证码"
          maxlength="6"
          class="code-input"
        />
      </div>
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showDialog, showLoadingToast, closeToast } from 'vant'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

// 状态
const cacheSize = ref('12.5MB')
const showChangePhone = ref(false)
const oldPhoneCode = ref('')

// 设置项
const settings = ref({
  pushNotification: true,
  soundAlert: true,
  smsNotify: false
})

// 实名认证状态
const realNameStatus = computed(() => {
  // 根据用户信息判断
  return userStore.userInfo.realNameStatus === 'verified' ? 'verified' : 'unverified'
})

const realNameText = computed(() => {
  return realNameStatus.value === 'verified' ? '已认证' : '未认证'
})

// 手机号脱敏
const maskedPhone = computed(() => {
  const phone = userStore.phone || '138****5678'
  return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
})

// 方法
const goBack = () => router.back()

const goToPage = (page) => {
  const routes = {
    phone: '/settings/phone',
    realname: '/settings/realname'
  }
  
  if (page === 'realname' && realNameStatus.value === 'verified') {
    showToast('您已完成实名认证')
    return
  }
  
  if (routes[page]) {
    // 可以跳转到对应页面或显示弹窗
    showToast(`${page}功能开发中`)
  }
}

const verifyOldPhone = () => {
  if (oldPhoneCode.value.length !== 6) {
    showToast('请输入6位验证码')
    return false
  }
  showToast('验证成功，请输入新手机号')
}

const clearCache = async () => {
  showDialog({
    title: '提示',
    message: '确定要清除缓存吗？',
    showCancelButton: true
  }).then(async () => {
    const toast = showLoadingToast({
      message: '正在清除...',
      forbidClick: true,
      duration: 0
    })
    
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    closeToast()
    cacheSize.value = '0KB'
    showToast('缓存已清除')
  }).catch(() => {})
}

const checkUpdate = () => {
  showLoadingToast({
    message: '检查中...',
    duration: 800
  }).then(() => {
    showToast('当前已是最新版本')
  })
}

const aboutUs = () => {
  showDialog({
    title: '关于花小龙打车',
    message: '版本：v1.0.0\n\n花小龙打车 - 您的出行好伙伴\n让每一次出行都更便捷、更实惠',
    confirmButtonText: '知道了'
  })
}

const handleLogout = () => {
  showDialog({
    title: '提示',
    message: '确定要退出登录吗？',
    showCancelButton: true,
    confirmButtonText: '确定退出'
  }).then(() => {
    userStore.logout()
    showToast('已退出登录')
    router.replace('/splash')
  }).catch(() => {})
}
</script>

<style scoped>
.settings-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 80px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: white;
  position: sticky;
  top: 0;
  z-index: 10;
}

.title {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
}

.section {
  margin-top: 12px;
}

.section-title {
  font-size: 13px;
  color: #9CA3AF;
  padding: 12px 16px 8px;
  font-weight: 500;
}

.menu-list {
  background: white;
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #F9FAFB;
  cursor: pointer;
  transition: background 0.2s;
}

.menu-item:last-child {
  border-bottom: none;
}

.menu-item:active {
  background: #F9FAFB;
}

.left {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 15px;
  color: var(--text-primary);
}

.right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.value {
  font-size: 13px;
  color: #9CA3AF;
}

.status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
}

.status.verified {
  background: #ECFDF5;
  color: #059669;
}

.status.unverified {
  background: #FEF3C7;
  color: #D97706;
}

.logout-btn {
  width: calc(100% - 32px);
  margin: 24px 16px 30px;
  height: 48px;
  background: white;
  border: 1px solid #D1D5DB;
  border-radius: 24px;
  font-size: 16px;
  color: #DC2626;
  cursor: pointer;
  transition: all 0.2s;
}

.logout-btn:active {
  background: #FEF2F2;
}

.dialog-content {
  padding: 20px 0;
  text-align: center;
}

.dialog-content p {
  font-size: 14px;
  color: #6B7280;
  margin-bottom: 16px;
}

.code-input {
  width: calc(100% - 32px);
  height: 44px;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  padding: 0 16px;
  font-size: 16px;
  text-align: center;
  outline: none;
}
</style>
