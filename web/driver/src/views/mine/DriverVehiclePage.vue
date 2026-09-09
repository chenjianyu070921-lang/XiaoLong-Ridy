<template>
  <main class="mine-page-shell">
    <header class="mine-page-header">
      <button type="button" class="page-back" aria-label="返回" @click="goBack">
        <van-icon name="arrow-left" />
      </button>
      <div class="page-heading">
        <p>我的</p>
        <h1>车辆信息</h1>
      </div>
    </header>
    <DriverVehiclePanel />
  </main>
</template>

<script setup>
import { useRouter } from 'vue-router'
import DriverVehiclePanel from '@/components/driver-home/DriverVehiclePanel.vue'
import '@/styles/driver-mine-pages.css'

const router = useRouter()
function goBack() {
  router.back()
}
</script>
