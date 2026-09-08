<template>
  <div class="location-panel">
    <div class="panel-head">
      <span class="panel-label">{{ type === 'pickup' ? '上车点' : '目的地' }}</span>
      <span v-if="resolving" class="panel-loading"><van-loading size="14px" />正在获取地址...</span>
    </div>

    <transition name="fade" mode="out-in">
      <div :key="location?.name || 'empty'" class="panel-body">
        <p class="location-name">{{ location?.name || (resolving ? '正在获取位置...' : '拖动地图，把图钉移到目标位置') }}</p>
        <p v-if="location?.address" class="location-address">{{ location.address }}</p>
      </div>
    </transition>

    <button
      type="button"
      class="btn-primary confirm-btn"
      :disabled="!canConfirm || confirming"
      @click="confirm"
    >
      {{ confirming ? '确认中...' : (type === 'pickup' ? '确认上车点' : '确认目的地') }}
    </button>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { hasValidCoordinate } from '@/utils/location'

const props = defineProps({
  type: { type: String, default: 'destination' },
  location: { type: Object, default: null },
  resolving: { type: Boolean, default: false }
})

const emit = defineEmits(['confirm'])

// 仅当已解析出有效坐标与名称时才允许确认，避免提交空地点。
const canConfirm = computed(() => Boolean(props.location?.name && hasValidCoordinate(props.location.latitude, props.location.longitude)))

// confirming 用于阻止用户连续点击确认按钮，emit 后由父组件执行返回，组件随即销毁。
const confirming = ref(false)

function confirm() {
  if (!canConfirm.value || confirming.value) return
  confirming.value = true
  emit('confirm')
}
</script>

<style scoped>
.location-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.panel-label {
  color: #7C3AED;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 1px;
}

.panel-loading {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #6B7280;
  font-size: 12px;
}

.panel-body {
  margin: 8px 0 12px;
  min-height: 58px;
}

.location-name {
  font-size: 17px;
  font-weight: 600;
  color: #111827;
  line-height: 1.35;
}

.location-address {
  margin-top: 4px;
  overflow: hidden;
  color: #6B7280;
  font-size: 13px;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.confirm-btn {
  width: 100%;
  height: 48px;
  margin-top: auto;
  font-size: 16px;
}

.confirm-btn:disabled { opacity: 0.55; }

.fade-enter-active,
.fade-leave-active { transition: opacity 0.18s ease; }
.fade-enter-from,
.fade-leave-to { opacity: 0; }
</style>
