<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { punishmentApi } from '../../api/modules'

const tab = ref('punishments')
const loading = ref(false)
const punishments = ref([])
const rules = ref([])
const appeals = ref([])
const dialog = ref(false)
const form = reactive({ driver_id: '', rule_id: '', actions: '["no_dispatch"]', reason: '', penalty_cents: 0, request_id: '' })

const load = async () => {
  loading.value = true
  try {
    const [p, r, a] = await Promise.all([
      punishmentApi.punishments({ page: 1, page_size: 50 }),
      punishmentApi.rules({ page: 1, page_size: 50 }),
      punishmentApi.appeals({ page: 1, page_size: 50 }),
    ])
    punishments.value = p?.list || p?.data?.list || []
    rules.value = r?.list || r?.data?.list || []
    appeals.value = a?.list || a?.data?.list || []
  } finally { loading.value = false }
}
const create = async () => {
  if (!form.request_id) form.request_id = `admin-${Date.now()}`
  await punishmentApi.createPunishment({ ...form, driver_id: Number(form.driver_id), rule_id: Number(form.rule_id || 0), penalty_cents: Number(form.penalty_cents) })
  ElMessage.success('处罚单已创建，正在异步执行')
  dialog.value = false
  await load()
}
const cancel = async (row) => {
  await ElMessageBox.confirm('确认撤销该处罚单？', '提示', { type: 'warning' })
  await punishmentApi.cancel(row.id, { reason: '后台撤销', request_id: `cancel-${row.id}-${Date.now()}` })
  ElMessage.success('撤销请求已提交')
  await load()
}
onMounted(load)
</script>

<template>
  <div class="page">
    <div class="toolbar"><h2>处罚管理</h2><el-button type="primary" @click="dialog = true">创建处罚</el-button></div>
    <el-tabs v-model="tab">
      <el-tab-pane label="处罚单" name="punishments"><el-table v-loading="loading" :data="punishments" stripe><el-table-column prop="punishment_no" label="处罚单号"/><el-table-column prop="driver_id" label="司机ID"/><el-table-column prop="actions" label="动作"/><el-table-column prop="status" label="状态"/><el-table-column label="操作"><template #default="{ row }"><el-button link type="danger" @click="cancel(row)">撤销</el-button></template></el-table-column></el-table></el-tab-pane>
      <el-tab-pane label="处罚规则" name="rules"><el-table :data="rules" stripe><el-table-column prop="name" label="规则名称"/><el-table-column prop="violation_type" label="违规类型"/><el-table-column prop="actions" label="动作"/><el-table-column prop="status" label="状态"/></el-table></el-tab-pane>
      <el-tab-pane label="处罚申诉" name="appeals"><el-table :data="appeals" stripe><el-table-column prop="appeal_no" label="申诉单号"/><el-table-column prop="punishment_id" label="处罚ID"/><el-table-column prop="driver_id" label="司机ID"/><el-table-column prop="status" label="状态"/><el-table-column prop="content" label="申诉内容"/></el-table></el-tab-pane>
    </el-tabs>
    <el-dialog v-model="dialog" title="创建处罚" width="520px"><el-form label-width="100px"><el-form-item label="司机ID"><el-input v-model="form.driver_id"/></el-form-item><el-form-item label="规则ID"><el-input v-model="form.rule_id"/></el-form-item><el-form-item label="处罚动作"><el-input v-model="form.actions"/></el-form-item><el-form-item label="处罚原因"><el-input v-model="form.reason" type="textarea"/></el-form-item><el-form-item label="罚款金额"><el-input-number v-model="form.penalty_cents" :min="0"/></el-form-item><el-form-item label="请求幂等号"><el-input v-model="form.request_id"/></el-form-item></el-form><template #footer><el-button @click="dialog = false">取消</el-button><el-button type="primary" @click="create">提交</el-button></template></el-dialog>
  </div>
</template>

<style scoped>
.page { min-height: 100%; }
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
h2 { margin: 0; color: var(--text-color, #2e2c4e); }
</style>
