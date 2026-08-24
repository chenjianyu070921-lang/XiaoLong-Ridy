<script setup>
// 工单证据列表：读取既有 evidence 查询接口，局部失败不会阻断工单详情和流转操作。
import { onMounted, ref, watch } from 'vue'
import { workOrdersApi } from '../api/modules'
import { pageData, text } from '../utils/format'

const props = defineProps({ workOrderId: { type: [Number, String], required: true }, refreshKey: { type: Number, default: 0 } })
const loading = ref(false)
const error = ref('')
const rows = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

// load 从后端读取证据分页；错误保留在局部区域，避免覆盖工单主体信息。
const load = async () => {
  loading.value = true
  error.value = ''
  try {
    const data = await workOrdersApi.evidenceList(props.workOrderId, { page: page.value, page_size: pageSize.value })
    const result = pageData(data)
    rows.value = result.list
    total.value = result.total
  } catch (err) {
    error.value = err?.message || '证据记录加载失败'
  } finally {
    loading.value = false
  }
}
watch(() => [props.workOrderId, props.refreshKey], () => { page.value = 1; load() })
onMounted(load)
</script>

<template>
  <section class="evidence" v-loading="loading"><div class="evidence-head"><h2>证据记录</h2><el-button link type="primary" @click="load">刷新</el-button></div><el-alert v-if="error" :title="error" type="error" show-icon :closable="false" /><el-table v-else :data="rows" empty-text="暂无证据记录"><el-table-column prop="evidence_type" label="证据类型" width="120"/><el-table-column prop="content" label="说明" min-width="180"><template #default="scope">{{ text(scope.row.content) }}</template></el-table-column><el-table-column prop="evidence_url" label="证据地址" min-width="240"><template #default="scope"><el-link v-if="scope.row.evidence_url" :href="scope.row.evidence_url" target="_blank" type="primary">查看证据</el-link><span v-else>-</span></template></el-table-column><el-table-column prop="created_at" label="提交时间" width="180"/></el-table><el-pagination v-if="total > pageSize" v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="prev, pager, next" @current-change="load" /></section>
</template>

<style scoped>
.evidence{margin-top:16px;padding:18px;background:#101d2b;border:1px solid #1d3042;border-radius:8px}.evidence-head{display:flex;align-items:center;justify-content:space-between;margin-bottom:14px}.evidence-head h2{margin:0;color:#f4f7fb;font-size:16px}.evidence :deep(.el-pagination){margin-top:14px}
</style>
