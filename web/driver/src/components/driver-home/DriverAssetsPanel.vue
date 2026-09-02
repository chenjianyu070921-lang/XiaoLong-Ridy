<template>
  <section class="group-panel">
    <div class="panel-segments">
      <button type="button" :class="{ active: activeSection === 'wallet' }" @click="activeSection = 'wallet'">钱包</button>
      <button type="button" :class="{ active: activeSection === 'vehicle' }" @click="activeSection = 'vehicle'">车辆</button>
      <button type="button" :class="{ active: activeSection === 'certification' }" @click="activeSection = 'certification'">资质</button>
    </div>

    <DriverWalletPanel
      v-if="activeSection === 'wallet'"
      :income-summary="incomeSummary"
      :today-income="todayIncome"
      :week-income="weekIncome"
      :income-bills="incomeBills"
      :format-price="formatPrice"
      :format-time="formatTime"
      @load-income="$emit('load-income')"
      @open-withdraw="$emit('open-withdraw')"
    />
    <DriverVehiclePanel
      v-else-if="activeSection === 'vehicle'"
      :driver-store="driverStore"
      :vehicle-form="vehicleForm"
      :format-vehicle-status="formatVehicleStatus"
      @load-vehicle="$emit('load-vehicle')"
      @submit-vehicle="$emit('submit-vehicle')"
      @submit-vehicle-update="$emit('submit-vehicle-update')"
      @remove-vehicle="$emit('remove-vehicle')"
    />
    <DriverCertificationPanel
      v-else
      :driver-store="driverStore"
      :certification-form="certificationForm"
      :cert-uploading="certUploading"
      :cert-items="certItems"
      :cert-submitting="certSubmitting"
      :cert-status-icon="certStatusIcon"
      :format-certification-status="formatCertificationStatus"
      @load-certification="$emit('load-certification')"
      @read-cert-file="(...args) => $emit('read-cert-file', ...args)"
      @remove-cert-image="$emit('remove-cert-image', $event)"
      @submit-certification="$emit('submit-certification')"
    />
  </section>
</template>

<script setup>
import { ref } from 'vue'
import DriverWalletPanel from './DriverWalletPanel.vue'
import DriverVehiclePanel from './DriverVehiclePanel.vue'
import DriverCertificationPanel from './DriverCertificationPanel.vue'

defineProps({
  driverStore: { type: Object, required: true },
  vehicleForm: { type: Object, required: true },
  certificationForm: { type: Object, required: true },
  certUploading: { type: Object, required: true },
  certItems: { type: Array, required: true },
  certSubmitting: { type: Boolean, required: true },
  certStatusIcon: { type: String, required: true },
  incomeSummary: { type: Object, required: true },
  todayIncome: { type: Object, required: true },
  weekIncome: { type: Object, required: true },
  incomeBills: { type: Array, required: true },
  formatPrice: { type: Function, required: true },
  formatTime: { type: Function, required: true },
  formatVehicleStatus: { type: Function, required: true },
  formatCertificationStatus: { type: Function, required: true }
})

defineEmits([
  'load-income',
  'open-withdraw',
  'load-vehicle',
  'submit-vehicle',
  'submit-vehicle-update',
  'remove-vehicle',
  'load-certification',
  'read-cert-file',
  'remove-cert-image',
  'submit-certification'
])

const activeSection = ref('wallet')
</script>
