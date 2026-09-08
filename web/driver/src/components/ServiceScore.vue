<template>
  <span class="svc-score" :style="{ '--svc-color': color }">
    <svg class="svc-ring" :width="size" :height="size" viewBox="0 0 36 36">
      <circle class="svc-track" cx="18" cy="18" :r="radius" />
      <circle
        class="svc-bar"
        cx="18"
        cy="18"
        :r="radius"
        :stroke-dasharray="`${circumference} ${circumference}`"
        :stroke-dashoffset="dashOffset"
        transform="rotate(-90 18 18)"
      />
    </svg>
    <span class="svc-center">
      <b>{{ display }}</b>
      <small v-if="level">{{ level }}</small>
    </span>
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  score: { type: [Number, String], default: null },
  size: { type: Number, default: 40 }
})

const radius = 15.9
const circumference = 2 * Math.PI * radius

const numeric = computed(() => {
  const n = Number(props.score)
  return Number.isFinite(n) && n > 0 ? n : null
})
const percent = computed(() => {
  if (numeric.value == null) return 0
  return Math.max(0, Math.min(100, numeric.value))
})
const dashOffset = computed(() => circumference * (1 - percent.value / 100))
const display = computed(() => (numeric.value == null ? '--' : String(Math.round(numeric.value))))
const level = computed(() => {
  if (numeric.value == null) return ''
  if (numeric.value >= 90) return '优'
  if (numeric.value >= 75) return '良'
  if (numeric.value >= 60) return '中'
  return '差'
})
const color = computed(() => {
  if (numeric.value == null) return '#9aa3b2'
  if (numeric.value >= 90) return '#22c55e'
  if (numeric.value >= 75) return '#6366f1'
  if (numeric.value >= 60) return '#f59e0b'
  return '#ef4444'
})
</script>

<style scoped>
.svc-score { position: relative; display: inline-flex; align-items: center; justify-content: center; }
.svc-track { fill: none; stroke: var(--driver-track); stroke-width: 3.2; }
.svc-bar { fill: none; stroke: var(--svc-color); stroke-width: 3.2; stroke-linecap: round; transition: stroke-dashoffset 0.4s ease; }
.svc-center { position: absolute; display: flex; flex-direction: column; align-items: center; line-height: 1; }
.svc-center b { font-size: 13px; color: var(--svc-color); font-weight: 700; }
.svc-center small { font-size: 9px; color: var(--svc-color); margin-top: 1px; }
</style>
