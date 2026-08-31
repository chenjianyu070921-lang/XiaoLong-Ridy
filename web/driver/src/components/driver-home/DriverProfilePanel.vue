<template>
  <section class="h5-panel">
    <div class="section-title">
      <h2>我的</h2>
      <button type="button" @click="$emit('refresh-dashboard')">刷新</button>
    </div>
    <div class="info-list">
      <p><b>司机ID</b><span>{{ driverStore.driverId || '--' }}</span></p>
      <p><b>手机号</b><span>{{ driverStore.driver.phone || '--' }}</span></p>
      <p><b>状态</b><span>{{ formatDriverStatus(driverStore.driver.status) }}</span></p>
    </div>
    <van-form class="form-stack" @submit="$emit('submit-profile')">
      <van-field v-model="profileForm.realName" label="姓名" placeholder="司机姓名" />
      <van-field v-model="profileForm.avatarUrl" label="头像地址" placeholder="头像 URL" />
      <van-field v-model="profileForm.driverLicenseNo" label="驾驶证号" placeholder="驾驶证号" />
      <button class="primary-action" type="submit">保存资料</button>
    </van-form>
  </section>
</template>

<script setup>
defineProps({
  driverStore: { type: Object, required: true },
  profileForm: { type: Object, required: true },
  formatDriverStatus: { type: Function, required: true }
})

defineEmits(['refresh-dashboard', 'submit-profile'])
</script>
