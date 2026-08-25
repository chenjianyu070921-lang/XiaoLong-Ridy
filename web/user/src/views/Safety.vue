<template><div class="safety-page"><van-nav-bar title="安全保障" left-arrow @click-left="$router.back()"/><div class="hero"><van-icon name="shield-o" size="48"/><h2>安全出行指南</h2><p>多重安全保障，守护每一次出行</p></div><div class="list"><div class="row"><span>隐私号码保护</span><van-switch v-model="privacy" size="20"/></div><div class="row"><span>异常行程自动分享</span><van-switch v-model="shareTrip" size="20"/></div><div class="row clickable" @click="addContact"><span>紧急联系人</span><span class="value">{{ contact || '未设置' }} <van-icon name="arrow"/></span></div><div class="row"><span>安全专线</span><span class="value">24小时服务</span></div></div><button class="alarm" @click="showToast('报警信息已提交')"><van-icon name="warning-o"/> 一键报警</button></div></template>
<script setup>
import { ref } from 'vue'
import { showToast, showConfirmDialog } from 'vant'
const privacy=ref(true); const shareTrip=ref(false); const contact=ref(localStorage.getItem('emergency-contact')||'')
const addContact=async()=>{const result=prompt('请输入紧急联系人手机号'); if(result){contact.value=result;localStorage.setItem('emergency-contact',result);showToast('紧急联系人已添加')}}
</script>
<style scoped>
.safety-page{min-height:100vh;background:#eef0ff}.hero{padding:30px 24px;color:#fff;background:linear-gradient(135deg,#6366f1,#9333ea)}.hero h2{margin:12px 0 6px}.hero p{margin:0;opacity:.9}.list{padding:12px 20px}.row{display:flex;justify-content:space-between;align-items:center;padding:20px 0;border-bottom:1px solid #dfe3f4;font-size:16px}.value{color:#9ca3af}.clickable{cursor:pointer}.alarm{position:fixed;bottom:28px;left:24px;right:24px;padding:15px;border:0;border-radius:14px;color:#fff;background:#ff3d61;font-size:17px;font-weight:700}
</style>