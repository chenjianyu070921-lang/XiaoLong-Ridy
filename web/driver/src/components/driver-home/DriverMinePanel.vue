<template>
  <section class="driver-mine-page">
    <header class="mine-header">
      <button type="button" class="mine-avatar" aria-label="编辑司机资料" @click="$emit('edit-profile')">
        <img v-if="avatarUrl" :src="avatarUrl" alt="司机头像" />
        <span v-else class="avatar-fallback">{{ driverInitial }}</span>
      </button>

      <div class="mine-copy">
        <p>司机端我的</p>
        <h1>{{ driverName }}</h1>
        <div class="mine-tags">
          <span>{{ statusText }}</span>
          <span>{{ onlineText }}</span>
        </div>
      </div>

      <button type="button" class="mine-settings" aria-label="编辑资料" @click="$emit('edit-profile')">
        <van-icon name="setting-o" />
      </button>
    </header>

    <div class="panel-segments mine-panel-segments">
      <button type="button" :class="{ active: activeSection === 'profile' }" @click="activeSection = 'profile'">资料</button>
      <button type="button" :class="{ active: activeSection === 'trajectory' }" @click="activeSection = 'trajectory'">检测范围</button>
    </div>

    <template v-if="activeSection === 'profile'">
      <section class="summary-band">
        <div class="summary-item">
          <span>今日收入</span>
          <strong>{{ todayIncomeText }}</strong>
        </div>
        <div class="summary-item">
          <span>服务分</span>
          <strong>{{ serviceScoreText }}</strong>
        </div>
        <div class="summary-item">
          <span>车辆状态</span>
          <strong>{{ vehicleText }}</strong>
        </div>
        <div class="summary-item">
          <span>资质状态</span>
          <strong>{{ certificationText }}</strong>
        </div>
      </section>

      <section class="menu-card">
        <button
          v-for="item in menuItems"
          :key="item.label"
          type="button"
          class="menu-item"
          @click="handleMenuClick(item)"
        >
          <div class="menu-left">
            <van-icon :name="item.icon" :size="20" :color="item.color" />
            <span>{{ item.label }}</span>
          </div>
          <div class="menu-right">
            <span v-if="item.extra" class="extra">{{ item.extra }}</span>
            <van-icon name="arrow" size="14" color="#9ca3af" />
          </div>
        </button>
      </section>

      <section class="service-card">
        <h3>其他服务</h3>
        <div class="service-grid">
          <button
            v-for="item in serviceItems"
            :key="item.label"
            type="button"
            class="service-item"
            @click="handleServiceClick(item)"
          >
            <van-icon :name="item.icon" :size="24" :color="item.color" />
            <span>{{ item.label }}</span>
          </button>
        </div>
      </section>

      <button class="logout-btn" type="button" @click="handleLogout">
        退出登录
      </button>
    </template>

    <DriverTrajectoryPanel
      v-else
      :trajectory-order-id="trajectoryOrderId"
      :trajectory-error="trajectoryError"
      :trajectory-points="trajectoryPoints"
      :format-time="formatTime"
      @update:trajectory-order-id="$emit('update:trajectoryOrderId', $event)"
      @load-trajectory="$emit('load-trajectory')"
    />
  </section>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showDialog, showToast } from 'vant'
import { buildDriverMineMenuItems, buildDriverMineServices } from './driver-mine-data'
import DriverTrajectoryPanel from './DriverTrajectoryPanel.vue'

const props = defineProps({
  driverStore: { type: Object, required: true },
  serviceScore: { type: [String, Number], default: '--' },
  todayIncome: { type: Object, default: () => ({}) },
  incomeSummary: { type: Object, default: () => ({}) },
  formatPrice: { type: Function, required: true },
  formatDriverStatus: { type: Function, required: true },
  formatVehicleStatus: { type: Function, required: true },
  formatCertificationStatus: { type: Function, required: true },
  trajectoryOrderId: { type: [String, Number], default: '' },
  trajectoryError: { type: String, default: '' },
  trajectoryPoints: { type: Array, default: () => [] },
  formatTime: { type: Function, required: true },
  defaultSection: { type: String, default: 'profile' }
})

const emit = defineEmits(['edit-profile', 'update:trajectoryOrderId', 'load-trajectory'])
const router = useRouter()
const activeSection = ref(props.defaultSection || 'profile')

watch(() => props.defaultSection, (value) => {
  activeSection.value = value || 'profile'
}, { immediate: true })

const avatarUrl = computed(() => props.driverStore.driver?.avatarUrl || '')
const driverName = computed(() => props.driverStore.displayName || '司机')
const driverInitial = computed(() => driverName.value.slice(0, 1) || '--')
const statusText = computed(() => props.formatDriverStatus(props.driverStore.driver?.status))
const onlineText = computed(() => {
  if (props.driverStore.tripPhase === 'trip') return '行程中'
  if (props.driverStore.onlineStatus === 1) return '在线接单'
  if (props.driverStore.onlineStatus === 2) return '接单后待接驾'
  return '当前离线'
})
const vehicleText = computed(() => props.formatVehicleStatus(props.driverStore.vehicle?.status))
const certificationText = computed(() => props.formatCertificationStatus(props.driverStore.certification?.auditStatus))
const serviceScoreText = computed(() => normalizeScore(props.serviceScore))
const incomeSummaryText = computed(() => formatMoney(props.incomeSummary?.totalIncomeCents))
const todayIncomeText = computed(() => formatMoney(props.todayIncome?.totalIncomeCents))
const menuItems = computed(() => {
  const items = buildDriverMineMenuItems()
  return items.map((item) => ({ ...item, extra: menuExtra(item.label) }))
})
const serviceItems = computed(() => buildDriverMineServices())

function formatMoney(cents) {
  const value = Number(cents ?? props.incomeSummary?.totalIncomeCents ?? 0)
  return props.formatPrice(Number.isFinite(value) ? value : 0)
}

function normalizeScore(value) {
  const score = Number(value)
  return Number.isFinite(score) ? score.toFixed(1) : String(value || '--')
}

function menuExtra(label) {
  switch (label) {
    case '个人资料':
      return driverName.value
    case '车辆信息':
      return vehicleText.value
    case '资质认证':
      return certificationText.value
    case '钱包提现':
      return todayIncomeText.value
    case '收益明细':
      return incomeSummaryText.value
    case '订单记录':
      return '查看全部'
    case '评价与服务':
      return '司机业务'
    default:
      return ''
  }
}

function handleMenuClick(item) {
  const target = item.route || item.path
  if (target) {
    router.push(target)
    return
  }
  switch (item.action) {
    case 'edit-profile':
      emit('edit-profile')
      break
    case 'show-service':
      showToast('评价和服务记录请在司机工作台查看')
      break
    default:
      break
  }
}

function handleServiceClick(item) {
  const messages = {
    '联系客服': '客服功能已接入',
    '安全中心': '安全中心已接入',
    '帮助反馈': '帮助反馈已接入',
    '邀请有礼': '邀请有礼已接入'
  }
  showToast(messages[item.label] || '功能已接入')
}

function handleLogout() {
  showDialog({
    title: '提示',
    message: '确定要退出登录吗？',
    showCancelButton: true,
    confirmButtonText: '退出',
    cancelButtonText: '取消'
  }).then(() => {
    props.driverStore.logout()
    router.replace('/login')
    showToast('已退出登录')
  }).catch(() => {})
}
</script>

<style scoped>
.driver-mine-page {
  display: grid;
  gap: 12px;
  padding: 10px 0 18px;
}

.mine-panel-segments {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.mine-header {
  display: grid;
  grid-template-columns: 52px minmax(0, 1fr) 36px;
  gap: 12px;
  align-items: center;
  padding: 16px;
  border-radius: 8px;
  background: linear-gradient(135deg, #0f172a, #2563eb);
  color: #fff;
}

.mine-avatar,
.mine-settings,
.service-item {
  border: 0;
  background: transparent;
}

.mine-avatar {
  width: 52px;
  height: 52px;
  padding: 0;
}

.mine-avatar img,
.avatar-fallback {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  object-fit: cover;
  background: rgba(255, 255, 255, 0.16);
  border: 2px solid rgba(255, 255, 255, 0.35);
}

.avatar-fallback {
  display: grid;
  place-items: center;
  font-size: 20px;
  font-weight: 800;
}

.mine-copy {
  min-width: 0;
}

.mine-copy p {
  margin: 0 0 4px;
  color: rgba(255, 255, 255, 0.76);
  font-size: 12px;
}

.mine-copy h1 {
  margin: 0;
  font-size: 22px;
  line-height: 1.2;
}

.mine-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.mine-tags span {
  padding: 4px 8px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  font-size: 12px;
}

.mine-settings {
  display: grid;
  place-items: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  color: #fff;
  background: rgba(255, 255, 255, 0.12);
}

.summary-band,
.menu-card,
.service-card {
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.06);
}

.summary-band {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  padding: 14px;
}

.summary-item {
  min-width: 0;
  padding: 10px 12px;
  border-radius: 8px;
  background: #f8fafc;
}

.summary-item span {
  display: block;
  color: #6b7280;
  font-size: 12px;
}

.summary-item strong {
  display: block;
  margin-top: 6px;
  color: #0f172a;
  font-size: 15px;
  line-height: 1.2;
}

.menu-card {
  overflow: hidden;
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 14px 16px;
  border: 0;
  border-bottom: 1px solid #f3f4f6;
  background: transparent;
}

.menu-item:last-child {
  border-bottom: 0;
}

.menu-left,
.menu-right {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.menu-left span {
  color: #111827;
  font-size: 14px;
  font-weight: 600;
}

.menu-right {
  justify-content: flex-end;
}

.extra {
  max-width: 130px;
  overflow: hidden;
  color: #6b7280;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.service-card {
  padding: 16px;
}

.service-card h3 {
  margin: 0 0 12px;
  color: #111827;
  font-size: 15px;
}

.service-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}

.service-item {
  display: grid;
  justify-items: center;
  gap: 8px;
  padding: 10px 0;
  color: #6b7280;
}

.service-item span {
  font-size: 12px;
  text-align: center;
}

.logout-btn {
  width: 100%;
  min-height: 48px;
  border: 1px solid #fecaca;
  border-radius: 8px;
  background: #fff;
  color: #dc2626;
  font-size: 15px;
  font-weight: 700;
}
</style>
