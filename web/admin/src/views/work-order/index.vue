<script setup>
// 投诉与申诉工单独立页面：列表 + 详情 + 新建/流转/证据 + 批量处理。
// 工单状态/优先级/来源按工单语义独立维护，避免与其他实体枚举串用。
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Plus, Search } from '@element-plus/icons-vue'
import { workOrdersApi } from '../../api/modules'
import { text, pageData } from '../../utils/format'
import { useUserStore } from '../../store/user'
import BusinessFormDialog from '../../components/BusinessFormDialog.vue'
import WorkOrderEvidenceList from '../../components/WorkOrderEvidenceList.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const rows = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ status: undefined, assignee_id: '', work_order_type: undefined })
const detail = ref(null)
const dialog = ref(false)
const dialogType = ref('')
const dialogRecord = ref({})
const submitting = ref(false)
const evidenceRefreshKey = ref(0)
const selectionRows = ref([])
const error = ref('')

const isDetail = computed(() => !!route.params.id)
const workOrderStatusText = (v) => ({ 1: '待处理', 2: '处理中', 3: '已仲裁', 4: '已关闭' }[v] || text(v))
const priorityText = (v) => ({ 1: '低', 2: '中', 3: '高' }[v] || text(v))
const sourceTypeText = (v) => ({ order: '订单', user: '用户', driver: '司机' }[v] || text(v))
// allowedActions 同时按服务端状态机和当前管理员角色收敛操作入口，减少必然失败的请求。
const allowedActions = (status) => {
  const byStatus = { 1: ['assign'], 2: ['follow', 'arbitrate'], 3: ['close'], 4: ['reopen'], 5: ['reopen'] }[status] || []
  if (userStore.admin?.role === 1) return byStatus
  if (userStore.admin?.role === 2) return byStatus.filter((actionName) => ['assign', 'follow'].includes(actionName))
  return byStatus.filter((actionName) => actionName === 'follow')
}
const columns = [
  ['id', '工单号'], ['title', '标题'], ['source_type', '来源'], ['priority', '优先级'],
  ['status', '状态'], ['created_at', '创建时间'],
]

const listParams = () => {
  const params = { page: page.value, page_size: pageSize.value }
  if (filters.status !== undefined && filters.status !== '') params.status = filters.status
  if (filters.work_order_type !== undefined && filters.work_order_type !== '') params.work_order_type = filters.work_order_type
  if (filters.assignee_id !== '') {
    const numberValue = Number(filters.assignee_id)
    if (Number.isInteger(numberValue) && numberValue > 0) params.assignee_id = numberValue
  }
  return params
}

const load = async () => {
  loading.value = true
  error.value = ''
  try {
    const data = await workOrdersApi.list(listParams())
    const p = pageData(data)
    rows.value = p.list
    total.value = p.total
  } catch (err) {
    error.value = err?.message || '工单列表加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

const reset = () => {
  filters.status = undefined
  filters.assignee_id = ''
  filters.work_order_type = undefined
  page.value = 1
  load()
}

const loadDetail = async () => {
  error.value = ''
  try {
    detail.value = await workOrdersApi.detail(route.params.id)
  } catch (err) {
    error.value = err?.message || '工单详情加载失败，请稍后重试'
  }
}

const detailLabel = (key, value) => {
  if (key === 'status') return workOrderStatusText(value)
  if (key === 'priority') return priorityText(value)
  if (key === 'source_type') return sourceTypeText(value)
  return text(value)
}
const detailEntries = computed(() =>
  Object.entries(detail.value || {})
    .filter(([key, value]) => value !== null && value !== undefined && value !== '')
    .map(([key, value]) => ({ key, value: typeof value === 'object' ? JSON.stringify(value) : detailLabel(key, value) })),
)

const openDetail = (row) => router.push(`/work-orders/${row.id}`)
const action = (name, row = {}) => {
  const record = { ...row, allowed_actions: allowedActions(row.status) }
  if (name === 'workOrderAction' && !record.allowed_actions.length) {
    ElMessage.warning('当前工单状态或管理员角色没有可执行动作')
    return
  }
  dialogType.value = name
  dialogRecord.value = record
  dialog.value = true
}
const selectedIDs = computed(() => selectionRows.value.map((row) => row.id).filter(Boolean))
const batchAction = (name) => {
  if (!selectedIDs.value.length) return ElMessage.warning('请先勾选需要处理的记录')
  action(name, { ids: selectedIDs.value })
}
const openCreate = () => action('createWorkOrder')

const dialogTitle = computed(() => ({
  createWorkOrder: '新建工单',
  workOrderAction: '工单流转',
  batchWorkOrderAction: '批量处理工单',
  workOrderEvidence: '添加工单证据',
}[dialogType.value] || '确认操作'))

const submitAction = async ({ payload, record }) => {
  submitting.value = true
  try {
    const id = record?.id
    const handlers = {
      createWorkOrder: () => workOrdersApi.create(payload),
      workOrderAction: () => workOrdersApi.action(id, payload),
      batchWorkOrderAction: () => workOrdersApi.batchAction(payload),
      workOrderEvidence: () => workOrdersApi.evidence(id, payload),
    }
    await handlers[dialogType.value]?.()
    if (dialogType.value === 'workOrderEvidence') evidenceRefreshKey.value += 1
    dialog.value = false
    selectionRows.value = []
    ElMessage.success('操作成功')
    if (isDetail.value) await loadDetail()
    else await load()
  } catch (err) {
    error.value = err?.message || '工单操作失败，请刷新后重试'
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  if (isDetail.value) { await loadDetail(); return }
  await load()
})
watch(() => route.path, async () => {
  detail.value = null
  selectionRows.value = []
  page.value = 1
  if (isDetail.value) { await loadDetail(); return }
  await load()
})
</script>

<template>
  <section class="work-order-page" v-loading="loading">
    <template v-if="isDetail">
      <div class="page-head">
        <div><span class="eyebrow">工单与风控 / 工单详情</span><h1>{{ detail?.title || '工单详情' }}</h1><p>查看工单信息、跟进流转并添加证据</p></div>
        <div class="actions">
          <el-button type="primary" :disabled="!allowedActions(detail?.status).length" @click="action('workOrderAction', detail || {})">工单流转</el-button>
          <el-button @click="action('workOrderEvidence', detail || {})">添加证据</el-button>
          <el-button @click="router.back()">返回列表</el-button>
        </div>
      </div>
      <el-alert v-if="error" class="page-error" :title="error" type="error" show-icon :closable="false"><template #default><el-button link type="danger" @click="isDetail ? loadDetail() : load()">重新加载</el-button></template></el-alert>
      <div class="panel detail-grid">
        <div v-for="item in detailEntries" :key="item.key" class="detail-item">
          <span>{{ item.key }}</span><strong>{{ item.value }}</strong>
        </div>
      </div>
      <WorkOrderEvidenceList :work-order-id="route.params.id" :refresh-key="evidenceRefreshKey" />
    </template>
    <template v-else>
      <div class="page-head">
        <div><span class="eyebrow">工单与风控 / 投诉与申诉</span><h1>投诉与申诉工单</h1><p>处理用户与司机的投诉申诉工单</p></div>
        <div class="actions">
          <el-button type="primary" :disabled="!selectedIDs.length" @click="batchAction('batchWorkOrderAction')">批量处理</el-button>
          <el-button :icon="Refresh" @click="load">刷新</el-button>
          <el-button type="primary" :icon="Plus" @click="openCreate">新建工单</el-button>
        </div>
      </div>
      <div class="panel filters">
        <el-select v-model="filters.status" clearable placeholder="工单状态">
          <el-option :value="1" label="待处理" /><el-option :value="2" label="处理中" /><el-option :value="3" label="已仲裁" /><el-option :value="4" label="已关闭" />
        </el-select>
        <el-select v-model="filters.work_order_type" clearable placeholder="工单类型">
          <el-option :value="1" label="投诉" /><el-option :value="2" label="申诉" /><el-option :value="3" label="其他" />
        </el-select>
        <el-input v-model="filters.assignee_id" type="number" placeholder="处理人 ID" clearable @keyup.enter="page=1;load" />
        <el-button type="primary" @click="page=1;load">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </div>
      <div class="panel table-panel">
        <el-table :data="rows" stripe empty-text="暂无工单" @selection-change="selectionRows=$event">
          <el-table-column type="selection" width="46" />
          <el-table-column v-for="c in columns" :key="c[0]" :prop="c[0]" :label="c[1]" min-width="140">
            <template #default="scope">
              <span :class="{ mono: c[0].includes('id') || c[0].includes('no') }">{{ detailLabel(c[0], scope.row[c[0]]) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" fixed="right" width="160">
            <template #default="scope">
              <el-button link type="primary" @click="openDetail(scope.row)">查看</el-button>
              <el-button link type="primary" :disabled="!allowedActions(scope.row.status).length" @click="action('workOrderAction', scope.row)">流转</el-button>
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
.work-order-page{color:var(--text-color,#2e2c4e)}.page-head{display:flex;align-items:flex-end;justify-content:space-between;margin-bottom:22px}.eyebrow{color:var(--brand,#6c5ce7);font-size:12px;letter-spacing:.1em;font-weight:600}.page-head h1{margin:7px 0 4px;font-size:26px;color:var(--text-color,#2e2c4e)}.page-head p{margin:0;color:var(--muted-color,#8b88a3)}.actions,.filters{display:flex;gap:10px;align-items:center}.page-error{margin-bottom:16px}.panel{background:var(--panel-bg,#fff);border:1px solid var(--border-color,#e5e4f0);border-radius:14px;box-shadow:var(--card-shadow,none)}.filters{padding:18px;margin-bottom:16px;flex-wrap:wrap}.filters .el-input{width:180px}.filters .el-select{width:160px}.table-panel{padding:0;overflow:hidden}.table-footer{display:flex;justify-content:space-between;align-items:center;padding:14px 18px;color:var(--muted-color,#8b88a3);border-top:1px solid var(--border-color,#e5e4f0)}.mono{font-family:ui-monospace;color:var(--brand,#6c5ce7)}.detail-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:0;padding:0;overflow:hidden;margin-bottom:16px}.detail-item{padding:16px 20px;border-bottom:1px solid var(--border-color,#e5e4f0)}.detail-item span{display:block;margin-bottom:7px;color:var(--muted-color,#8b88a3);font-size:12px}.detail-item strong{display:block;overflow-wrap:anywhere;color:var(--text-color,#2e2c4e);font-weight:500}@media(max-width:1000px){.detail-grid{grid-template-columns:1fr}.page-head{display:block}.actions{margin-top:16px}.filters{flex-wrap:wrap}}
</style>
