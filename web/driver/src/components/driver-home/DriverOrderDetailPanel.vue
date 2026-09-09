<template>
  <van-popup
    :show="visible"
    position="bottom"
    teleport="#driver-home-popups"
    class="order-detail"
    :style="{ left: '0', right: '0', width: 'min(100vw, 390px)', height: '100%', margin: '0 auto', borderRadius: 0 }"
    @update:show="(v) => emit('update:visible', v)"
  >
    <!-- 模块1：顶部标题栏（保留状态渐变头部 + 一口价标签） -->
    <header class="od-header" :class="statusClass">
      <button type="button" class="od-close" aria-label="关闭" @click="close">
        <van-icon name="cross" />
      </button>
      <div class="od-header-text">
        <div class="od-title-row">
          <strong class="od-status">{{ statusLabel }}</strong>
          <span v-if="orderTypeTag" class="od-type-tag">{{ orderTypeTag }}</span>
        </div>
        <small class="od-orderno">订单号 {{ order?.orderNo || '--' }}</small>
      </div>
    </header>

    <div class="od-body">
      <!-- 模块2：预估到手收入卡片（黄色高亮，非红色） -->
      <section class="od-income-card">
        <span class="od-income-label">本单预估到手收入</span>
        <strong class="od-income-value">{{ incomeText }}</strong>
        <small class="od-income-note">行程结束后按实际里程结算，已扣除平台服务费</small>
      </section>

      <!-- 模块3：行程起止地址卡片（上车点/目的地 + 乘客备注） -->
      <section class="od-route">
        <div class="od-route-row">
          <div class="od-dot from"></div>
          <div class="od-route-meta">
            <span class="od-tag">上车点</span>
            <strong>{{ order?.fromAddress || '--' }}</strong>
            <small v-if="order?.passengerRemark" class="od-remark">乘客备注：{{ order.passengerRemark }}</small>
          </div>
        </div>
        <div class="od-route-line" aria-hidden="true"></div>
        <div class="od-route-row">
          <div class="od-dot to"></div>
          <div class="od-route-meta">
            <span class="od-tag">目的地</span>
            <strong>{{ order?.toAddress || '--' }}</strong>
          </div>
        </div>
        <span v-if="isPickedUp" class="od-picked-flag">
          <van-icon name="checked" /> 已接到乘客
        </span>
      </section>

      <!-- 模块4：行程信息网格（2x2 + 接驾/剩余距离） -->
      <section class="od-metrics">
        <div class="od-metric">
          <span>车型</span>
          <strong>{{ carTypeLabel }}</strong>
        </div>
        <div class="od-metric">
          <span>预估里程</span>
          <strong>{{ distanceText }} km</strong>
        </div>
        <div class="od-metric">
          <span>预计时长</span>
          <strong>{{ durationText }}</strong>
        </div>
        <div class="od-metric">
          <span>下单时间</span>
          <strong>{{ createTimeText }}</strong>
        </div>
        <template v-if="hasPickupDistance || hasRemainMileage">
          <div class="od-metric">
            <span>接驾距离</span>
            <strong>{{ pickupDistanceText }}</strong>
          </div>
          <div class="od-metric">
            <span>剩余距离</span>
            <strong>{{ remainMileageText }}</strong>
          </div>
        </template>
      </section>

      <!-- 模块5：操作功能（联系乘客 / 上报问题） -->
      <section class="od-actions">
        <button type="button" class="od-action" @click="onContact">
          <van-icon name="phone-o" />
          <span>联系乘客</span>
        </button>
        <button type="button" class="od-action" @click="reportVisible = true">
          <van-icon name="warning-o" />
          <span>上报问题</span>
        </button>
      </section>

      <!-- 模块6：安全提示（弱辅助文本） -->
      <p class="od-safety">行车注意安全，如遇危险可一键求助；行程中请勿操作手机。</p>

      <!-- 取消 / 退款原因 -->
      <section v-if="showCancelBlock" class="od-cancel">
        <van-icon name="warning-o" />
        <div>
          <strong>订单已{{ Number(order?.status) === 7 ? '退款' : '取消' }}</strong>
          <span>{{ cancelActorLabel }}：{{ order?.cancelReason || '未填写原因' }}</span>
        </div>
      </section>
    </div>

    <!-- 聊天入口：司机端与乘客沟通（仅活跃订单） -->
    <button v-if="canChat" type="button" class="od-chat-entry" @click="emit('open-chat', order)">
      <van-icon name="comment-o" />
      <span>与乘客聊天</span>
      <i v-if="hasUnread" class="od-unread-dot"></i>
    </button>

    <!-- 模块7：底部操作区（复制订单号 / 导航） -->
    <footer class="od-footer">
      <button type="button" class="od-action ghost" @click="copyOrderNo">
        <van-icon name="orders-o" />
        <span>复制订单号</span>
      </button>
      <button v-if="canNavigate" type="button" class="od-action primary" @click="onNavigate">
        <van-icon name="guide-o" />
        <span>{{ navButtonText }}</span>
      </button>
    </footer>

    <!-- 上报问题选择弹层 -->
    <van-action-sheet
      v-model:show="reportVisible"
      :actions="reportActions"
      cancel-text="取消"
      description="请选择上报的问题类型"
      @select="onReportSelect"
    />
  </van-popup>
</template>

<script setup>
import { computed, ref } from 'vue'
import { showToast } from 'vant'
import {
  formatPrice,
  formatDistance,
  formatDuration,
  formatTime,
  formatOrderStatus
} from '@/utils/driver-format'

const props = defineProps({
  visible: { type: Boolean, default: false },
  order: { type: Object, default: null },
  // 父级 DriverHome 注入：用于“导航去起点”（接驾阶段）与“联系乘客”
  navigateToPickup: { type: Function, default: null },
  contactPassenger: { type: Function, default: null },
  // 该订单是否有未读聊天消息（来自司机端 WS 推送，由 DriverHome 维护）
  hasUnread: { type: Boolean, default: false }
})
const emit = defineEmits(['update:visible', 'open-chat'])

function close() {
  emit('update:visible', false)
}

const statusLabel = computed(() => formatOrderStatus(props.order?.status))

// 顶栏按订单状态渐变配色（红仅用于取消/异常，符合网约车语义）
const statusClass = computed(() => {
  const s = Number(props.order?.status || 0)
  if (s === 1) return 'warn'
  if (s === 2) return 'info'
  if (s === 3) return 'ok'
  if (s === 5) return 'done'
  if (s === 6 || s === 7) return 'bad'
  return 'idle'
})

// 一口价等订单类型标签：后端未返回时自动隐藏
const orderTypeTag = computed(() => {
  const t = props.order?.orderType
  return t ? String(t) : ''
})

// 模块2：到手收入。后端若有 preDriverIncome（元）优先，否则退回预估价（分）
const incomeText = computed(() => {
  const o = props.order
  if (!o) return '--'
  if (o.preDriverIncome != null && o.preDriverIncome !== '') {
    return '¥' + Number(o.preDriverIncome).toFixed(2)
  }
  return formatPrice(o.estimatedPriceCents)
})

// 模块3：车型
function formatCarType(carType) {
  return { 1: '经济型', 2: '舒适型', 3: '商务型' }[Number(carType || 0)] || '经济型'
}
const carTypeLabel = computed(() => formatCarType(props.order?.carType))

// 模块4：字段映射（缺失字段优雅降级）
const distanceText = computed(() => formatDistance(props.order?.estimatedDistanceM))
const durationText = computed(() => formatDuration(props.order?.estimatedDurationS))
const createTimeText = computed(() => formatTime(props.order?.createdAt))
const pickupDistanceText = computed(() => String(props.order?.pickupDistance || '--'))
const remainMileageText = computed(() => String(props.order?.remainMileage || '--'))
const hasPickupDistance = computed(() => !!props.order?.pickupDistance)
const hasRemainMileage = computed(() => !!props.order?.remainMileage)

// 载客阶段（行程中）标记“已接到乘客”
const isPickedUp = computed(() => Number(props.order?.status || 0) === 3)

// 聊天入口仅在订单活跃期（已接单/行程中/待支付）开放，等待接单与已关闭阶段不可聊
const canChat = computed(() => {
  const s = Number(props.order?.status || 0)
  return s === 2 || s === 3 || s === 4
})

// 模块7：导航按钮——接驾去起点，载客去终点
const canNavigate = computed(() => {
  const s = Number(props.order?.status || 0)
  return s === 2 || s === 3
})
const navButtonText = computed(() => (Number(props.order?.status || 0) === 3 ? '导航去终点' : '导航去起点'))

function onNavigate() {
  const s = Number(props.order?.status || 0)
  if (s === 3) {
    navigateToDestination()
  } else if (typeof props.navigateToPickup === 'function') {
    props.navigateToPickup(props.order)
  } else {
    showToast('暂不支持导航')
  }
}

// 终点导航（载客阶段）：复用与首页一致的 amap URI 方案
function navigateToDestination() {
  const o = props.order
  const lng = Number(o?.toLongitude)
  const lat = Number(o?.toLatitude)
  if (!Number.isFinite(lng) || !Number.isFinite(lat) || (lng === 0 && lat === 0)) {
    showToast('暂无目的地坐标')
    return
  }
  const name = encodeURIComponent(o?.toAddress || '目的地')
  const url = `https://uri.amap.com/navigation?to=${lng},${lat},${name}&mode=car&policy=1&coordinate=gaode`
  window.open(url, '_blank')
}

// 模块5：联系乘客（复用父级 contactPassenger，含号码脱敏降级）
function onContact() {
  if (typeof props.contactPassenger === 'function') {
    props.contactPassenger(props.order)
  } else {
    showToast('暂无可联系号码')
  }
}

// 模块5：上报问题弹层
const reportVisible = ref(false)
const reportActions = [
  { name: '乘客未到' },
  { name: '找不到上车点' },
  { name: '乘客要求改目的地' },
  { name: '乘客取消订单' },
  { name: '其他异常' }
]
function onReportSelect(action) {
  reportVisible.value = false
  showToast('已上报：' + action.name)
}

// 取消 / 退款原因块
const showCancelBlock = computed(() => {
  const s = Number(props.order?.status || 0)
  return s === 6 || s === 7
})
const cancelActorLabel = computed(() => {
  const by = props.order?.cancelBy
  if (by === 'user') return '乘客取消'
  if (by === 'driver') return '司机取消'
  if (by === 'system') return '系统取消'
  return '取消原因'
})

async function copyOrderNo() {
  const no = String(props.order?.orderNo || '').trim()
  if (!no) {
    showToast('订单号为空')
    return
  }
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(no)
    } else {
      const ta = document.createElement('textarea')
      ta.value = no
      ta.setAttribute('readonly', '')
      ta.style.position = 'fixed'
      ta.style.opacity = '0'
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
    }
    showToast('订单号已复制')
  } catch (e) {
    showToast('复制失败，请手动选择')
  }
}
</script>

<style scoped>
.order-detail {
  display: flex;
  flex-direction: column;
  background: var(--driver-soft);
}

/* 模块1：顶栏渐变 */
.od-header {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  padding: 18px 16px 22px;
  color: var(--driver-on-primary);
  background: linear-gradient(135deg, var(--driver-primary) 0%, var(--driver-primary) 100%);
}
.od-header.warn { background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%); }
.od-header.info { background: linear-gradient(135deg, #3B82F6 0%, #2563EB 100%); }
.od-header.ok   { background: linear-gradient(135deg, #10B981 0%, #059669 100%); }
.od-header.done { background: linear-gradient(135deg, #6B7280 0%, #4B5563 100%); }
.od-header.bad  { background: linear-gradient(135deg, #EF4444 0%, #DC2626 100%); }
.od-header.idle { background: linear-gradient(135deg, var(--driver-primary) 0%, var(--driver-primary) 100%); }

.od-close {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: rgba(255, 255, 255, .18);
  color: var(--driver-on-primary);
  font-size: 18px;
}
.od-header-text { display: flex; flex-direction: column; gap: 6px; min-width: 0; }
.od-title-row { display: flex; align-items: center; gap: 10px; }
.od-status { font-size: 20px; font-weight: 800; }
.od-type-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, .55);
  background: rgba(255, 255, 255, .15);
  color: var(--driver-on-primary);
  font-size: 12px;
  font-weight: 700;
}
.od-orderno { font-size: 12px; opacity: .85; }

/* 内容区 */
.od-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px 14px 18px;
  display: grid;
  gap: 12px;
  align-content: start;
}

/* 模块2：收入卡片 */
.od-income-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 22px 14px 18px;
  border-radius: 18px;
  background: var(--driver-card);
  box-shadow: 0 8px 20px rgba(15, 23, 42, .05);
}
.od-income-label { color: var(--driver-muted); font-size: 12px; }
.od-income-value {
  color: var(--driver-accent);  /* 花小猪黄，非红色，避免“扣款”错觉 */
  font-size: 32px;
  line-height: 1.05;
  font-weight: 800;
  letter-spacing: -.5px;
}
.od-income-note { color: var(--driver-muted); font-size: 12px; text-align: center; }

/* 模块3：路线卡 */
.od-route {
  position: relative;
  display: grid;
  gap: 6px;
  padding: 16px 16px 16px 32px;
  border-radius: 16px;
  background: var(--driver-card);
  box-shadow: 0 6px 16px rgba(15, 23, 42, .04);
}
.od-route-row {
  display: grid;
  grid-template-columns: 14px minmax(0, 1fr);
  align-items: center;
  gap: 12px;
  min-height: 44px;
}
.od-route-meta { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.od-tag {
  display: inline-flex;
  align-items: center;
  align-self: flex-start;
  padding: 2px 8px;
  border-radius: 6px;
  background: var(--driver-soft);
  color: var(--driver-muted);
  font-size: 11px;
  font-weight: 700;
}
.od-route-meta strong {
  color: var(--driver-ink);
  font-size: 14px;
  line-height: 1.4;
  font-weight: 700;
  overflow-wrap: anywhere;
  word-break: break-all;
}
.od-remark {
  color: var(--driver-muted);
  font-size: 12px;
  overflow-wrap: anywhere;
  word-break: break-all;
}
.od-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}
.od-dot.from { background: #10B981; box-shadow: 0 0 0 3px rgba(16, 185, 129, .18); }
.od-dot.to   { background: #EF4444; box-shadow: 0 0 0 3px rgba(239, 68, 68, .18); }
.od-route-line {
  position: absolute;
  left: 21px;
  top: 38px;
  bottom: 38px;
  width: 2px;
  background: repeating-linear-gradient(to bottom, var(--driver-line) 0 4px, transparent 4px 8px);
}
.od-picked-flag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-top: 4px;
  padding: 4px 10px;
  border-radius: 999px;
  background: var(--driver-st-completed-bg);
  color: var(--driver-st-completed-fg);
  font-size: 12px;
  font-weight: 700;
  justify-self: start;
}

/* 模块4：信息网格 */
.od-metrics {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}
.od-metric {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 14px;
  border-radius: 12px;
  background: var(--driver-card);
  box-shadow: 0 4px 12px rgba(15, 23, 42, .04);
}
.od-metric span { color: var(--driver-muted); font-size: 12px; }
.od-metric strong {
  color: var(--driver-ink);
  font-size: 15px;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 模块5：操作按钮 */
.od-actions {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}
.od-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-height: 44px;
  padding: 0 14px;
  border: 1px solid var(--driver-line);
  border-radius: 22px;
  background: var(--driver-card);
  color: var(--driver-ink);
  font-size: 14px;
  font-weight: 700;
}
.od-action .van-icon { font-size: 16px; }
/* 模块7：底部主按钮沿用主题紫 */
.od-action.primary {
  border: 0;
  background: var(--driver-primary);
  color: var(--driver-on-primary);
  box-shadow: 0 6px 14px rgba(91, 92, 255, .24);
}
.od-action.ghost { background: var(--driver-card); }

/* 模块6：安全提示 */
.od-safety {
  margin: 0;
  padding: 10px 12px;
  border-radius: 10px;
  background: var(--driver-soft);
  color: var(--driver-faint);
  font-size: 12px;
  line-height: 1.5;
}

/* 取消 / 退款条 */
.od-cancel {
  display: grid;
  grid-template-columns: 24px minmax(0, 1fr);
  gap: 10px;
  align-items: center;
  padding: 12px 14px;
  border-radius: 12px;
  background: #FEF2F2;
  color: #B91C1C;
  border: 1px solid #FECACA;
}
.od-cancel .van-icon { font-size: 20px; }
.od-cancel strong { font-size: 14px; display: block; }
.od-cancel span { font-size: 12px; opacity: .9; }

/* 模块7：底部操作区 */
.od-footer {
  display: grid;
  grid-template-columns: 1fr 1.4fr;
  gap: 10px;
  padding: 12px 14px calc(12px + env(safe-area-inset-bottom));
  background: var(--driver-card);
  border-top: 1px solid var(--driver-line);
}

/* 聊天入口 */
.od-chat-entry {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin: 0 14px;
  height: 44px;
  border: 1px solid var(--driver-primary);
  border-radius: 22px;
  background: var(--driver-card);
  color: var(--driver-primary);
  font-size: 14px;
  font-weight: 700;
}
.od-unread-dot {
  position: absolute;
  top: 8px;
  right: calc(50% - 56px);
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #EF4444;
}
</style>
