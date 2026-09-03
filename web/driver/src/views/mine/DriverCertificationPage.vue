<template>
  <main class="mine-page-shell">
    <header class="mine-page-header">
      <button type="button" class="page-back" aria-label="返回" @click="goHome">
        <van-icon name="arrow-left" />
      </button>
      <div class="page-heading">
        <p>我的</p>
        <h1>资质认证</h1>
      </div>
      <button type="button" class="page-action" @click="loadCertification">
        <van-icon name="replay" />
      </button>
    </header>

    <DriverCertificationPanel
      :driver-store="driverStore"
      :certification-form="certificationForm"
      :cert-uploading="certUploading"
      :cert-items="certItems"
      :cert-submitting="certSubmitting"
      :cert-status-icon="certStatusIcon"
      :format-certification-status="formatCertificationStatus"
      @load-certification="loadCertification"
      @read-cert-file="readCertFile"
      @remove-cert-image="removeCertImage"
      @submit-certification="submitCertification"
    />
  </main>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import DriverCertificationPanel from '@/components/driver-home/DriverCertificationPanel.vue'
import { useDriverAssets } from '@/composables/useDriverAssets'
import { formatCertificationStatus } from '@/utils/driver-format'
import '@/styles/driver-home-panels.css'
import '@/styles/driver-mine-pages.css'

const router = useRouter()
const {
  driverStore,
  certificationForm,
  certUploading,
  certItems,
  certSubmitting,
  certStatusIcon,
  loadCertification,
  readCertFile,
  removeCertImage,
  submitCertification
} = useDriverAssets()

onMounted(() => {
  void driverStore.refreshProfile({ silentError: true })
  void loadCertification({ silentError: true })
})

function goHome() {
  router.back()
}
</script>
