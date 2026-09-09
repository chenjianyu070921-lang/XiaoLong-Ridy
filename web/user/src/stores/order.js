import { defineStore } from 'pinia'
import { ref } from 'vue'
import { createLocation, LOCATION_TYPE, toOrderParams } from '@/utils/location'

export const useOrderStore = defineStore('order', () => {
  // 当前订单信息
  const currentOrder = ref(null)

  // 上车点与目的地，保存完整 Location 对象（含详细地址），供叫车页展示与价格预估使用。
  const pickupLocation = ref(null)
  const destinationLocation = ref(null)
  
  // 下单参数
  const orderParams = ref({
    carType: '',
    fromAddress: '',
    fromLat: 0,
    fromLng: 0,
    toAddress: '',
    toLat: 0,
    toLng: 0,
    couponId: '',
    cityCode: ''
  })

  // 车型列表
  const carTypes = ref([
    {
      type: 1,
      name: '特惠快车',
      icon: '🚗',
      price: '--',
      time: '~3分钟',
      desc: '经济实惠，舒适出行',
      selected: false
    },
    {
      type: 2,
      name: '快车',
      icon: '🚙',
      price: '--',
      time: '~2分钟',
      desc: '快速到达，品质出行',
      selected: true
    },
    {
      type: 3,
      name: '专车',
      icon: '🚕',
      price: '--',
      time: '~5分钟',
      desc: '高端专享，尊贵体验',
      selected: false
    }
  ])

  // 地图选点页跨页传递的结果，形如 { type, location }，由 Home / OrderCreate 在返回后消费。
  const pickedLocation = ref(null)

  // 设置当前订单
  function setCurrentOrder(order) {
    currentOrder.value = order
  }

  // 写入上车点：保存完整 Location，同时同步 orderParams 供下单/估价继续使用。
  function setPickupLocation(location) {
    const loc = createLocation(location)
    pickupLocation.value = loc
    orderParams.value = { ...orderParams.value, ...toOrderParams(loc, LOCATION_TYPE.PICKUP) }
  }

  // 写入目的地：语义同上车点。
  function setDestinationLocation(location) {
    const loc = createLocation(location)
    destinationLocation.value = loc
    orderParams.value = { ...orderParams.value, ...toOrderParams(loc, LOCATION_TYPE.DESTINATION) }
  }

  // 保存地图选点结果，等待来源页面取用。
  function setPickedLocation(value) {
    pickedLocation.value = value
  }

  // 取出并清空地图选点结果，避免同一次选点被重复消费。
  function takePickedLocation() {
    const value = pickedLocation.value
    pickedLocation.value = null
    return value
  }

  // 设置下单参数
  function setOrderParams(params) {
    orderParams.value = { ...orderParams.value, ...params }
  }

  // 重置下单参数
  function resetOrderParams() {
    orderParams.value = {
      carType: '',
      fromAddress: '',
      fromLat: 0,
      fromLng: 0,
      toAddress: '',
      toLat: 0,
      toLng: 0,
      couponId: '',
      userCouponId: 0,
      cityCode: '',
      estimatedDistanceM: 0,
      estimatedDurationS: 0
    }
  }

  // 选择车型
  function selectCarType(type) {
    carTypes.value.forEach(car => {
      car.selected = car.type === type
    })
    orderParams.value.carType = type
  }

  return {
    currentOrder,
    orderParams,
    carTypes,
    pickedLocation,
    pickupLocation,
    destinationLocation,
    setCurrentOrder,
    setOrderParams,
    resetOrderParams,
    selectCarType,
    setPickupLocation,
    setDestinationLocation,
    setPickedLocation,
    takePickedLocation
  }
})
