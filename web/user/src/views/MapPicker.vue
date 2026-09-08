<template>
  <MapPicker
    :type="pickerType"
    :city="city"
    :city-code="cityCode"
    :initial-location="initialLocation"
    @confirm="onConfirm"
    @back="goBack"
  />
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useOrderStore } from '@/stores/order'
import { createLocation, hasValidCoordinate } from '@/utils/location'
import MapPicker from '@/components/map-picker/MapPicker.vue'

const route = useRoute()
const router = useRouter()
const orderStore = useOrderStore()

// 由路由参数决定本次选点类型，非法值统一按目的地处理。
const pickerType = computed(() => (route.query.type === 'pickup' ? 'pickup' : 'destination'))
const city = computed(() => String(route.query.city || ''))
const cityCode = computed(() => String(route.query.cityCode || ''))

// 调用方可选传入初始坐标，缺省时由选点页自行定位。
const initialLocation = computed(() => {
  const location = createLocation({ latitude: route.query.lat, longitude: route.query.lng, city: city.value })
  return hasValidCoordinate(location.latitude, location.longitude) ? location : null
})

// 确认后把 Location 对象交给 order store，来源页面返回后各自消费。
function onConfirm(location) {
  orderStore.setPickedLocation({ type: pickerType.value, location })
  router.back()
}

const goBack = () => router.back()
</script>
