<template>
  <section class="group-panel">
    <div class="panel-segments">
      <button type="button" :class="{ active: activeSection === 'profile' }" @click="activeSection = 'profile'">资料</button>
      <button type="button" :class="{ active: activeSection === 'reviews' }" @click="activeSection = 'reviews'">评价</button>
      <button type="button" :class="{ active: activeSection === 'trajectory' }" @click="activeSection = 'trajectory'">轨迹</button>
    </div>

    <DriverProfilePanel
      v-if="activeSection === 'profile'"
      :driver-store="driverStore"
      :profile-form="profileForm"
      :format-driver-status="formatDriverStatus"
      @refresh-dashboard="$emit('refresh-dashboard')"
      @submit-profile="$emit('submit-profile')"
    />
    <DriverReviewsPanel
      v-else-if="activeSection === 'reviews'"
      :reviews="reviews"
      :format-time="formatTime"
      @load-reviews="$emit('load-reviews')"
    />
    <DriverTrajectoryPanel
      v-else
      :trajectory-order-id="trajectoryOrderId"
      :trajectory-error="trajectoryError"
      :trajectory-points="trajectoryPoints"
      :format-time="formatTime"
      @update:trajectory-order-id="$emit('update:trajectoryOrderId', $event)"
      @load-trajectory="$emit('load-trajectory')"
    />
  </section>
</template>

<script setup>
import { ref } from 'vue'
import DriverProfilePanel from './DriverProfilePanel.vue'
import DriverReviewsPanel from './DriverReviewsPanel.vue'
import DriverTrajectoryPanel from './DriverTrajectoryPanel.vue'

defineProps({
  driverStore: { type: Object, required: true },
  profileForm: { type: Object, required: true },
  reviews: { type: Array, required: true },
  trajectoryOrderId: { type: [String, Number], required: true },
  trajectoryError: { type: String, required: true },
  trajectoryPoints: { type: Array, required: true },
  formatDriverStatus: { type: Function, required: true },
  formatTime: { type: Function, required: true }
})

defineEmits([
  'refresh-dashboard',
  'submit-profile',
  'load-reviews',
  'update:trajectoryOrderId',
  'load-trajectory'
])

const activeSection = ref('profile')
</script>
