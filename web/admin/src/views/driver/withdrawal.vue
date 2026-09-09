<script setup>
// 司机提现审核页：列表 + 详情行内展示 + 打款成功/失败审核。
// 提现状态语义对齐 02_driver_module.sql：1申请中 2打款成功 3打款失败。
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Search, CircleCheck, Close } from '@element-plus/icons-vue'
import { driversApi } from '../../api/modules'
import { text, pageData } from '../../utils/format'
import BusinessFormDialog from '../../components/BusinessFormDialog.vue'

const route = useRoute()
const loading = ref(false)
const rows = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: undefined, driver_id: undefined })
const rowDetail = ref(null)
const dialog = ref(false)
const dialogType = ref('')
const dialogRecord = ref({})
const submitting = ref(false)

const withdrawStatusText = (v) => ({ 1: '申请中', 2: '打款成功', 3: '打款失败' }[v] || text(v))
const columns = [
  ['withdraw_no', '提现单号'], ['driver_id', '司机 ID'], ['amount', '金额'], ['payee_name', '收款人'],
  ['pay_account', '收款账户'], ['status', '状态'], ['remark', '备注'],
  ['applied_at', '申请时间'], ['paid_at', '打款时间'], ['created_at', '创建时间'],
]

const listParams = () => {
  const params = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) params.keyword = filters.keyword
  if (filters.status !== undefined && filters.status !== '') params.status = filters.status
  if (filters.driver_id) params.driver_id = Number(filters.driver_id)
  return params
}

const load = async () => {
  loading.value = true
  try {
    const data = await driversApi.withdrawals(listParams())
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
  filters.driver_id = undefined
  page.value = 1
  load()
}

const label = (key, value) => (key === 'status' ? withdrawStatusText(value) : text(value))
const openDetail = (row) => { rowDetail.value = row }
const rowDetailEntries = computed(() => Object.entries(rowDetail.value || {}).filter(([key, value]) => value !== null && value !== undefined && value !== '').map(([key, value]) => ({ key, value: typeof value === 'object' ? JSON.stringify(value) : label(key, value) })))

const action = (name, row = {}) => { dialogType.value = name; dialogRecord.value = row; dialog.value = true }
const dialogTitle = computed(() => ({ approveWithdraw: '提现打款成功', rejectWithdraw: '提现打款失败' }[dialogType.value] || '审核提现'))
const submitAction = async ({ payload, record }) => {
  submitting.value = true
  try {
    if (dialogType.value === 'approveWithdraw') await driversApi.approveWithdraw(record?.id, payload)
    else if (dialogType.value === 'rejectWithdraw') await driversApi.rejectWithdraw(record?.id, payload)
    dialog.value = false
    ElMessage.success('操作成功')
    await load()
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  // 支持从司机详情跳转时预填 driver_id 筛选条件。
  if (route.query.driver_id) filters.driver_id = Number(route.query.driver_id)
  await load()
})
watch(() => route.fullPath, async () => {
  filters.driver_id = route.query.driver_id ? Number(route.query.driver_id) : undefined
  page.value = 1
  await load()
})
</script>

<template>
  <section class="withdraw-page" v-loading="loading">
    <div class="page-head">
      <div><span class="eyebrow">用户与司机 / 司机提现</span><h1>司机提现审核</h1><p>审核申请中的提现申请并记录打款结果</p></div>
      <div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button></div>
    </div>
    <div class="panel filters">
      <el-input v-model="filters.keyword" clearable :prefix-icon="Search" placeholder="提现单号、收款人或账户" @keyup.enter="page=1;load" />
      <el-select v-model="filters.status" clearable placeholder="提现状态">
        <el-option :value="1" label="申请中" /><el-option :value="2" label="打款成功" /><el-option :value="3" label="打款失败" />
      </el-select>
      <el-input v-model="filters.driver_id" type="number" clearable placeholder="司机 ID" @keyup.enter="page=1;load" />
      <el-button type="primary" @click="page=1;load">查询</el-button>
      <el-button @click="reset">重置</el-button>
    </div>
    <div class="panel table-panel">
      <el-table :data="rows" stripe empty-text="暂无提现申请">
        <el-table-column v-for="c in columns" :key="c[0]" :prop="c[0]" :label="c[1]" min-width="150">
          <template #default="scope">
            <span :class="{ mono: c[0].includes('id') || c[0].includes('no'), amount: c[0] === 'amount' }">{{ label(c[0], scope.row[c[0]]) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="240">
          <template #default="scope">
            <el-button link type="primary" @click="openDetail(scope.row)">查看</el-button>
            <template v-if="scope.row.status === 1">
              <el-button link type="success" :icon="CircleCheck" @click="action('approveWithdraw', scope.row)">打款成功</el-button>
              <el-button link type="danger" :icon="Close" @click="action('rejectWithdraw', scope.row)">打款失败</el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>
      <div class="table-footer"><span>共 {{ total }} 条记录</span><el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="prev, pager, next, sizes" @current-change="load" @size-change="load" /></div>
    </div>
    <el-dialog :model-value="!!rowDetail" title="提现申请详情" width="720px" destroy-on-close @update:model-value="(v) => { if (!v) rowDetail = null }">
      <div class="panel detail-grid" style="border:none;box-shadow:none">
        <div v-for="item in rowDetailEntries" :key="item.key" class="detail-item"><span>{{ item.key }}</span><strong>{{ item.value }}</strong></div>
      </div>
    </el-dialog>
    <BusinessFormDialog v-model="dialog" :type="dialogType" :title="dialogTitle" :record="dialogRecord" :loading="submitting" @submit="submitAction" />
  </section>
</template>

<style scoped>
.withdraw-page{color:var(--text-color,#2e2c4e)}.page-head{display:flex;align-items:flex-end;justify-content:space-between;margin-bottom:22px}.eyebrow{color:var(--brand,#6c5ce7);font-size:12px;letter-spacing:.1em;font-weight:600}.page-head h1{margin:7px 0 4px;font-size:26px;color:var(--text-color,#2e2c4e)}.page-head p{margin:0;color:var(--muted-color,#8b88a3)}.actions,.filters{display:flex;gap:10px;align-items:center}.panel{background:var(--panel-bg,#fff);border:1px solid var(--border-color,#e5e4f0);border-radius:14px;box-shadow:var(--card-shadow,none)}.filters{padding:18px;margin-bottom:16px;flex-wrap:wrap}.filters .el-input{width:240px}.filters .el-select{width:160px}.table-panel{padding:0;overflow:hidden}.table-footer{display:flex;justify-content:space-between;align-items:center;padding:14px 18px;color:var(--muted-color,#8b88a3);border-top:1px solid var(--border-color,#e5e4f0)}.mono{font-family:ui-monospace;color:var(--brand,#6c5ce7)}.amount{color:#c0392b;font-weight:600}.detail-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:0;padding:0;overflow:hidden}.detail-item{padding:16px 20px;border-bottom:1px solid var(--border-color,#e5e4f0)}.detail-item span{display:block;margin-bottom:7px;color:var(--muted-color,#8b88a3);font-size:12px}.detail-item strong{display:block;overflow-wrap:anywhere;color:var(--text-color,#2e2c4e);font-weight:500}@media(max-width:1000px){.detail-grid{grid-template-columns:1fr}.page-head{display:block}.actions{margin-top:16px}.filters{flex-wrap:wrap}}
</style>
