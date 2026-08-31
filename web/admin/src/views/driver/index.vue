<script setup>
// 司机管理独立页面：列表 + 详情 + 冻结。
// 司机服务状态语义：1=待审核 2=服务中 3=已冻结，独立维护避免与其他实体状态串用。
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { driversApi } from '../../api/modules'
import { text, pageData } from '../../utils/format'
import BusinessFormDialog from '../../components/BusinessFormDialog.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const rows = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: undefined })
const detail = ref(null)
const dialog = ref(false)
const dialogType = ref('')
const dialogRecord = ref({})
const submitting = ref(false)

const isDetail = computed(() => !!route.params.id)
const driverStatusText = (v) => ({ 1: '待审核', 2: '服务中', 3: '已冻结' }[v] || text(v))

const columns = [
  ['id', '司机 ID'], ['phone', '手机号'], ['real_name', '姓名'], ['status', '状态'],
  ['online_status', '在线状态'], ['created_at', '注册时间'],
]

const listParams = () => {
  const params = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) params.keyword = filters.keyword
  if (filters.status !== undefined && filters.status !== '') params.status = filters.status
  return params
}

const load = async () => {
  loading.value = true
  try {
    const data = await driversApi.list(listParams())
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
  page.value = 1
  load()
}

const loadDetail = async () => {
  detail.value = await driversApi.detail(route.params.id)
}

const detailLabel = (key, value) => (key === 'status' ? driverStatusText(value) : text(value))
const detailEntries = computed(() =>
  Object.entries(detail.value || {})
    .filter(([key, value]) => value !== null && value !== undefined && value !== '')
    .map(([key, value]) => ({ key, value: typeof value === 'object' ? JSON.stringify(value) : detailLabel(key, value) })),
)

const openDetail = (row) => router.push(`/drivers/${row.id}`)
const action = (name, row = {}) => { dialogType.value = name; dialogRecord.value = row; dialog.value = true }

const dialogTitle = computed(() => ({ freezeDriver: '冻结司机' }[dialogType.value] || '确认操作'))
const submitAction = async ({ payload, record }) => {
  submitting.value = true
  try {
    if (dialogType.value === 'freezeDriver') await driversApi.freeze(record?.id, payload)
    dialog.value = false
    ElMessage.success('操作成功')
    if (isDetail.value) await loadDetail()
    else await load()
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
  page.value = 1
  if (isDetail.value) { await loadDetail(); return }
  await load()
})
</script>

<template>
  <section class="driver-page" v-loading="loading">
    <template v-if="isDetail">
      <div class="page-head">
        <div><span class="eyebrow">用户与司机 / 司机详情</span><h1>{{ detail?.real_name || detail?.phone || '司机详情' }}</h1><p>查看司机基础资料与服务状态</p></div>
        <div class="actions"><el-button @click="router.back()">返回列表</el-button></div>
      </div>
      <div class="panel detail-grid">
        <div v-for="item in detailEntries" :key="item.key" class="detail-item">
          <span>{{ item.key }}</span><strong>{{ item.value }}</strong>
        </div>
      </div>
    </template>
    <template v-else>
      <div class="page-head">
        <div><span class="eyebrow">用户与司机 / 司机列表</span><h1>司机列表</h1><p>查询司机账号、冻结违规司机</p></div>
        <div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button></div>
      </div>
      <div class="panel filters">
        <el-input v-model="filters.keyword" clearable :prefix-icon="Search" placeholder="手机号、姓名或车牌" @keyup.enter="page=1;load" />
        <el-select v-model="filters.status" clearable placeholder="服务状态">
          <el-option :value="1" label="待审核" /><el-option :value="2" label="服务中" /><el-option :value="3" label="已冻结" />
        </el-select>
        <el-button type="primary" @click="page=1;load">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </div>
      <div class="panel table-panel">
        <el-table :data="rows" stripe empty-text="暂无司机">
          <el-table-column v-for="c in columns" :key="c[0]" :prop="c[0]" :label="c[1]" min-width="140">
            <template #default="scope">
              <span :class="{ mono: c[0].includes('id') || c[0].includes('no') }">{{ c[0] === 'status' ? driverStatusText(scope.row.status) : text(scope.row[c[0]]) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" fixed="right" width="200">
            <template #default="scope">
              <el-button link type="primary" @click="openDetail(scope.row)">查看</el-button>
              <el-button v-if="scope.row.status != 3" link type="warning" @click="action('freezeDriver', scope.row)">冻结</el-button>
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
.driver-page{color:var(--text-color,#2e2c4e)}.page-head{display:flex;align-items:flex-end;justify-content:space-between;margin-bottom:22px}.eyebrow{color:var(--brand,#6c5ce7);font-size:12px;letter-spacing:.1em;font-weight:600}.page-head h1{margin:7px 0 4px;font-size:26px;color:var(--text-color,#2e2c4e)}.page-head p{margin:0;color:var(--muted-color,#8b88a3)}.actions,.filters{display:flex;gap:10px;align-items:center}.panel{background:var(--panel-bg,#fff);border:1px solid var(--border-color,#e5e4f0);border-radius:14px;box-shadow:var(--card-shadow,none)}.filters{padding:18px;margin-bottom:16px;flex-wrap:wrap}.filters .el-input{width:240px}.filters .el-select{width:160px}.table-panel{padding:0;overflow:hidden}.table-footer{display:flex;justify-content:space-between;align-items:center;padding:14px 18px;color:var(--muted-color,#8b88a3);border-top:1px solid var(--border-color,#e5e4f0)}.mono{font-family:ui-monospace;color:var(--brand,#6c5ce7)}.detail-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:0;padding:0;overflow:hidden}.detail-item{padding:16px 20px;border-bottom:1px solid var(--border-color,#e5e4f0)}.detail-item span{display:block;margin-bottom:7px;color:var(--muted-color,#8b88a3);font-size:12px}.detail-item strong{display:block;overflow-wrap:anywhere;color:var(--text-color,#2e2c4e);font-weight:500}@media(max-width:1000px){.detail-grid{grid-template-columns:1fr}.page-head{display:block}.actions{margin-top:16px}.filters{flex-wrap:wrap}}
</style>
