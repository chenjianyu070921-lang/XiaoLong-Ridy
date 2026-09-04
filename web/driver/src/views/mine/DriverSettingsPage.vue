<template>
  <main class="driver-settings-page">
    <header class="settings-header">
      <button type="button" class="back-button" aria-label="返回" @click="goBack">
        <van-icon name="arrow-left" />
      </button>
      <div>
        <span>设置</span>
        <h1>设置</h1>
      </div>
    </header>

    <section class="settings-group">
      <button type="button" class="settings-row" @click="openProfile">
        <span class="row-label"><van-icon name="manager-o" />个人信息</span>
        <span class="row-value">{{ driverStore.displayName || '司机师傅' }}<i>›</i></span>
      </button>
      <div class="settings-row">
        <span class="row-label"><van-icon name="bulb-o" />夜间模式</span>
        <van-switch v-model="darkMode" />
      </div>
    </section>

    <button type="button" class="logout-action" @click="logoutDriver">退出登录</button>
  </main>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { showConfirmDialog } from 'vant'
import { useDriverStore } from '@/stores/driver'

const router = useRouter()
const driverStore = useDriverStore()

const darkMode = computed({
  get: () => driverStore.darkMode,
  set: (value) => driverStore.setDarkMode(value)
})

function openProfile() {
  router.push('/profile/edit')
}

function goBack() {
  router.back()
}

async function logoutDriver() {
  try {
    await showConfirmDialog({ title: '退出登录', message: '确定要退出当前账号吗？' })
  } catch {
    return
  }
  driverStore.logout()
  router.replace('/login')
}
</script>

<style scoped>
.driver-settings-page {
  min-height: 100vh;
  padding: 16px 12px calc(24px + env(safe-area-inset-bottom));
  background: var(--driver-bg);
  color: var(--driver-ink);
}

.settings-header {
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
  background: var(--driver-card);
  color: var(--driver-ink);
  font-size: 20px;
  box-shadow: 0 4px 14px rgba(15, 23, 42, .08);
}

.settings-header span {
  color: var(--driver-muted);
  font-size: 12px;
}

.settings-header h1 {
  margin: 2px 0 0;
  font-size: 22px;
  line-height: 1.2;
}

.settings-group {
  margin-top: 12px;
  overflow: hidden;
  border-radius: 12px;
  background: var(--driver-card);
  box-shadow: 0 4px 14px rgba(15, 23, 42, .06);
}

.settings-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 16px;
  border: 0;
  border-bottom: 1px solid var(--driver-line);
  background: transparent;
  color: var(--driver-ink);
  text-align: left;
  font-size: 15px;
}

.settings-row:last-child {
  border-bottom: 0;
}

.row-label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.row-value {
  color: var(--driver-muted);
}

.row-value i {
  font-style: normal;
  font-size: 20px;
  margin-left: 8px;
}

.logout-action {
  width: 100%;
  margin-top: 24px;
  padding: 14px;
  border: 0;
  border-radius: 12px;
  background: var(--driver-card);
  color: var(--driver-danger);
  font-size: 16px;
  font-weight: 700;
  box-shadow: 0 4px 14px rgba(15, 23, 42, .06);
}
</style>
