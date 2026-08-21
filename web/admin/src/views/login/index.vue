<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login } from '../../api/auth'
import { useUserStore } from '../../store/user'
const router = useRouter(); const store = useUserStore(); const loading = ref(false); const form = reactive({ username: '', password: '' })
const submit = async () => { if (!form.username || !form.password) return ElMessage.warning('请输入账号和密码'); loading.value = true; try { const data = await login(form); store.setLogin(data.token, data.admin); router.push('/dashboard') } finally { loading.value = false } }
</script>
<template><main class="login-page"><div class="login-shell"><div class="brand"><div class="mark">行</div><div><h1>小隆出行</h1><p>运营管理中心</p></div></div><div class="login-copy"><span>OPERATIONS CONSOLE</span><h2>让每一次出行<br>都可被清晰管理</h2></div><el-form :model="form" @submit.prevent="submit"><el-form-item><el-input v-model="form.username" placeholder="管理员账号" size="large" /></el-form-item><el-form-item><el-input v-model="form.password" type="password" show-password placeholder="登录密码" size="large" @keyup.enter="submit" /></el-form-item><el-button type="primary" size="large" :loading="loading" class="login-button" @click="submit">进入管理后台</el-button></el-form><small class="copyright">小隆出行 · 管理后台</small></div></main></template>
<style scoped>.login-page{min-height:100vh;background:#07101a;display:grid;place-items:center}.login-shell{width:min(420px,calc(100% - 40px));background:#0e1b29;border:1px solid #1d3041;border-radius:14px;padding:44px;box-shadow:0 25px 70px #0008}.brand{display:flex;align-items:center;gap:12px}.mark{width:42px;height:42px;border-radius:12px;background:#ff7625;color:#161c22;display:grid;place-items:center;font-size:22px;font-weight:800}.brand h1{margin:0;color:#f4f8fb;font-size:22px}.brand p{margin:3px 0 0;color:#8397a9}.login-copy{margin:54px 0 30px}.login-copy span{font-size:11px;color:#ff873b;letter-spacing:.16em}.login-copy h2{color:#fff;font-size:29px;line-height:1.25;margin:10px 0}.login-button{width:100%;margin-top:6px}.copyright{display:block;text-align:center;color:#657a8e;margin-top:34px}</style>
