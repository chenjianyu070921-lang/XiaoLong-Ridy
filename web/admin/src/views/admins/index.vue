<script setup>
// 超级管理员专属账号管理页：所有写操作均由后端再次校验当前角色。
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import { adminsApi } from '../../api/modules'
import { useUserStore } from '../../store/user'
import { roleText } from '../../utils/enums'

const store = useUserStore()
const loading = ref(false)
const dialogVisible = ref(false)
const dialogMode = ref('create')
const submitting = ref(false)
const rows = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', role: undefined, status: undefined })
const form = reactive({ id: 0, username: '', real_name: '', password: '', role: 2 })
const isSuperAdmin = computed(() => store.admin?.role === 1)

const resetForm = () => Object.assign(form, { id: 0, username: '', real_name: '', password: '', role: 2 })
const load = async () => {
  loading.value = true
  try {
    const data = await adminsApi.list({ page: page.value, page_size: pageSize.value, ...filters })
    rows.value = data?.list || []
    total.value = data?.total || 0
  } finally {
    loading.value = false
  }
}
const openCreate = () => { resetForm(); dialogMode.value = 'create'; dialogVisible.value = true }
const openEdit = (row) => {
  Object.assign(form, { id: row.id, username: row.username, real_name: row.real_name, password: '', role: row.role })
  dialogMode.value = 'edit'
  dialogVisible.value = true
}
const submit = async () => {
  if (!form.real_name || !form.role || (dialogMode.value === 'create' && (!form.username || form.password.length < 6))) {
    ElMessage.warning(dialogMode.value === 'create' ? '请完整填写账号、姓名和至少 6 位密码' : '请填写姓名和角色')
    return
  }
  submitting.value = true
  try {
    if (dialogMode.value === 'create') await adminsApi.create({ username: form.username, real_name: form.real_name, password: form.password, role: form.role })
    else await adminsApi.update(form.id, { real_name: form.real_name, role: form.role })
    dialogVisible.value = false
    ElMessage.success(dialogMode.value === 'create' ? '管理员创建成功' : '管理员资料已更新')
    await load()
  } finally { submitting.value = false }
}
const toggleStatus = async (row) => {
  await ElMessageBox.confirm(`确认${row.status === 1 ? '停用' : '启用'}管理员“${row.username}”吗？`, '操作确认', { type: 'warning' })
  await adminsApi.setStatus(row.id, { status: row.status === 1 ? 2 : 1, reason: row.status === 1 ? '超级管理员停用' : '超级管理员启用' })
  ElMessage.success('状态已更新')
  await load()
}
const resetPassword = async (row) => {
  const password = window.prompt(`请输入“${row.username}”的新密码（至少 6 位）`)
  if (!password) return
  if (password.length < 6) return ElMessage.warning('密码长度不能少于 6 位')
  await adminsApi.resetPassword(row.id, { password })
  ElMessage.success('密码已重置')
}
onMounted(() => { if (isSuperAdmin.value) load() })
</script>

<template>
  <section class="admin-page">
    <div v-if="!isSuperAdmin" class="forbidden">只有超级管理员可以访问管理员管理。</div>
    <template v-else>
      <div class="page-head">
        <div><span class="eyebrow">系统设置 / 权限主体</span><h1>管理员管理</h1><p>统一管理后台账号、角色和登录状态</p></div>
        <div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button><el-button type="primary" :icon="Plus" @click="openCreate">新增管理员</el-button></div>
      </div>
      <div class="panel filters">
        <el-input v-model="filters.keyword" clearable :prefix-icon="Search" placeholder="账号或姓名" @keyup.enter="page=1;load" />
        <el-select v-model="filters.role" clearable placeholder="角色"><el-option :value="1" label="超级管理员" /><el-option :value="2" label="运营" /><el-option :value="3" label="客服" /></el-select>
        <el-select v-model="filters.status" clearable placeholder="状态"><el-option :value="1" label="启用" /><el-option :value="2" label="停用" /></el-select>
        <el-button type="primary" @click="page=1;load">查询</el-button>
      </div>
      <div class="panel table-panel">
        <el-table v-loading="loading" :data="rows" stripe empty-text="暂无管理员">
          <el-table-column prop="id" label="ID" width="90" />
          <el-table-column prop="username" label="账号" min-width="180" />
          <el-table-column prop="real_name" label="姓名" min-width="160" />
          <el-table-column label="角色" min-width="160"><template #default="{ row }">{{ roleText(row.role) }}</template></el-table-column>
          <el-table-column label="状态" min-width="120"><template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag></template></el-table-column>
          <el-table-column label="操作" fixed="right" width="330">
            <template #default="{ row }">
              <el-button link type="primary" :disabled="row.id === store.admin?.id" @click="openEdit(row)">编辑</el-button>
              <el-button link type="warning" :disabled="row.id === store.admin?.id" @click="toggleStatus(row)">{{ row.status === 1 ? '停用' : '启用' }}</el-button>
              <el-button link type="danger" :disabled="row.id === store.admin?.id" @click="resetPassword(row)">重置密码</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="table-footer"><span>共 {{ total }} 条记录</span><el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="prev, pager, next, sizes" @current-change="load" @size-change="load" /></div>
      </div>
    </template>
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新增管理员' : '编辑管理员'" width="460px">
      <el-form label-width="90px">
        <el-form-item v-if="dialogMode === 'create'" label="登录账号"><el-input v-model="form.username" /></el-form-item>
        <el-form-item label="姓名"><el-input v-model="form.real_name" /></el-form-item>
        <el-form-item label="角色"><el-select v-model="form.role"><el-option :value="1" label="超级管理员" /><el-option :value="2" label="运营" /><el-option :value="3" label="客服" /></el-select></el-form-item>
        <el-form-item v-if="dialogMode === 'create'" label="初始密码"><el-input v-model="form.password" type="password" show-password /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" :loading="submitting" @click="submit">保存</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.admin-page{color:#dce7f3}.page-head{display:flex;align-items:flex-end;justify-content:space-between;margin-bottom:24px}.eyebrow{color:#6f879d;font-size:12px;letter-spacing:.08em}.page-head h1{margin:7px 0 4px;font-size:28px;color:#f4f7fb}.page-head p{margin:0;color:#8293a5}.actions,.filters{display:flex;gap:10px;align-items:center}.panel{background:#101d2b;border:1px solid #1d3042;border-radius:10px}.filters{padding:18px;margin-bottom:16px}.filters .el-input{width:280px}.filters .el-select{width:160px}.table-panel{overflow:hidden}.table-footer{display:flex;justify-content:space-between;align-items:center;padding:14px 18px;color:#8092a4}.forbidden{padding:40px;background:#101d2b;border:1px solid #1d3042;border-radius:10px;color:#f5a623}@media(max-width:900px){.page-head{display:block}.actions{margin-top:16px}.filters{flex-wrap:wrap}}
</style>
