import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useOrderStore = defineStore('order', () => {
  // 当前订单信息
  const currentOrder = ref(null)
  
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

  // 设置当前订单
  function setCurrentOrder(order) {
    currentOrder.value = order
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
      couponId: ''
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
    setCurrentOrder,
    setOrderParams,
    resetOrderParams,
    selectCarType
  }
})
