<template>
  <section class="h5-panel">
    <div class="section-title">
      <h2>车辆</h2>
      <button type="button" @click="$emit('load-vehicle')">查询</button>
    </div>
    <div class="info-list">
      <p><b>车辆ID</b><span>{{ driverStore.vehicle?.id || '--' }}</span></p>
      <p><b>车牌号</b><span>{{ driverStore.vehicle?.plateNo || '--' }}</span></p>
      <p><b>车型</b><span>{{ driverStore.vehicle?.brand || '--' }} {{ driverStore.vehicle?.model || '' }}</span></p>
      <p><b>状态</b><span>{{ formatVehicleStatus(driverStore.vehicle?.status) }}</span></p>
    </div>
    <van-form class="form-stack">
      <van-field v-model="vehicleForm.plateNo" label="车牌号" placeholder="粤B12345" />
      <van-field v-model="vehicleForm.brand" label="品牌" placeholder="BYD" />
      <van-field v-model="vehicleForm.model" label="型号" placeholder="Han" />
      <van-field v-model="vehicleForm.color" label="颜色" placeholder="黑色" />
      <van-field v-model.number="vehicleForm.vehicleType" type="number" label="车辆类型" placeholder="1" />
      <van-field v-model="vehicleForm.registrationDate" type="date" label="注册日期" />
      <van-field v-model="vehicleForm.insuranceNo" label="保险单号" placeholder="INS-001" />
      <van-field v-model="vehicleForm.insuranceExpireAt" type="date" label="保险到期" />
      <button class="primary-action" type="button" @click="$emit('submit-vehicle')">提交车辆</button>
    </van-form>
    <div class="two-actions">
      <button type="button" @click="$emit('submit-vehicle-update')">更新</button>
      <button type="button" class="danger" @click="$emit('remove-vehicle')">删除</button>
    </div>
  </section>
</template>

<script setup>
defineProps({
  driverStore: { type: Object, required: true },
  vehicleForm: { type: Object, required: true },
  formatVehicleStatus: { type: Function, required: true }
})

defineEmits(['load-vehicle', 'submit-vehicle', 'submit-vehicle-update', 'remove-vehicle'])
</script>
