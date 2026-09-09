<script setup>
// 管理员注册页：字段严格对齐 RegisterRequest，注册成功后复用服务端返回的登录会话。
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { register } from '../../api/auth'
import { useUserStore } from '../../store/user'

const router = useRouter()
const store = useUserStore()
const loading = ref(false)
const form = reactive({ username: '', password: '', confirmPassword: '', real_name: '', role: 2 })

// submit 校验两次密码一致后注册；是否允许创建后续管理员由后端角色权限统一裁决。
const submit = async () => {
  if (!form.username || !form.password || !form.real_name) return ElMessage.warning('请完整填写管理员信息')
  if (form.password !== form.confirmPassword) return ElMessage.warning('两次输入的密码不一致')
  loading.value = true
  try {
    const data = await register({ username: form.username, password: form.password, real_name: form.real_name, role: form.role })
    store.setLogin(data.token, data.admin)
    ElMessage.success('注册成功')
    router.replace('/dashboard')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="register-page"><div class="register-shell"><div class="brand"><img class="mark" src="/huaxiaolong-logo.png" alt="花小龙出行" /><div><h1>注册管理员</h1><p>花小龙出行运营管理中心</p></div></div><el-form :model="form" @submit.prevent="submit"><el-form-item label="管理员账号"><el-input v-model="form.username" autocomplete="username" /></el-form-item><el-form-item label="姓名"><el-input v-model="form.real_name" /></el-form-item><el-form-item label="角色"><el-select v-model="form.role"><el-option label="超级管理员" :value="1"/><el-option label="运营人员" :value="2"/><el-option label="客服人员" :value="3"/></el-select></el-form-item><el-form-item label="登录密码"><el-input v-model="form.password" type="password" show-password autocomplete="new-password" /></el-form-item><el-form-item label="确认密码"><el-input v-model="form.confirmPassword" type="password" show-password autocomplete="new-password" @keyup.enter="submit" /></el-form-item><el-button type="primary" :loading="loading" class="register-button" @click="submit">提交注册</el-button><el-button text class="back-button" @click="router.push('/login')">返回登录</el-button></el-form></div></main>
</template>

<style scoped>
.register-page{min-height:100vh;display:grid;place-items:center;background:linear-gradient(135deg,#6a5ae2 0%,#7f6cec 45%,#9a8ff2 100%)}.register-shell{width:min(480px,calc(100% - 40px));padding:40px;background:var(--panel-bg,#fff);border:1px solid var(--border-color,#e5e4f0);border-radius:18px;box-shadow:0 30px 80px rgba(40,28,110,.35)}.brand{display:flex;align-items:center;gap:12px;margin-bottom:30px}.mark{width:42px;height:42px;object-fit:cover;background:#6c5ce7;border-radius:12px}.brand h1{margin:0;color:var(--text-color,#2e2c4e);font-size:22px}.brand p{margin:3px 0 0;color:var(--muted-color,#8b88a3)}.register-button{width:100%}.back-button{display:block;margin:16px auto 0}
</style>
