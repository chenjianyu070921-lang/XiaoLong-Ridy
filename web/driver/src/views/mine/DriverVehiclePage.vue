<template>
  <main class="mine-page-shell">
    <header class="mine-page-header">
      <button type="button" class="page-back" aria-label="返回" @click="goHome">
        <van-icon name="arrow-left" />
      </button>
      <div class="page-heading">
        <p>我的</p>
        <h1>车辆信息</h1>
      </div>
      <button type="button" class="page-action" @click="loadVehicle">
        <van-icon name="replay" />
      </button>
    </header>

    <DriverVehiclePanel
      :driver-store="driverStore"
      :vehicle-form="vehicleForm"
      :format-vehicle-status="formatVehicleStatus"
      @load-vehicle="loadVehicle"
      @submit-vehicle="submitVehicle"
      @submit-vehicle-update="submitVehicleUpdate"
      @remove-vehicle="removeVehicle"
    />
  </main>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import DriverVehiclePanel from '@/components/driver-home/DriverVehiclePanel.vue'
import { useDriverAssets } from '@/composables/useDriverAssets'
import { formatVehicleStatus } from '@/utils/driver-format'
import '@/styles/driver-home-panels.css'
import '@/styles/driver-mine-pages.css'

const router = useRouter()
const {
  driverStore,
  vehicleForm,
  loadVehicle,
  submitVehicle,
  submitVehicleUpdate,
  removeVehicle
} = useDriverAssets()

onMounted(() => {
  void driverStore.refreshProfile({ silentError: true })
  void loadVehicle({ silentError: true })
})

function goHome() {
  router.back()
}
</script>
