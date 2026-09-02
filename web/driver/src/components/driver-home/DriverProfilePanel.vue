<template>
  <section class="h5-panel">
    <div class="section-title">
      <h2>我的</h2>
      <button type="button" @click="$emit('refresh-dashboard')">刷新</button>
    </div>
    <div class="info-list">
      <p><b>司机ID</b><span>{{ driverStore.driverId || '--' }}</span></p>
      <p><b>手机号</b><span>{{ driverStore.driver.phone || '--' }}</span></p>
      <p><b>姓名</b><span>{{ driverStore.driver.realName || '--' }}</span></p>
      <p><b>身份证</b><span>{{ driverStore.driver.idCardNo || '--' }}</span></p>
      <p><b>驾驶证</b><span>{{ driverStore.driver.driverLicenseNo || '--' }}</span></p>
      <p><b>状态</b><span>{{ formatDriverStatus(driverStore.driver.status) }}</span></p>
    </div>
    <button class="primary-action" type="button" @click="$emit('edit-profile')">编辑资料</button>
  </section>
</template>

<script setup>
defineProps({
  driverStore: { type: Object, required: true },
  formatDriverStatus: { type: Function, required: true }
})

defineEmits(['refresh-dashboard', 'edit-profile'])
</script>
