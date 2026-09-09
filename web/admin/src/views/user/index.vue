<script setup>
// 用户管理独立页面：列表 + 详情 + 冻结/解冻 + 订单/优惠券历史。
// 状态枚举按用户语义独立维护，避免与订单等其他实体串用。
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { usersApi } from '../../api/modules'
import { text, pageData } from '../../utils/format'
import BusinessFormDialog from '../../components/BusinessFormDialog.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const rows = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: undefined, date_range: [] })
const detail = ref(null)
const dialog = ref(false)
const dialogType = ref('')
const dialogRecord = ref({})
const submitting = ref(false)

// 用户历史（详情页签）
const userHistoryTab = ref('orders')
const userHistoryRows = ref([])
const userHistoryTotal = ref(0)
const userHistoryPage = ref(1)
const userHistoryStatus = ref(undefined)
const userHistoryLoading = ref(false)

const isDetail = computed(() => !!route.params.id)
const userStatusText = (v) => ({ 1: '正常', 2: '冻结' }[v] || text(v))

const columns = [
  ['id', 'ID'], ['phone', '手机号'], ['nickname', '昵称'], ['real_name', '实名'],
  ['status', '状态'], ['created_at', '注册时间'],
]

const listParams = () => {
  const params = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) params.keyword = filters.keyword
  if (filters.status !== undefined && filters.status !== '') params.status = filters.status
  if (Array.isArray(filters.date_range) && filters.date_range.length === 2) {
    params.start_time = filters.date_range[0]
    params.end_time = filters.date_range[1]
  }
  return params
}

const load = async () => {
  loading.value = true
  try {
    const data = await usersApi.list(listParams())
    const p = pageData(data)
    rows.value = p.list
    total.value = p.total
  } finally {
    loading.value = false
  }
}

const reset = () => {
  filters.keyword = ''
  filters.status = undefined
  filters.date_range = []
  page.value = 1
  load()
}

const loadDetail = async () => {
  detail.value = await usersApi.detail(route.params.id)
}

// 详情字段渲染：status 用用户语义映射，其余原样展示。
const detailLabel = (key, value) => (key === 'status' ? userStatusText(value) : text(value))
const detailEntries = computed(() =>
  Object.entries(detail.value || {})
    .filter(([key, value]) => value !== null && value !== undefined && value !== '')
    .map(([key, value]) => ({ key, value: typeof value === 'object' ? JSON.stringify(value) : detailLabel(key, value) })),
)

// 用户历史：订单/优惠券两个页签，各自独立分页与状态枚举。
const userHistoryStatusOptions = computed(() => (userHistoryTab.value === 'orders'
  ? [[1, '待接单'], [2, '已接单'], [3, '行程中'], [4, '待支付'], [5, '已完成'], [6, '已取消']]
  : [[1, '未使用'], [2, '已使用'], [3, '已过期'], [4, '已锁定']]))
const historyStatusText = (status) => (userHistoryTab.value === 'orders'
  ? ({ 1: '待接单', 2: '已接单', 3: '行程中', 4: '待支付', 5: '已完成', 6: '已取消' }[status] || text(status))
  : ({ 1: '未使用', 2: '已使用', 3: '已过期', 4: '已锁定' }[status] || text(status)))
const formatUnixTime = (timestamp) => {
  const value = Number(timestamp)
  return Number.isFinite(value) && value > 0 ? new Date(value * 1000).toLocaleString('zh-CN', { hour12: false }) : '-'
}
const couponTypeText = (type) => ({ 1: '满减券', 2: '折扣券', 3: '立减券' }[type] || text(type))
const formatCents = (cents) => {
  const value = Number(cents)
  return Number.isFinite(value) ? `¥${(value / 100).toFixed(2)}` : '-'
}
const loadUserHistory = async () => {
  if (!route.params.id) return
  userHistoryLoading.value = true
  try {
    const params = { page: userHistoryPage.value, page_size: 20 }
    if (userHistoryStatus.value !== undefined) params.status = userHistoryStatus.value
    const data = userHistoryTab.value === 'orders'
      ? await usersApi.orders(route.params.id, params)
      : await usersApi.coupons(route.params.id, params)
    const result = pageData(data)
    userHistoryRows.value = result.list
    userHistoryTotal.value = result.total
  } finally {
    userHistoryLoading.value = false
  }
}

const openDetail = (row) => router.push(`/users/${row.id}`)
const action = (name, row = {}) => { dialogType.value = name; dialogRecord.value = row; dialog.value = true }

const dialogTitle = computed(() => ({ freeze: '冻结用户', unfreeze: '解冻用户' }[dialogType.value] || '确认操作'))
const submitAction = async ({ payload, record }) => {
  submitting.value = true
  try {
    const id = record?.id
    const handlers = {
      freeze: () => usersApi.freeze(id, payload),
      unfreeze: () => usersApi.unfreeze(id, payload),
    }
    await handlers[dialogType.value]?.()
    dialog.value = false
    ElMessage.success('操作成功')
    if (isDetail.value) await loadDetail()
    else await load()
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  if (isDetail.value) { await loadDetail(); await loadUserHistory(); return }
  await load()
})
watch(() => route.path, async () => {
  detail.value = null
  page.value = 1
  userHistoryPage.value = 1
  if (isDetail.value) { await loadDetail(); await loadUserHistory(); return }
  await load()
})
</script>

<template>
  <section class="user-page" v-loading="loading">
    <template v-if="isDetail">
      <div class="page-head">
        <div><span class="eyebrow">用户与司机 / 用户详情</span><h1>{{ detail?.nickname || detail?.phone || '用户详情' }}</h1><p>查看用户基础资料、订单与优惠券历史</p></div>
        <div class="actions"><el-button @click="router.back()">返回列表</el-button></div>
      </div>
      <div class="panel detail-grid">
        <div v-for="item in detailEntries" :key="item.key" class="detail-item">
          <span>{{ item.key }}</span><strong>{{ item.value }}</strong>
        </div>
      </div>
      <section class="panel history-panel">
        <div class="history-toolbar">
          <el-select v-model="userHistoryStatus" clearable placeholder="全部状态" @change="userHistoryPage=1;loadUserHistory()">
            <el-option v-for="option in userHistoryStatusOptions" :key="option[0]" :label="option[1]" :value="option[0]" />
          </el-select>
        </div>
        <el-tabs v-model="userHistoryTab" @tab-change="userHistoryStatus=undefined;userHistoryPage=1;loadUserHistory()">
          <el-tab-pane label="订单历史" name="orders">
            <el-table :data="userHistoryRows" v-loading="userHistoryLoading">
              <el-table-column prop="order_no" label="订单号" />
              <el-table-column prop="from_address" label="出发地" />
              <el-table-column prop="to_address" label="目的地" />
              <el-table-column prop="estimated_price" label="预估金额"><template #default="scope">¥{{ text(scope.row.estimated_price) }}</template></el-table-column>
              <el-table-column prop="status" label="状态"><template #default="scope">{{ historyStatusText(scope.row.status) }}</template></el-table-column>
              <el-table-column prop="created_at" label="下单时间" />
            </el-table>
          </el-tab-pane>
          <el-tab-pane label="优惠券历史" name="coupons">
            <el-table :data="userHistoryRows" v-loading="userHistoryLoading">
              <el-table-column prop="name" label="优惠券" />
              <el-table-column prop="type" label="类型"><template #default="scope">{{ couponTypeText(scope.row.type) }}</template></el-table-column>
              <el-table-column prop="face_value_cents" label="面值"><template #default="scope">{{ formatCents(scope.row.face_value_cents) }}</template></el-table-column>
              <el-table-column prop="status" label="状态"><template #default="scope">{{ historyStatusText(scope.row.status) }}</template></el-table-column>
              <el-table-column prop="received_at" label="领取时间"><template #default="scope">{{ formatUnixTime(scope.row.received_at) }}</template></el-table-column>
              <el-table-column prop="expire_at" label="到期时间"><template #default="scope">{{ formatUnixTime(scope.row.expire_at) }}</template></el-table-column>
            </el-table>
          </el-tab-pane>
        </el-tabs>
        <el-pagination v-model:current-page="userHistoryPage" :total="userHistoryTotal" layout="prev, pager, next" @current-change="loadUserHistory" />
      </section>
    </template>
    <template v-else>
      <div class="page-head">
        <div><span class="eyebrow">用户与司机 / 用户管理</span><h1>用户管理</h1><p>查询乘客用户、冻结/解冻违规账号</p></div>
        <div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button></div>
      </div>
      <div class="panel filters">
        <el-input v-model="filters.keyword" clearable :prefix-icon="Search" placeholder="手机号、昵称或实名" @keyup.enter="page=1;load" />
        <el-select v-model="filters.status" clearable placeholder="用户状态">
          <el-option :value="1" label="正常" /><el-option :value="2" label="冻结" />
        </el-select>
        <el-date-picker v-model="filters.date_range" type="datetimerange" value-format="YYYY-MM-DD HH:mm:ss" start-placeholder="注册开始" end-placeholder="注册结束" />
        <el-button type="primary" @click="page=1;load">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </div>
      <div class="panel table-panel">
        <el-table :data="rows" stripe empty-text="暂无用户">
          <el-table-column v-for="c in columns" :key="c[0]" :prop="c[0]" :label="c[1]" min-width="140">
            <template #default="scope">
              <span :class="{ mono: c[0].includes('id') || c[0].includes('no') }">{{ c[0] === 'status' ? userStatusText(scope.row.status) : text(scope.row[c[0]]) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" fixed="right" width="220">
            <template #default="scope">
              <el-button link type="primary" @click="openDetail(scope.row)">查看</el-button>
              <el-button link type="warning" @click="action(scope.row.status == 1 ? 'freeze' : 'unfreeze', scope.row)">{{ scope.row.status == 1 ? '冻结' : '解冻' }}</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="table-footer"><span>共 {{ total }} 条记录</span><el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="prev, pager, next, sizes" @current-change="load" @size-change="load" /></div>
      </div>
    </template>
    <BusinessFormDialog v-model="dialog" :type="dialogType" :title="dialogTitle" :record="dialogRecord" :loading="submitting" @submit="submitAction" />
  </section>
</template>

<style scoped>
.user-page{color:var(--text-color,#2e2c4e)}.page-head{display:flex;align-items:flex-end;justify-content:space-between;margin-bottom:22px}.eyebrow{color:var(--brand,#6c5ce7);font-size:12px;letter-spacing:.1em;font-weight:600}.page-head h1{margin:7px 0 4px;font-size:26px;color:var(--text-color,#2e2c4e)}.page-head p{margin:0;color:var(--muted-color,#8b88a3)}.actions,.filters,.history-toolbar{display:flex;gap:10px;align-items:center}.panel{background:var(--panel-bg,#fff);border:1px solid var(--border-color,#e5e4f0);border-radius:14px;box-shadow:var(--card-shadow,none)}.filters{padding:18px;margin-bottom:16px;flex-wrap:wrap}.filters .el-input{width:240px}.filters .el-select,.history-toolbar .el-select{width:160px}.table-panel{padding:0;overflow:hidden}.table-footer{display:flex;justify-content:space-between;align-items:center;padding:14px 18px;color:var(--muted-color,#8b88a3);border-top:1px solid var(--border-color,#e5e4f0)}.mono{font-family:ui-monospace;color:var(--brand,#6c5ce7)}.detail-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:0;padding:0;overflow:hidden}.detail-item{padding:16px 20px;border-bottom:1px solid var(--border-color,#e5e4f0)}.detail-item span{display:block;margin-bottom:7px;color:var(--muted-color,#8b88a3);font-size:12px}.detail-item strong{display:block;overflow-wrap:anywhere;color:var(--text-color,#2e2c4e);font-weight:500}.history-panel{padding:20px;margin-top:16px}.history-toolbar{justify-content:flex-end;margin-bottom:8px}.history-panel :deep(.el-pagination){margin-top:14px;justify-content:flex-end}@media(max-width:1000px){.detail-grid{grid-template-columns:1fr}.page-head{display:block}.actions{margin-top:16px}.filters{flex-wrap:wrap}}
</style>
