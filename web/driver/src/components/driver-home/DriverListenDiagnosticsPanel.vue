<template>
  <van-popup
    :show="visible"
    position="bottom"
    teleport="#driver-home-popups"
    class="listen-diagnostics"
    :style="{ left: '0', right: '0', width: 'min(100vw, 390px)', height: '100%', margin: '0 auto', borderRadius: 0 }"
    @update:show="(v) => emit('update:visible', v)"
  >
    <div class="diag-header">
      <button type="button" class="back" @click="close">返回</button>
      <span class="diag-title">听单检测</span>
    </div>

    <div class="diag-body">
      <!-- 总状态条：结论先行 -->
      <div class="status-bar" :class="overall.statusClass">
        <van-icon :name="overall.icon" class="status-icon" />
        <div class="status-text">
          <strong>{{ overall.title }}</strong>
          <small v-if="overall.subtitle">{{ overall.subtitle }}</small>
        </div>
      </div>

      <!-- 两个司机能懂的指标卡 -->
      <div class="metric-cards">
        <div class="metric-card">
          <van-icon name="signal" class="mc-icon" />
          <div class="mc-body">
            <span class="m-label">网络</span>
            <strong class="m-value" :class="netClass">{{ netLabel }}</strong>
          </div>
        </div>
        <div class="metric-card">
          <van-icon name="clock-o" class="mc-icon" />
          <div class="mc-body">
            <span class="m-label">接单速度</span>
            <strong class="m-value" :class="listenClass">{{ listenLabel }}</strong>
          </div>
        </div>
      </div>

      <!-- 人话建议 -->
      <p v-if="overall.tip" class="tip">{{ overall.tip }}</p>

      <!-- 一键检测 -->
      <button type="button" class="run-btn" :disabled="speedTesting" @click="runSpeedTest">
        {{ speedTesting ? `检测中 ${speedProgress}/${SPEED_ROUNDS}` : (speedResult ? '重新测一下' : '测一下网络') }}
      </button>

      <!-- 技术明细（默认收起） -->
      <div class="advanced">
        <button type="button" class="adv-toggle" @click="showAdvanced = !showAdvanced">
          <span>{{ showAdvanced ? '收起技术明细' : '查看技术明细' }}</span>
          <van-icon :name="showAdvanced ? 'arrow-up' : 'arrow-down'" />
        </button>
        <template v-if="showAdvanced">
          <div class="diag-card">
            <div class="card-title">推送链路状态</div>
            <div class="diag-row">
              <span>实时推送通道</span>
              <b :class="wsConnected ? 'ok' : 'bad'">{{ wsConnected ? '已连接' : '未连接（轮询兜底中）' }}</b>
            </div>
            <div class="diag-row"><span>轮询兜底间隔</span><b>10 秒</b></div>
            <div class="diag-row"><span>听单延迟样本</span><b>{{ samples.length }} 条</b></div>
          </div>

          <div class="diag-card">
            <div class="card-title">后台网速（HTTP 往返）</div>
            <template v-if="speedResult">
              <div class="metric-grid">
                <div class="metric"><span>平均</span><b>{{ speedResult.avg }} ms</b></div>
                <div class="metric"><span>最快</span><b>{{ speedResult.min }} ms</b></div>
                <div class="metric"><span>最慢</span><b>{{ speedResult.max }} ms</b></div>
                <div class="metric"><span>抖动</span><b>±{{ speedResult.jitter }} ms</b></div>
              </div>
              <p class="grade" :class="speedResult.gradeClass">网络评级：{{ speedResult.grade }}</p>
            </template>
            <p v-if="speedError" class="diag-error">{{ speedError }}</p>
          </div>

          <div class="diag-card">
            <div class="card-title">听单速度（后台派单 → 司机端）</div>
            <template v-if="stats.count > 0">
              <div class="metric-grid">
                <div class="metric"><span>平均链路延迟</span><b>{{ stats.avgLink }} ms</b></div>
                <div class="metric"><span>平均全程延迟</span><b>{{ stats.avgE2E }} ms</b></div>
              </div>
              <p class="grade" :class="stats.gradeClass">听单评级：{{ stats.grade }}</p>
              <ul class="sample-list">
                <li v-for="(s, i) in samples" :key="s.receivedAt + '-' + i" class="sample-row">
                  <span>{{ formatClock(s.receivedAt) }}</span>
                  <span>订单 {{ s.orderId || '--' }}</span>
                  <span>链路 {{ s.linkDelay ?? '--' }} ms</span>
                  <span>全程 {{ s.e2eDelay ?? '--' }} ms</span>
                </li>
              </ul>
            </template>
            <p v-else class="diag-hint">暂无派单推送样本。保持在线，收到新派单后自动记录延迟。</p>
          </div>
        </template>
      </div>
    </div>
  </van-popup>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { listDriverDispatches } from '@/api/driver'

const props = defineProps({
  visible: { type: Boolean, default: false },
  wsConnected: { type: Boolean, default: false },
  // 派单推送延迟样本：{ orderId, receivedAt, linkDelay, e2eDelay }，由 DriverHome 在收到推送时记录
  samples: { type: Array, default: () => [] }
})
const emit = defineEmits(['update:visible'])

const SPEED_ROUNDS = 5
const speedTesting = ref(false)
const speedProgress = ref(0)
const speedResult = ref(null)
const speedError = ref('')
const showAdvanced = ref(false)

function close() {
  emit('update:visible', false)
}

function formatClock(ts) {
  if (!ts) return '--'
  const d = new Date(ts)
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}`
}

// gradeHttp 按 HTTP 往返均值评级（移动网络经验阈值）。
function gradeHttp(avg) {
  if (avg < 100) return { grade: '极佳', gradeClass: 'ok' }
  if (avg < 250) return { grade: '良好', gradeClass: 'ok' }
  if (avg < 600) return { grade: '一般', gradeClass: 'warn' }
  return { grade: '较差', gradeClass: 'bad' }
}

// gradeDispatch 按推送链路延迟均值评级（WS 推送通常远小于轮询周期）。
function gradeDispatch(avgLink) {
  if (avgLink == null) return { grade: '暂无数据', gradeClass: 'warn' }
  if (avgLink < 500) return { grade: '实时', gradeClass: 'ok' }
  if (avgLink < 3000) return { grade: '可用', gradeClass: 'warn' }
  return { grade: '延迟明显，建议检查网络', gradeClass: 'bad' }
}

// runSpeedTest 连续请求后台派单记录接口测 HTTP 往返延迟。
async function runSpeedTest() {
  speedTesting.value = true
  speedError.value = ''
  speedResult.value = null
  const rttList = []
  try {
    for (let i = 0; i < SPEED_ROUNDS; i += 1) {
      speedProgress.value = i + 1
      const startedAt = performance.now()
      await listDriverDispatches({ page: 1, pageSize: 1 }, { silentError: true })
      rttList.push(Math.round(performance.now() - startedAt))
    }
  } catch (e) {
    speedError.value = e?.response?.data?.message || '检测失败，请确认网络后重试'
    speedTesting.value = false
    return
  }
  const avg = Math.round(rttList.reduce((a, b) => a + b, 0) / rttList.length)
  const min = Math.min(...rttList)
  const max = Math.max(...rttList)
  const jitter = Math.round(max - min)
  speedResult.value = { avg, min, max, jitter, ...gradeHttp(avg) }
  speedTesting.value = false
}

const stats = computed(() => {
  const linkSamples = props.samples.map((s) => s.linkDelay).filter((v) => Number.isFinite(v))
  const e2eSamples = props.samples.map((s) => s.e2eDelay).filter((v) => Number.isFinite(v))
  const avg = (list) => (list.length ? Math.round(list.reduce((a, b) => a + b, 0) / list.length) : null)
  const avgLink = avg(linkSamples)
  return {
    count: props.samples.length,
    avgLink,
    avgE2E: avg(e2eSamples),
    ...gradeDispatch(avgLink)
  }
})

// ── 方案 A：结论先行 ──
function severityOf(cls) {
  if (cls === 'ok') return 0
  if (cls === 'warn') return 1
  if (cls === 'bad') return 2
  return null
}

const netSeverity = computed(() => (speedResult.value ? severityOf(speedResult.value.gradeClass) : null))
const listenSeverity = computed(() => (stats.value.count > 0 ? severityOf(stats.value.gradeClass) : null))

// 综合网络与接单速度，给司机一个总判断（绿/橙/红）。
const overall = computed(() => {
  const severities = [netSeverity.value, listenSeverity.value].filter((v) => v !== null)
  if (severities.length === 0) {
    return { statusClass: 'idle', icon: 'question-o', title: '尚未检测', subtitle: '点下方按钮测一下网络', tip: '' }
  }
  const worst = Math.max(...severities)
  if (worst === 2) {
    return { statusClass: 'bad', icon: 'warning', title: '暂时听不到单', subtitle: '网络较差，建议检查后重试', tip: tipFor(worst) }
  }
  if (worst === 1) {
    return { statusClass: 'warn', icon: 'info', title: '听单较慢，注意网络', subtitle: '高峰期可能慢几秒', tip: tipFor(worst) }
  }
  return { statusClass: 'ok', icon: 'success', title: '可正常听单', subtitle: '网络良好', tip: '' }
})

function tipFor(worst) {
  if (netSeverity.value === worst && netSeverity.value != null) {
    return '建议切换到 4G/5G，或到信号更好的地方再试。'
  }
  if (listenSeverity.value === worst && listenSeverity.value != null) {
    return '收到订单较慢，建议切换网络或重启 App 后重试。'
  }
  return '建议检查网络后重试。'
}

// 司机能懂的标签：把 ms / 评级文字翻译成「快/一般/差」「实时/略慢/明显延迟」。
const netLabel = computed(() => {
  if (!speedResult.value) return '未测'
  const g = speedResult.value.grade
  if (g === '极佳' || g === '良好') return '快'
  if (g === '一般') return '一般'
  return '差'
})
const netClass = computed(() => (speedResult.value ? speedResult.value.gradeClass : 'idle'))

const listenLabel = computed(() => {
  if (stats.value.count === 0) return '暂无派单'
  const g = stats.value.grade
  if (g === '实时') return '实时'
  if (g === '可用') return '略慢'
  return '明显延迟'
})
const listenClass = computed(() => (stats.value.count > 0 ? stats.value.gradeClass : 'idle'))

// 进入面板且无网速数据时自动测一次，减少手动操作。
watch(
  () => props.visible,
  (v) => {
    if (v && !speedResult.value && !speedTesting.value) {
      runSpeedTest()
    }
  }
)
</script>

<style scoped>
.listen-diagnostics {
  display: flex;
  flex-direction: column;
  background: var(--driver-soft);
}
.diag-header {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--driver-card);
  border-bottom: 1px solid var(--driver-line);
  padding: 12px;
}
.diag-header .back {
  border: none;
  background: none;
  color: var(--driver-primary);
  font-size: 14px;
  padding: 4px 8px;
}
.diag-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--driver-ink);
}
.diag-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: grid;
  gap: 12px;
  align-content: start;
}

/* 状态条 */
.status-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  border-radius: 12px;
  padding: 14px 16px;
}
.status-bar.ok { background: #ECFDF5; color: #059669; }
.status-bar.warn { background: #FFFBEB; color: #D97706; }
.status-bar.bad { background: #FEF2F2; color: #DC2626; }
.status-bar.idle { background: #F3F4F6; color: #6B7280; }
.status-icon { font-size: 26px; }
.status-text { display: flex; flex-direction: column; }
.status-text strong { font-size: 16px; font-weight: 700; }
.status-text small { font-size: 12px; opacity: .85; margin-top: 2px; }

/* 指标卡 */
.metric-cards {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}
.metric-card {
  background: var(--driver-card);
  border: 1px solid var(--driver-line);
  border-radius: 12px;
  padding: 14px;
  display: flex;
  align-items: center;
  gap: 10px;
}
.mc-icon { font-size: 22px; color: var(--driver-primary); }
.mc-body { display: flex; flex-direction: column; }
.m-label { font-size: 12px; color: var(--driver-muted); }
.m-value { font-size: 18px; font-weight: 700; color: var(--driver-ink); }
.m-value.ok { color: #059669; }
.m-value.warn { color: #D97706; }
.m-value.bad { color: #DC2626; }
.m-value.idle { color: var(--driver-muted); }

/* 建议 */
.tip {
  margin: 0;
  background: #FFF7ED;
  border: 1px solid #FED7AA;
  color: #C2410C;
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 13px;
  line-height: 1.6;
}

/* 检测按钮 */
.run-btn {
  min-height: 44px;
  border: none;
  border-radius: 22px;
  background: var(--driver-primary);
  color: #fff;
  font-size: 15px;
  font-weight: 700;
}
.run-btn:disabled { opacity: .6; }

/* 高级明细 */
.advanced { display: grid; gap: 12px; }
.adv-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  border: 1px dashed var(--driver-line);
  background: none;
  border-radius: 10px;
  padding: 10px;
  color: var(--driver-muted);
  font-size: 13px;
}
.diag-card {
  background: var(--driver-card);
  border: 1px solid var(--driver-line);
  border-radius: 12px;
  padding: 14px;
  display: grid;
  gap: 10px;
}
.card-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--driver-ink);
}
.diag-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--driver-muted);
  font-size: 13px;
}
.diag-row b.ok { color: #059669; }
.diag-row b.bad { color: #DC2626; }
.metric-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}
.metric {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  background: var(--driver-bg);
  border-radius: 8px;
  padding: 8px 10px;
  color: var(--driver-muted);
  font-size: 12px;
}
.metric b { color: var(--driver-ink); font-size: 14px; }
.grade { margin: 0; font-size: 13px; font-weight: 700; }
.grade.ok { color: #059669; }
.grade.warn { color: #D97706; }
.grade.bad { color: #DC2626; }
.diag-hint {
  margin: 0;
  color: var(--driver-muted);
  font-size: 12px;
  line-height: 1.6;
}
.diag-error { margin: 0; color: #DC2626; font-size: 13px; }
.sample-list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 6px;
}
.sample-row {
  display: grid;
  grid-template-columns: repeat(4, auto);
  justify-content: space-between;
  gap: 6px;
  color: var(--driver-muted);
  font-size: 12px;
  padding: 6px 8px;
  background: var(--driver-bg);
  border-radius: 8px;
}
</style>
