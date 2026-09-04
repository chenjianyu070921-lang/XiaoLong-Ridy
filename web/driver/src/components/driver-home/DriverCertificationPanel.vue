<template>
  <section class="h5-panel cert-panel">
    <div class="section-title">
      <h2>车辆资质</h2>
      <button type="button" @click="$emit('load-certification')">刷新</button>
    </div>
    <div v-if="driverStore.certification" class="cert-status-card" :class="'status-' + (driverStore.certification.auditStatus || 0)">
      <div class="cert-status-icon"><van-icon :name="certStatusIcon" /></div>
      <div class="cert-status-info">
        <strong>{{ formatCertificationStatus(driverStore.certification.auditStatus) }}</strong>
        <span v-if="driverStore.certification.auditRemark">{{ driverStore.certification.auditRemark }}</span>
      </div>
    </div>
    <div class="cert-vehicle-card">
      <div class="cert-vehicle-label">认证车辆</div>
      <div v-if="driverStore.vehicle" class="cert-vehicle-info">
        <span class="cert-plate">{{ driverStore.vehicle.plateNo || '未绑定' }}</span>
        <span class="cert-vehicle-model">{{ driverStore.vehicle.brand || '' }} {{ driverStore.vehicle.model || '' }}</span>
      </div>
      <div v-else class="cert-vehicle-empty"><span>请先在“车辆”页绑定车辆</span></div>
    </div>
    <div class="cert-form">
      <van-field v-model="certificationForm.idCardNo" label="身份证号" placeholder="请输入身份证号" clearable />
      <van-field v-model="certificationForm.realName" label="真实姓名" placeholder="请输入真实姓名" clearable />
      <van-field v-model="certificationForm.driverLicenseNo" label="驾照编号" placeholder="请输入驾驶证编号" clearable />
    </div>
    <p class="cert-tip">资质校验以身份证号、真实姓名与驾驶证编号进行，提交后进入人工审核。</p>
    <button class="primary-action cert-submit-btn" :disabled="certSubmitting" type="button" @click="$emit('submit-certification')">
      {{ certSubmitting ? '提交中...' : '提交资质审核' }}
    </button>
  </section>
</template>

<script setup>
defineProps({
  driverStore: { type: Object, required: true },
  certificationForm: { type: Object, required: true },
  certSubmitting: { type: Boolean, required: true },
  certStatusIcon: { type: String, required: true },
  formatCertificationStatus: { type: Function, required: true }
})

defineEmits(['load-certification', 'submit-certification'])
</script>
