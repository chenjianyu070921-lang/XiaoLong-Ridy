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
    <div class="cert-upload-grid">
      <div v-for="item in certItems" :key="item.key" class="cert-upload-card">
        <div class="cert-upload-header">
          <span class="cert-upload-title">{{ item.title }}</span>
          <span class="cert-upload-tip">{{ item.tip }}</span>
        </div>
        <div class="cert-upload-area" :class="{ uploaded: certificationForm[item.key], uploading: certUploading[item.key] }" @click="triggerUpload(item.key)">
          <template v-if="certUploading[item.key]">
            <van-loading size="24px" color="#6d4aff" />
            <span class="cert-upload-text">上传中...</span>
          </template>
          <template v-else-if="certificationForm[item.key]">
            <img :src="certificationForm[item.key]" :alt="item.title" class="cert-preview-img" />
            <div class="cert-upload-mask"><van-icon name="photograph" /><span>重新上传</span></div>
            <button type="button" class="cert-delete-btn" @click.stop="$emit('remove-cert-image', item.key)"><van-icon name="cross" /></button>
          </template>
          <template v-else>
            <van-icon name="photograph" class="cert-upload-icon" />
            <span class="cert-upload-text">点击上传</span>
          </template>
        </div>
        <input :ref="el => fileRefs[item.key] = el" type="file" accept="image/*" class="cert-file-input" @change="$emit('read-cert-file', $event, item.key)" />
      </div>
    </div>
    <button class="primary-action cert-submit-btn" :disabled="certSubmitting" type="button" @click="$emit('submit-certification')">
      {{ certSubmitting ? '提交中...' : '提交资质审核' }}
    </button>
  </section>
</template>

<script setup>
import { reactive } from 'vue'

defineProps({
  driverStore: { type: Object, required: true },
  certificationForm: { type: Object, required: true },
  certUploading: { type: Object, required: true },
  certItems: { type: Array, required: true },
  certSubmitting: { type: Boolean, required: true },
  certStatusIcon: { type: String, required: true },
  formatCertificationStatus: { type: Function, required: true }
})

defineEmits(['load-certification', 'read-cert-file', 'remove-cert-image', 'submit-certification'])

const fileRefs = reactive({})

function triggerUpload(field) {
  fileRefs[field]?.click()
}
</script>
