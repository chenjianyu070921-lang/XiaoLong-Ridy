<script setup>
// 订单详情页面：按订单域结构化展示主订单、状态、派单、价格、支付、结算和轨迹信息。
// 后台取消、人工改派和退款均调用 admin 网关已开放接口，提交成功后重新拉取详情。
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
import { ordersApi } from '../../api/modules'
import { carTypeText, orderStatusTag, orderStatusText, orDash } from '../../utils/enums'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const canceling = ref(false)
const trackLoading = ref(false)
const redispatching = ref(false)
const refunding = ref(false)
const trackVisible = ref(false)
const trackPoints = ref([])
const detail = ref(null)

const order = computed(() => detail.value?.order || {})
const statusLogs = computed(() => detail.value?.status_logs || [])
const dispatchRecords = computed(() => detail.value?.dispatch_records || [])
const price = computed(() => detail.value?.price)
const payment = computed(() => detail.value?.payment)
const settlement = computed(() => detail.value?.settlement)
const canCancel = computed(() => ![5, 6].includes(order.value.status))

// 订单状态、支付状态和派单状态使用后端约定的数值枚举转换为运营可读文本。
const paymentStatusText = (status) => ({ 1: '待支付', 2: '已支付', 3: '支付异常', 4: '已退款' }[status] || orDash(status))
const dispatchStatusText = (status) => ({ 1: '待派单', 2: '已派单', 3: '已拒绝', 4: '已超时', 5: '已取消' }[status] || orDash(status))
const channelText = (channel) => ({ 1: '微信', 2: '支付宝', 3: '余额' }[channel] || orDash(channel))
const settlementStatusText = (status) => ({ 1: '待结算', 2: '已结算', 3: '结算失败' }[status] || orDash(status))
const dispatchTypeText = (type) => ({ 1: '自动派单', 2: '人工派单' }[type] || orDash(type))
const money = (value) => value === null || value === undefined || value === '' ? '-' : `¥${value}`
const distance = (value) => value === null || value === undefined || value === '' ? '-' : `${value} m`
const duration = (value) => value === null || value === undefined || value === '' ? '-' : `${value} 秒`
const requestID = (prefix) => crypto.randomUUID ? `${prefix}-${crypto.randomUUID()}` : `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`

// loadDetail 读取真实订单详情，主订单与关联数据由后端一次聚合返回。
const loadDetail = async () => {
  loading.value = true
  try {
    detail.value = await ordersApi.detail(route.params.id)
  } finally {
    loading.value = false
  }
}

// loadTrack 读取订单轨迹点。轨迹接口为只读能力，失败时不影响主订单详情展示。
const loadTrack = async () => {
  trackLoading.value = true
  try {
    const data = await ordersApi.track(order.value.id || route.params.id, { limit: 1000 })
    trackPoints.value = data?.points || []
    trackVisible.value = true
  } finally {
    trackLoading.value = false
  }
}

// cancelOrder 执行后台取消订单，成功后重新读取详情，避免使用本地乐观状态。
const cancelOrder = async () => {
  const result = await ElMessageBox.prompt('请输入取消原因', `取消订单「${order.value.order_no || order.value.id}」`, {
    confirmButtonText: '确认取消',
    cancelButtonText: '返回',
    inputValidator: (value) => value?.trim() ? true : '取消原因不能为空',
  })
  canceling.value = true
  try {
    await ordersApi.cancel(order.value.id, { reason: result.value.trim(), request_id: requestID('cancel') })
    ElMessage.success('订单已取消')
    await loadDetail()
  } finally {
    canceling.value = false
  }
}

// redispatchOrder 执行后台人工改派，目标司机 ID、原因和幂等号按接口契约提交给后端。
const redispatchOrder = async () => {
  const driverInput = await ElMessageBox.prompt('请输入新的司机 ID', `人工改派「${order.value.order_no || order.value.id}」`, {
    confirmButtonText: '下一步',
    cancelButtonText: '返回',
    inputValidator: (value) => /^[1-9]\d*$/.test(String(value || '').trim()) ? true : '司机 ID 必须是正整数',
  })
  const reasonInput = await ElMessageBox.prompt('请输入改派原因', '人工改派原因', {
    confirmButtonText: '确认改派',
    cancelButtonText: '返回',
    inputValidator: (value) => value?.trim() ? true : '改派原因不能为空',
  })
  redispatching.value = true
  try {
    await ordersApi.redispatch(order.value.id, {
      new_driver_id: Number(driverInput.value),
      reason: reasonInput.value.trim(),
      request_id: requestID('redispatch'),
    })
    ElMessage.success('订单已提交改派')
    await loadDetail()
  } finally {
    redispatching.value = false
  }
}

// refundOrder 执行后台退款。退款金额按元输入、转为分提交，避免前端浮点值直接进入接口。
const refundOrder = async () => {
  const amountInput = await ElMessageBox.prompt('请输入退款金额（元）', `订单退款「${order.value.order_no || order.value.id}」`, {
    confirmButtonText: '下一步',
    cancelButtonText: '返回',
    inputValidator: (value) => /^\d+(\.\d{1,2})?$/.test(String(value || '').trim()) && Number(value) > 0 ? true : '退款金额必须大于 0，最多两位小数',
  })
  const reasonInput = await ElMessageBox.prompt('请输入退款原因', '退款原因', {
    confirmButtonText: '确认退款',
    cancelButtonText: '返回',
    inputValidator: (value) => value?.trim() ? true : '退款原因不能为空',
  })
  refunding.value = true
  try {
    await ordersApi.refund(order.value.id, {
      refund_amount_cents: Math.round(Number(amountInput.value) * 100),
      reason: reasonInput.value.trim(),
      request_id: requestID('refund'),
    })
    ElMessage.success('退款请求已提交')
    await loadDetail()
  } finally {
    refunding.value = false
  }
}

// 订单 ID 变化时重新读取，兼容列表连续打开不同订单时组件被路由复用的情况。
watch(() => route.params.id, loadDetail, { immediate: true })
</script>

<template>
  <section class="order-detail" v-loading="loading">
    <div class="page-head">
      <div>
        <span class="eyebrow">订单管理 / 订单详情</span>
        <h1>订单 {{ orDash(order.order_no) }}</h1>
        <p>查看订单状态、派单、价格、支付和结算关联信息</p>
      </div>
      <div class="head-actions">
        <el-button :icon="Refresh" @click="loadDetail">刷新</el-button>
        <el-button :icon="ArrowLeft" @click="router.back()">返回列表</el-button>
      </div>
    </div>

    <div class="action-bar">
      <el-button type="warning" :loading="canceling" :disabled="!canCancel" @click="cancelOrder">取消订单</el-button>
      <el-button :loading="trackLoading" @click="loadTrack">轨迹查看</el-button>
      <el-button :loading="redispatching" :disabled="!canCancel" @click="redispatchOrder">人工改派</el-button>
      <el-button type="danger" :loading="refunding" @click="refundOrder">退款</el-button>
    </div>

    <section class="panel">
      <div class="section-title">订单主信息</div>
      <div class="info-grid">
        <div><span>订单号</span><strong>{{ orDash(order.order_no) }}</strong></div>
        <div><span>订单状态</span><el-tag :type="orderStatusTag(order.status)">{{ orderStatusText(order.status) }}</el-tag></div>
        <div><span>用户 ID</span><strong>{{ orDash(order.user_id) }}</strong></div>
        <div><span>司机 ID</span><strong>{{ orDash(order.driver_id) }}</strong></div>
        <div><span>车型</span><strong>{{ carTypeText(order.car_type) }}</strong></div>
        <div><span>下单时间</span><strong>{{ orDash(order.created_at) }}</strong></div>
        <div><span>出发地</span><strong>{{ orDash(order.from_address) }}</strong></div>
        <div><span>目的地</span><strong>{{ orDash(order.to_address) }}</strong></div>
        <div><span>预估距离</span><strong>{{ distance(order.estimated_distance_m) }}</strong></div>
        <div><span>预估时长</span><strong>{{ duration(order.estimated_duration_s) }}</strong></div>
        <div><span>预估金额</span><strong>{{ money(order.estimated_price) }}</strong></div>
        <div><span>更新时间</span><strong>{{ orDash(order.updated_at) }}</strong></div>
        <div v-if="order.cancel_reason"><span>取消原因</span><strong>{{ order.cancel_reason }}</strong></div>
        <div v-if="order.cancel_by"><span>取消方</span><strong>{{ order.cancel_by }}</strong></div>
      </div>
    </section>

    <section class="panel">
      <div class="section-title">订单状态流水</div>
      <el-table :data="statusLogs" empty-text="暂无状态流水">
        <el-table-column prop="from_status" label="原状态" width="110"><template #default="scope">{{ orderStatusText(scope.row.from_status) }}</template></el-table-column>
        <el-table-column prop="to_status" label="新状态" width="110"><template #default="scope">{{ orderStatusText(scope.row.to_status) }}</template></el-table-column>
        <el-table-column prop="operator_type" label="操作方" width="120" />
        <el-table-column prop="operator_id" label="操作人" width="100" />
        <el-table-column prop="remark" label="备注" min-width="220" show-overflow-tooltip />
        <el-table-column prop="created_at" label="时间" width="180" />
      </el-table>
    </section>

    <section class="panel">
      <div class="section-title">派单记录</div>
      <el-table :data="dispatchRecords" empty-text="暂无派单记录">
        <el-table-column prop="driver_id" label="司机 ID" width="110" />
        <el-table-column prop="dispatch_type" label="派单方式" width="130"><template #default="scope">{{ dispatchTypeText(scope.row.dispatch_type) }}</template></el-table-column>
        <el-table-column prop="status" label="派单状态" width="130"><template #default="scope">{{ dispatchStatusText(scope.row.status) }}</template></el-table-column>
        <el-table-column prop="match_score" label="匹配分" width="120" />
        <el-table-column prop="remark" label="备注" min-width="220" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column prop="updated_at" label="更新时间" width="180" />
      </el-table>
    </section>

    <section v-if="trackVisible" class="panel">
      <div class="section-title">轨迹点</div>
      <el-table :data="trackPoints" empty-text="暂无轨迹点">
        <el-table-column prop="recorded_at" label="记录时间" width="180" />
        <el-table-column prop="driver_id" label="司机 ID" width="110" />
        <el-table-column prop="longitude" label="经度" width="140" />
        <el-table-column prop="latitude" label="纬度" width="140" />
        <el-table-column prop="speed_kmh" label="速度 km/h" width="130" />
        <el-table-column prop="direction" label="方向" width="100" />
      </el-table>
    </section>

    <div class="two-columns">
      <section class="panel">
        <div class="section-title">价格明细</div>
        <div v-if="price" class="info-grid compact">
          <div><span>计价规则 ID</span><strong>{{ orDash(price.price_rule_id) }}</strong></div>
          <div><span>预估金额</span><strong>{{ money(price.estimated_price) }}</strong></div>
          <div><span>实际金额</span><strong>{{ money(price.actual_price) }}</strong></div>
          <div><span>应付金额</span><strong>{{ money(price.payable_amount) }}</strong></div>
          <div><span>基础费用</span><strong>{{ money(price.base_fee) }}</strong></div>
          <div><span>距离费用</span><strong>{{ money(price.distance_fee) }}</strong></div>
          <div><span>时长费用</span><strong>{{ money(price.time_fee) }}</strong></div>
          <div><span>夜间费用</span><strong>{{ money(price.night_fee) }}</strong></div>
          <div><span>动态费用</span><strong>{{ money(price.dynamic_fee) }}</strong></div>
          <div><span>优惠金额</span><strong>{{ money(price.discount_amount) }}</strong></div>
          <div><span>平台补贴</span><strong>{{ money(price.platform_subsidy) }}</strong></div>
        </div>
        <el-empty v-else description="暂无价格明细" :image-size="60" />
      </section>

      <section class="panel">
        <div class="section-title">支付明细</div>
        <div v-if="payment" class="info-grid compact">
          <div><span>支付单号</span><strong>{{ orDash(payment.payment_no) }}</strong></div>
          <div><span>支付状态</span><strong>{{ paymentStatusText(payment.status) }}</strong></div>
          <div><span>支付金额</span><strong>{{ money(payment.amount) }}</strong></div>
          <div><span>支付渠道</span><strong>{{ channelText(payment.channel) }}</strong></div>
          <div><span>第三方流水号</span><strong>{{ orDash(payment.transaction_id) }}</strong></div>
          <div><span>退款金额</span><strong>{{ money(payment.refund_amount) }}</strong></div>
          <div><span>支付时间</span><strong>{{ orDash(payment.paid_at) }}</strong></div>
        </div>
        <el-empty v-else description="暂无支付记录" :image-size="60" />
      </section>
    </div>

    <section class="panel">
      <div class="section-title">结算明细</div>
      <div v-if="settlement" class="info-grid compact">
        <div><span>结算单号</span><strong>{{ orDash(settlement.settlement_no) }}</strong></div>
        <div><span>结算状态</span><strong>{{ settlementStatusText(settlement.status) }}</strong></div>
        <div><span>司机 ID</span><strong>{{ orDash(settlement.driver_id) }}</strong></div>
        <div><span>结算总额</span><strong>{{ money(settlement.total_amount) }}</strong></div>
        <div><span>平台抽佣比例</span><strong>{{ orDash(settlement.platform_commission_rate) }}</strong></div>
        <div><span>平台抽佣</span><strong>{{ money(settlement.platform_commission) }}</strong></div>
        <div><span>司机收入</span><strong>{{ money(settlement.driver_income) }}</strong></div>
        <div><span>结算时间</span><strong>{{ orDash(settlement.settled_at) }}</strong></div>
      </div>
      <el-empty v-else description="暂无结算记录" :image-size="60" />
    </section>
  </section>
</template>

<style scoped>
.order-detail{color:#dce7f3}.page-head{display:flex;align-items:flex-end;justify-content:space-between;margin-bottom:18px}.eyebrow{color:#6f879d;font-size:12px;letter-spacing:.08em}.page-head h1{margin:7px 0 4px;font-size:28px;color:#f4f7fb}.page-head p{margin:0;color:#8293a5}.head-actions,.action-bar{display:flex;gap:10px}.action-bar{margin-bottom:16px}.panel{padding:18px;margin-bottom:16px;background:#101d2b;border:1px solid #1d3042;border-radius:10px;overflow:auto}.section-title{margin-bottom:15px;color:#fff;font-size:17px;font-weight:700}.info-grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:0}.info-grid>div{min-width:0;padding:14px 16px;border-top:1px solid #1d3042}.info-grid span{display:block;margin-bottom:7px;color:#8293a5;font-size:12px}.info-grid strong{display:block;overflow-wrap:anywhere;color:#eaf1f7;font-weight:500}.compact{grid-template-columns:repeat(2,minmax(0,1fr))}.two-columns{display:grid;grid-template-columns:1fr 1fr;gap:16px}.two-columns .panel{margin-bottom:0}@media(max-width:1100px){.info-grid{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:700px){.page-head{display:block}.head-actions{margin-top:14px}.two-columns{grid-template-columns:1fr}.info-grid,.compact{grid-template-columns:1fr}.action-bar{flex-wrap:wrap}}
</style>
