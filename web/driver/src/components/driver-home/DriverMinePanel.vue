<template>
  <section class="group-panel">
    <header class="mine-hero">
      <div class="mine-avatar">{{ (driverStore.displayName || '司机').slice(0, 1) }}</div>
      <div class="mine-identity"><h2>{{ driverStore.displayName || '司机师傅' }}</h2><p>{{ driverStore.driver?.phone || '--' }}</p><span class="status-badge">{{ formatDriverStatus(driverStore.driver?.status) }}</span></div>
      <button type="button" class="mine-setting" aria-label="编辑资料" @click="$emit('edit-profile')"><van-icon name="edit" /></button>
    </header>
    <section class="income-card">
      <div><span>我的钱包（元）</span><strong>{{ formatPrice((incomeSummary?.totalIncomeCents || todayIncome?.totalIncomeCents || 0)) }}</strong><small>可提现收入 {{ formatPrice(incomeSummary?.withdrawableCents || 0) }}</small></div>
      <button type="button" @click="$emit('open-assets', 'wallet')">去提现 <van-icon name="arrow" /></button>
    </section>
    <section class="mine-orders section-block"><h3>订单概览 <button @click="$emit('open-orders')">全部订单 <van-icon name="arrow" /></button></h3><div class="order-stats"><span><van-icon name="orders-o" /><b>{{ orderStats.total }}</b><small>全部订单</small></span><span><van-icon name="clock-o" /><b>{{ orderStats.pending }}</b><small>待服务</small></span><span><van-icon name="location-o" /><b>{{ orderStats.serving }}</b><small>服务中</small></span><span><van-icon name="checked" /><b>{{ orderStats.done }}</b><small>已完成</small></span><span><van-icon name="close" /><b>{{ orderStats.cancelled }}</b><small>已取消</small></span></div></section>
    <section class="section-block"><h3>更多服务</h3><div class="tool-grid"><button v-for="tool in moreTools" :key="tool.label" type="button" @click="tool.action && $emit(tool.action)"><strong><van-icon :name="tool.icon" /></strong><span>{{ tool.label }}</span></button></div></section>
    <div class="mine-list">
      <button type="button" @click="$emit('open-assets', 'certification')">服务分 <span>{{ driverStore.driver?.serviceScore || '--' }} <i>›</i></span></button>
      <button type="button" @click="$emit('load-reviews')">乘客评价 <span><i>›</i></span></button>
      <button type="button" @click="$emit('refresh-dashboard')">听单检测 <span><i>›</i></span></button>
      <button type="button" @click="$emit('open-orders')">帮助中心 <span><i>›</i></span></button>
      <button type="button" @click="$emit('edit-profile')">设置 <span><i>›</i></span></button>
    </div>
    <!-- 资料、评价、轨迹子面板已按需求移除。 -->
    <!--
    <div class="panel-segments">
      <button type="button" :class="{ active: activeSection === 'profile' }" @click="activeSection = 'profile'">资料</button>
      <button type="button" :class="{ active: activeSection === 'reviews' }" @click="activeSection = 'reviews'">评价</button>
      <button type="button" :class="{ active: activeSection === 'trajectory' }" @click="activeSection = 'trajectory'">轨迹</button>
    </div>
    <DriverProfilePanel
      v-if="activeSection === 'profile'"
      :driver-store="driverStore"
      :format-driver-status="formatDriverStatus"
      @refresh-dashboard="$emit('refresh-dashboard')"
      @edit-profile="$emit('edit-profile')"
    />
    <DriverReviewsPanel
      v-else-if="activeSection === 'reviews'"
      :reviews="reviews"
      :format-time="formatTime"
      @load-reviews="$emit('load-reviews')"
    />
    <DriverTrajectoryPanel
      v-else
      :trajectory-order-id="trajectoryOrderId"
      :trajectory-error="trajectoryError"
      :trajectory-points="trajectoryPoints"
      :format-time="formatTime"
      @update:trajectory-order-id="$emit('update:trajectoryOrderId', $event)"
      @load-trajectory="$emit('load-trajectory')"
    />
    -->
  </section>
</template>

<script setup>
const props = defineProps({
  driverStore: { type: Object, required: true },
  reviews: { type: Array, required: true },
  trajectoryOrderId: { type: [String, Number], required: true },
  trajectoryError: { type: String, required: true },
  trajectoryPoints: { type: Array, required: true },
  formatDriverStatus: { type: Function, required: true },
  formatTime: { type: Function, required: true },
  defaultSection: { type: String, default: 'profile' }
  , incomeSummary: { type: Object, default: () => ({}) }, todayIncome: { type: Object, default: () => ({}) }, formatPrice: { type: Function, default: value => (Number(value || 0) / 100).toFixed(2) }
  , orderStats: { type: Object, default: () => ({ total: 0, pending: 0, serving: 0, done: 0, cancelled: 0 }) }
})

// 原型图中的工具入口，仅保留已有业务事件，未接入功能不执行跳转。
const moreTools = [
  { label: '车辆管理', icon: 'logistics', action: 'open-assets' }, { label: '费用明细', icon: 'balance-list-o' },
  { label: '发票中心', icon: 'description' }, { label: '银行卡', icon: 'card' }, { label: '邀请有奖', icon: 'friends-o' }
]

defineEmits([
  'refresh-dashboard',
  'edit-profile',
  'load-reviews',
  'update:trajectoryOrderId',
  'load-trajectory'
  , 'open-assets', 'open-orders'
])

</script>

<style scoped>
.group-panel { padding: 8px 0 24px; color: var(--driver-ink); }
.mine-hero { display:flex; align-items:center; gap:12px; padding:16px; border-radius:12px; background:#fff; box-shadow:0 4px 14px rgba(15,23,42,.06); }
.mine-avatar { width:56px; height:56px; border-radius:50%; background:#e8edff; color:var(--driver-primary); display:grid; place-items:center; font-size:24px; font-weight:800; }
.mine-identity { flex:1; min-width:0; }.mine-identity h2 { margin:0; font-size:19px; }.mine-identity p { margin:4px 0 5px; color:var(--driver-muted); font-size:12px; }
.status-badge { display:inline-flex; padding:3px 8px; border-radius:999px; background:#edf8f2; color:#14945b; font-size:11px; }
.mine-setting { display:grid; width:34px; height:34px; place-items:center; border:0; border-radius:50%; background:#f3f5f9; color:var(--driver-muted); font-size:17px; }
.income-card { display:flex; justify-content:space-between; align-items:center; margin:12px 0; padding:18px 16px; border-radius:12px; background:#172033; color:#fff; }
.income-card span,.income-card small { display:block; color:#b9c2d2; font-size:12px; }.income-card strong { display:block; font-size:28px; margin:6px 0; }.income-card button { display:flex; align-items:center; gap:3px; border:0; border-radius:8px; padding:9px 12px; background:#fff; color:#172033; font-weight:700; }
.section-block { margin:12px 0; padding:16px; border-radius:12px; background:#fff; box-shadow:0 4px 14px rgba(15,23,42,.05); }.section-block h3 { display:flex; justify-content:space-between; align-items:center; margin:0 0 14px; font-size:15px; }.section-block h3 button { display:flex; align-items:center; gap:2px; border:0; background:none; color:var(--driver-muted); font-size:12px; }
.order-stats { display:grid; grid-template-columns:repeat(5,1fr); gap:4px; }.order-stats span { display:grid; justify-items:center; gap:4px; color:var(--driver-muted); }.order-stats .van-icon { font-size:18px; color:var(--driver-primary); }.order-stats b { color:var(--driver-ink); font-size:17px; }.order-stats small { font-size:10px; white-space:nowrap; }
.tool-grid { display:grid; grid-template-columns:repeat(5,1fr); gap:15px 5px; }.tool-grid button { display:grid; justify-items:center; gap:6px; border:0; background:none; color:var(--driver-ink); font-size:11px; }.tool-grid strong { display:grid; width:38px; height:38px; place-items:center; border-radius:11px; background:#f1f4ff; color:var(--driver-primary); font-size:20px; }.tool-grid button:active strong { transform:scale(.94); }.mine-list { margin:12px 0; padding:0 16px; border-radius:12px; background:#fff; box-shadow:0 4px 14px rgba(15,23,42,.05); }.mine-list button { width:100%; display:flex; justify-content:space-between; padding:15px 0; border:0; border-bottom:1px solid var(--driver-line); background:transparent; text-align:left; color:var(--driver-ink); }.mine-list button:last-child { border-bottom:0; }.mine-list span { color:var(--driver-muted); }.mine-list i { font-style:normal; font-size:20px; margin-left:8px; }
@media (max-width:360px) { .order-stats small { font-size:9px; } }
</style>
