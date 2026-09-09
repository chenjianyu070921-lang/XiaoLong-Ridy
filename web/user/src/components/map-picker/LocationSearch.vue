<template>
  <div class="location-search">
    <div class="search-box">
      <van-icon name="search" size="18" color="#9CA3AF" />
      <input
        v-model="keyword"
        type="text"
        class="search-input"
        placeholder="搜索地点"
        autocomplete="off"
      />
      <van-icon v-if="keyword" name="clear" size="16" color="#9CA3AF" @click="clearKeyword" />
    </div>

    <transition name="slide-down">
      <div v-if="showPanel" class="search-results">
        <div v-if="loading" class="search-status"><van-loading size="18px">正在搜索...</van-loading></div>
        <template v-else-if="results.length">
          <button
            v-for="item in results"
            :key="item.poiId || `${item.latitude},${item.longitude}`"
            type="button"
            class="result-item"
            @click="select(item)"
          >
            <van-icon name="location-o" size="18" color="#7C3AED" />
            <span class="result-info">
              <span class="result-name">{{ item.name }}</span>
              <span class="result-address">{{ item.address }}</span>
            </span>
            <span v-if="item.distanceText" class="result-distance">{{ item.distanceText }}</span>
          </button>
        </template>
        <div v-else class="search-status">{{ message }}</div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { poiSearch } from '@/api/location'
import { createLocation, hasValidCoordinate } from '@/utils/location'

const props = defineProps({
  type: { type: String, default: 'destination' },
  city: { type: String, default: '' },
  cityCode: { type: String, default: '' },
  center: { type: Object, default: () => ({ latitude: 0, longitude: 0 }) }
})

const emit = defineEmits(['select'])

const keyword = ref('')
const results = ref([])
const loading = ref(false)
const message = ref('')
const showPanel = ref(false)
let searchTimer = null
let searchSequence = 0

function formatDistance(meters) {
  const distance = Number(meters)
  if (!Number.isFinite(distance) || distance <= 0) return ''
  return distance < 1000 ? `${Math.round(distance)}m` : `${(distance / 1000).toFixed(1)}km`
}

function clearKeyword() {
  keyword.value = ''
  results.value = []
  message.value = ''
  showPanel.value = false
  searchSequence += 1
}

function select(item) {
  emit('select', item)
  keyword.value = item.name
  results.value = []
  message.value = ''
  showPanel.value = false
}

// 关键词搜索走乘客端后端代理（locationsvc 高德 POI），city 参数用于限制搜索范围在城市内。
async function search(text) {
  const seq = ++searchSequence
  loading.value = true
  message.value = ''
  showPanel.value = true
  try {
    const result = await poiSearch({
      keyword: text,
      city: props.cityCode || props.city,
      lat: Number(props.center.latitude) || 0,
      lng: Number(props.center.longitude) || 0,
      radius: 30000,
      page: 1,
      size: 20
    })
    if (seq !== searchSequence) return
    const items = Array.isArray(result?.items) ? result.items : []
    results.value = items
      .map(item => ({ ...createLocation({ ...item, city: props.city }), distanceText: formatDistance(item.distance) }))
      .filter(item => item.name && hasValidCoordinate(item.latitude, item.longitude))
    message.value = results.value.length ? '' : '没有找到相关地点'
  } catch (error) {
    if (seq !== searchSequence) return
    console.warn('POI 搜索失败:', error)
    results.value = []
    message.value = '网络异常，请重试'
  } finally {
    if (seq === searchSequence) loading.value = false
  }
}

watch(keyword, value => {
  clearTimeout(searchTimer)
  const text = String(value || '').trim()
  if (!text) {
    searchSequence += 1
    results.value = []
    message.value = ''
    showPanel.value = false
    loading.value = false
    return
  }
  searchTimer = setTimeout(() => { void search(text) }, 350)
})
</script>

<style scoped>
.location-search {
  position: relative;
  z-index: 12;
  margin: 0 16px;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 44px;
  padding: 0 14px;
  border-radius: 22px;
  background: #FFFFFF;
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.14);
}

.search-input {
  flex: 1;
  height: 100%;
  border: none;
  outline: none;
  background: transparent;
  font-size: 15px;
  color: #1F2937;
}

.search-results {
  max-height: 42vh;
  margin-top: 8px;
  overflow-y: auto;
  border-radius: 14px;
  background: #FFFFFF;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.14);
}

.search-status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 28px 16px;
  color: #6B7280;
  font-size: 13px;
}

.result-item {
  width: 100%;
  min-height: 62px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: 0;
  border-bottom: 1px solid #F3F4F6;
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.result-item:last-child { border-bottom: 0; }

.result-info {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.result-name {
  overflow: hidden;
  color: #111827;
  font-size: 15px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-address {
  overflow: hidden;
  color: #6B7280;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-distance { flex-shrink: 0; color: #9CA3AF; font-size: 12px; }

.slide-down-enter-active,
.slide-down-leave-active { transition: opacity 0.2s ease, transform 0.2s ease; }
.slide-down-enter-from,
.slide-down-leave-to { opacity: 0; transform: translateY(-8px); }
</style>
