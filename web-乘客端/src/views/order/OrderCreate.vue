<template>
  <div class="order-create-page">
    <!-- 真实行程地图：按订单中的上车点和目的地坐标绘制高德驾车路线。 -->
    <div id="route-map" class="trip-map" aria-label="真实行程路线地图">
      <div v-if="routeLoading" class="route-map-state"><van-loading size="24px" vertical>正在规划路线...</van-loading></div>
      <div v-else-if="routeError" class="route-map-state route-error">{{ routeError }}</div>
      <div v-if="routeReady" class="route-summary-badge">{{ estimatedDistance }}</div>
    </div>

    <!-- 顶部导航 -->
    <div class="page-header">
      <div class="back-btn" @click="goBack">
        <van-icon name="arrow-left" size="20" color="#1F2937" />
      </div>
      <span class="title">确认行程</span>
      <div style="width: 32px;"></div>
    </div>

    <!-- 地址信息卡片 -->
    <div class="address-card">
      <button type="button" class="address-item address-edit-button" @click="editAddress('pickup')">
        <div class="dot from-dot"></div>
        <div class="address-info">
          <p class="label">上车点</p>
          <p class="value">{{ orderStore.orderParams.fromAddress || '我的位置' }}</p>
        </div>
        <van-icon name="edit" size="16" color="#7C3AED" />
      </button>

      <div class="address-connector">
        <div class="line"></div>
        <div class="distance">{{ estimatedDistance }}</div>
      </div>

      <button type="button" class="address-item address-edit-button" @click="editAddress('destination')">
        <div class="dot to-dot"></div>
        <div class="address-info">
          <p class="label">目的地</p>
          <p class="value">{{ orderStore.orderParams.toAddress || '请选择目的地' }}</p>
        </div>
        <van-icon name="edit" size="16" color="#7C3AED" />
      </button>
    </div>

    <!-- 车型选择卡片 -->
    <div class="car-type-card">
      <h3 class="card-title">选择出行方式</h3>
      
      <div class="car-list">
        <div 
          v-for="car in orderStore.carTypes" 
          :key="car.type"
          class="car-item"
          :class="{ selected: car.selected }"
          @click="selectCar(car.type)"
        >
          <div class="car-icon">{{ car.icon }}</div>
          <div class="car-info">
            <p class="car-name">{{ car.name }}</p>
            <p class="car-desc">{{ car.desc }}</p>
            <p class="car-time">{{ car.time }}到达</p>
          </div>
          <div class="car-price">
            <p class="price">¥{{ car.price }}</p>
            <van-icon 
              :name="car.selected ? 'success' : 'circle'" 
              :color="car.selected ? '#7C3AED' : '#D1D5DB'"
              size="20"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 价格明细：展示原始估价、优惠券抵扣和实时应付金额。 -->
    <div v-if="priceEstimate" class="estimate-summary">
      <div><span>行程预估</span><strong>¥{{ originalCarPrice }}</strong></div>
      <div v-if="Number(priceEstimate.discountAmountCents) > 0" class="discount-row">
        <span>优惠券抵扣</span><strong>-¥{{ discountPrice }}</strong>
      </div>
      <div class="payable-row"><span>最终应付</span><strong>¥{{ currentCarPrice }}</strong></div>
    </div>

    <!-- 优惠券 -->
    <div class="coupon-card" @click="showCouponPicker = true">
      <van-icon name="coupon-o" size="20" color="#F59E0B" />
      <span class="label">优惠券</span>
      <span class="value">{{ selectedCoupon ? `-${couponAmount(selectedCoupon)}元` : (coupons.length ? `${coupons.length}张可用` : '暂无可用') }}</span>
      <van-icon name="arrow" size="14" color="#9CA3AF" />
    </div>

    <!-- 备注 -->
    <div class="remark-card" @click="showRemarkInput = true">
      <van-icon name="edit" size="20" color="#6B7280" />
      <span class="label">备注</span>
      <span class="value">{{ remark || '添加备注（选填）' }}</span>
      <van-icon name="arrow" size="14" color="#9CA3AF" />
    </div>

    <!-- 底部价格和按钮 -->
    <div class="bottom-bar safe-area-bottom">
      <div class="price-info">
        <span class="label">预估费用</span>
        <span class="price">¥{{ currentCarPrice }}</span>
      </div>
      <button 
        class="btn-primary submit-btn"
        :disabled="!canSubmit"
        @click="submitOrder"
      >
        {{ loading ? '提交中...' : estimateLoading ? '正在估价' : '立即叫车' }}
      </button>
    </div>

    <!-- 优惠券弹窗 -->
    <van-popup v-model:show="showCouponPicker" position="bottom" round>
      <div class="coupon-popup">
        <div class="popup-header">
          <span>选择优惠券</span>
          <van-icon name="cross" size="20" @click="showCouponPicker = false" />
        </div>
        <div class="coupon-list">
          <div 
            v-for="(coupon, index) in coupons" 
            :key="index"
            class="coupon-item"
            :class="{ selected: selectedCoupon?.userCouponId === coupon.userCouponId }"
            @click="selectCoupon(coupon)"
          >
            <div class="coupon-amount">¥{{ couponAmount(coupon) }}</div>
            <div class="coupon-info">
              <p>{{ coupon.name }}</p>
              <p class="expire">{{ couponExpireDate(coupon) }}到期</p>
            </div>
            <van-icon 
              :name="selectedCoupon?.userCouponId === coupon.userCouponId ? 'success' : 'circle'"
              :color="selectedCoupon?.userCouponId === coupon.userCouponId ? '#7C3AED' : '#D1D5DB'"
              size="18"
            />
          </div>
        </div>
        <button class="confirm-coupon-btn" @click="confirmCoupon">确定</button>
      </div>
    </van-popup>

    <!-- 备注输入框 -->
    <van-dialog
      v-model:show="showRemarkInput"
      title="添加备注"
      show-cancel-button
      @confirm="confirmRemark"
    >
      <textarea
        v-model="remarkTemp"
        placeholder="请输入备注信息，如需要帮助搬运行李等"
        maxlength="50"
        rows="4"
        class="remark-textarea"
      ></textarea>
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import AMapLoader from '@amap/amap-jsapi-loader'
import { showToast, showLoadingToast, closeToast } from 'vant'
import { useOrderStore } from '@/stores/order'
import { createOrder, estimateOrder, getMyCoupons } from '@/api/order'

const router = useRouter()
const orderStore = useOrderStore()

// 状态
const showCouponPicker = ref(false)
const showRemarkInput = ref(false)
const remark = ref('')
const remarkTemp = ref('')
const selectedCoupon = ref(null)

// 将后端优惠券分转换为页面展示金额。
const couponAmount = (coupon) => (Number(coupon.faceValueCents || 0) / 100).toFixed(2)
const couponExpireDate = (coupon) => coupon.expireAt ? new Date(Number(coupon.expireAt) * 1000).toLocaleDateString() : '长期有效'
const loading = ref(false)
const estimateLoading = ref(false)
const estimateError = ref('')
const priceEstimate = ref(null)
let estimateRequestId = 0

// 路线地图状态由高德 Driving 插件的真实规划结果驱动。
const routeMap = ref(null)
const drivingService = ref(null)
const routeLoading = ref(false)
const routeReady = ref(false)
const routeError = ref('')
const routeDistanceMeters = ref(0)
const routeDurationSeconds = ref(0)

const formatDistance = (meters) => {
  if (!meters) return '--'
  return meters < 1000 ? Math.round(meters) + ' 米' : (meters / 1000).toFixed(1) + ' 公里'
}

const formatDuration = (seconds) => {
  if (!seconds) return '--'
  const minutes = Math.max(1, Math.ceil(seconds / 60))
  return minutes >= 60 ? Math.floor(minutes / 60) + ' 小时 ' + (minutes % 60) + ' 分钟' : minutes + ' 分钟'
}

const estimatedDistance = computed(() => {
  if (!routeReady.value) return routeLoading.value ? '路线规划中' : '等待路线'
  return formatDistance(routeDistanceMeters.value) + ' · 预计 ' + formatDuration(routeDurationSeconds.value)
})

const hasValidCoordinate = (lng, lat) => Number.isFinite(Number(lng)) && Number.isFinite(Number(lat)) && Number(lng) !== 0 && Number(lat) !== 0

// 加载高德驾车插件，并使用首页选定的真实上车点和目的地规划路线。
const initRouteMap = async () => {
  const params = orderStore.orderParams
  if (!hasValidCoordinate(params.fromLng, params.fromLat) || !hasValidCoordinate(params.toLng, params.toLat)) {
    routeError.value = '请返回首页选择有效的上车点和目的地'
    return
  }
  const amapKey = import.meta.env.VITE_AMAP_KEY || ''
  const securityCode = import.meta.env.VITE_AMAP_SECURITY_CODE || ''
  if (!amapKey) {
    routeError.value = '未配置高德地图 Key，无法规划路线'
    return
  }
  routeLoading.value = true
  routeError.value = ''
  if (securityCode) window._AMapSecurityConfig = { securityJsCode: securityCode }
  try {
    const AMap = await AMapLoader.load({ key: amapKey, version: '2.0', plugins: ['AMap.Driving'] })
    routeMap.value = new AMap.Map('route-map', { zoom: 13, viewMode: '2D' })
    drivingService.value = new AMap.Driving({ map: routeMap.value, policy: AMap.DrivingPolicy.LEAST_TIME, hideMarkers: false, showTraffic: true, autoFitView: true })
    const origin = [Number(params.fromLng), Number(params.fromLat)]
    const destination = [Number(params.toLng), Number(params.toLat)]
    drivingService.value.search(origin, destination, (status, result) => {
      routeLoading.value = false
      if (status !== 'complete' || !result.routes?.length) {
        routeError.value = '路线规划失败，请重新选择地点后再试'
        return
      }
      const route = result.routes[0]
      routeDistanceMeters.value = Number(route.distance) || 0
      routeDurationSeconds.value = Number(route.time) || 0
      routeReady.value = true
      orderStore.setOrderParams({
        estimatedDistanceM: routeDistanceMeters.value,
        estimatedDurationS: routeDurationSeconds.value
      })
      void refreshEstimate()
    })
  } catch (error) {
    console.error('AMap driving route error:', error)
    routeLoading.value = false
    routeError.value = '高德地图加载失败，请检查网络和地图配置'
  }
}

// 将后端以“分”为单位的金额格式化为前端展示的元。
const formatMoney = (cents) => (Math.max(0, Number(cents) || 0) / 100).toFixed(2)

// 当前选中车型的后端预估价格；未返回时显示占位符，避免展示静态假价格。
const currentCarPrice = computed(() => priceEstimate.value ? formatMoney(priceEstimate.value.payableAmountCents) : '--')
const originalCarPrice = computed(() => priceEstimate.value ? formatMoney(priceEstimate.value.originalPriceCents) : '--')
const discountPrice = computed(() => priceEstimate.value ? formatMoney(priceEstimate.value.discountAmountCents) : '0.00')

// 调用乘客网关预估接口。请求序号用于丢弃快速切换车型产生的过期响应。
const refreshEstimate = async () => {
  const params = orderStore.orderParams
  if (!routeReady.value || !params.carType) return
  const requestId = ++estimateRequestId
  estimateLoading.value = true
  estimateError.value = ''
  try {
    const result = await estimateOrder({
      carType: Number(params.carType),
      fromAddress: params.fromAddress,
      fromLongitude: Number(params.fromLng),
      fromLatitude: Number(params.fromLat),
      toAddress: params.toAddress,
      toLongitude: Number(params.toLng),
      toLatitude: Number(params.toLat),
      cityCode: params.cityCode || '',
      estimatedDistanceM: Number(routeDistanceMeters.value || params.estimatedDistanceM || 0),
      estimatedDurationS: Number(routeDurationSeconds.value || params.estimatedDurationS || 0),
      userCouponId: selectedCoupon.value?.userCouponId || 0
    })
    if (requestId !== estimateRequestId) return
    priceEstimate.value = result
    const selected = orderStore.carTypes.find(car => car.type === Number(params.carType))
    if (selected) selected.price = formatMoney(result.payableAmountCents)
  } catch (error) {
    if (requestId !== estimateRequestId) return
    console.error('Estimate order price error:', error)
    priceEstimate.value = null
    estimateError.value = '价格预估失败，请稍后重试'
  } finally {
    if (requestId === estimateRequestId) estimateLoading.value = false
  }
}

// 只有路线与服务端价格均已就绪时才允许提交，防止按静态金额创建订单。
const canSubmit = computed(() => {
  const params = orderStore.orderParams
  return Boolean(params.fromAddress && params.toAddress && hasValidCoordinate(params.fromLng, params.fromLat) && hasValidCoordinate(params.toLng, params.toLat) && params.carType && routeReady.value && priceEstimate.value && !estimateLoading.value)
})

// 当前登录用户从 usersvc 查询得到的真实可用优惠券列表。
const coupons = ref([])

// 选择车型
const selectCar = (type) => {
  orderStore.selectCarType(type)
  priceEstimate.value = null
  void refreshEstimate()
}

// 返回首页并自动打开对应地址的高德搜索弹窗，修改后沿用同一份订单状态。
const editAddress = (mode) => {
  router.replace({ path: '/home', query: { edit: mode } })
}

const selectDestination = () => {
  editAddress('destination')
}

// 选择优惠券
const selectCoupon = (coupon) => {
  if (selectedCoupon.value?.userCouponId === coupon.userCouponId) {
    selectedCoupon.value = null
  } else {
    selectedCoupon.value = coupon
  }
}

// 确认优惠券
const confirmCoupon = () => {
  showCouponPicker.value = false
  orderStore.setOrderParams({
    couponId: selectedCoupon.value?.couponId || 0,
    userCouponId: selectedCoupon.value?.userCouponId || 0
  })
  priceEstimate.value = null
  void refreshEstimate()
}

// 确认备注
const confirmRemark = () => {
  remark.value = remarkTemp.value
}

// 提交订单
const submitOrder = async () => {
  if (!canSubmit.value) {
    showToast('请完善行程信息')
    return
  }

  try {
    loading.value = true
    const toast = showLoadingToast({
      message: '正在叫车...',
      forbidClick: true,
      duration: 0
    })

    // 调用创建订单接口
    const params = orderStore.orderParams
    const orderData = await createOrder({
      carType: Number(params.carType),
      fromAddress: params.fromAddress,
      fromLongitude: Number(params.fromLng),
      fromLatitude: Number(params.fromLat),
      toAddress: params.toAddress,
      toLongitude: Number(params.toLng),
      toLatitude: Number(params.toLat),
      cityCode: params.cityCode || '',
      estimatedDistanceM: Number(routeDistanceMeters.value || params.estimatedDistanceM || 0),
      estimatedDurationS: Number(routeDurationSeconds.value || params.estimatedDurationS || 0),
      userCouponId: selectedCoupon.value?.userCouponId || 0,
      remark: remark.value
    })

    closeToast()
    
    // 设置当前订单并跳转到等待接单页
    orderStore.setCurrentOrder(orderData)
    router.replace('/order/waiting')
  } catch (error) {
    console.error('Create order error:', error)
  } finally {
    loading.value = false
  }
}

const goBack = () => router.back()

// 页面进入时加载当前用户可用优惠券，并补齐车型后规划路线。
onMounted(async () => {
  try {
    const result = await getMyCoupons(1)
    coupons.value = Array.isArray(result?.list) ? result.list : []
  } catch (error) {
    console.warn('加载优惠券失败:', error)
    coupons.value = []
  }
  const selectedCar = orderStore.carTypes.find(car => car.selected)
  if (selectedCar && !orderStore.orderParams.carType) orderStore.setOrderParams({ carType: selectedCar.type })
  initRouteMap()
})

onBeforeUnmount(() => {
  drivingService.value?.clear()
  routeMap.value?.destroy()
})
</script>

<style scoped>
.order-create-page {
  min-height: 100vh;
  background: #f5f5f5;
  position: relative;
}

.trip-map {
  height: 42vh;
  min-height: 300px;
  position: relative;
  overflow: hidden;
  background: #E5E7EB;
}

.route-map-state {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.86);
  color: #6B7280;
  z-index: 5;
}

.route-error { padding: 24px; color: #DC2626; font-size: 14px; text-align: center; }

.route-summary-badge {
  position: absolute;
  top: 68px;
  left: 50%;
  padding: 8px 12px;
  border-radius: 6px;
  background: rgba(31, 41, 55, 0.88);
  color: #FFFFFF;
  font-size: 12px;
  transform: translateX(-50%);
  white-space: nowrap;
  z-index: 4;
}

.page-header {
  position: absolute;
  top: 12px;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  z-index: 10;
}


.address-edit-button {
  width: 100%;
  border: 0;
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.address-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  margin-top: 4px;
  flex-shrink: 0;
}

.from-dot {
  background: #10B981;
}

.to-dot {
  background: #F59E0B;
}


.label {
  font-size: 12px;
  color: #6B7280;
  margin-bottom: 4px;
}

 .value {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 16px;
  font-weight: 500;
  color: var(--text-primary);
}

.address-connector {
  display: flex;
  align-items: center;
  padding: 8px 0;
  margin-left: 5px;
}

.line {
  width: 24px;
  height: 2px;
  background: #D1D5DB;
}

.distance {
  font-size: 11px;
  color: #9CA3AF;
  margin-left: 8px;
}

.car-type-card {
  margin: 0 16px 12px;
  background: white;
  border-radius: 12px;
  padding: 16px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
}

.car-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.car-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border: 2px solid transparent;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.car-item.selected {
  background: #F5F3FF;
  border-color: #7C3AED;
}

.car-icon {
  font-size: 32px;
}

.car-info {
  flex: 1;
}

.car-name {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.car-desc {
  font-size: 12px;
  color: #6B7280;
  margin-bottom: 2px;
}

.car-time {
  font-size: 12px;
  color: #10B981;
}

.car-price {
  text-align: right;
}

.price {
  font-size: 18px;
  font-weight: 600;
  color: #EF4444;
  margin-bottom: 4px;
}

.estimate-panel {
  margin: 0 16px 12px;
  padding: 16px;
  background: #FFFFFF;
  border-left: 3px solid #7C3AED;
}

.estimate-heading,
.estimate-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.estimate-heading h3 {
  margin-bottom: 4px;
  font-size: 15px;
  color: #1F2937;
}

.estimate-heading p,
.estimate-meta,
.estimate-notice {
  color: #6B7280;
  font-size: 12px;
}

.estimate-total {
  color: #DC2626;
  font-size: 22px;
}

.estimate-meta {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #E5E7EB;
}

.discount-text { color: #059669; }
.estimate-error { margin-top: 10px; color: #DC2626; font-size: 12px; }
.estimate-notice { margin-top: 10px; line-height: 1.5; }
.retry-estimate { border: 0; background: transparent; color: #7C3AED; font-weight: 600; cursor: pointer; }
.coupon-card,
.remark-card {
  margin: 0 16px 12px;
  background: white;
  border-radius: 12px;
  padding: 14px 16px;
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.coupon-card .label,
.remark-card .label {
  font-size: 14px;
  color: var(--text-primary);
}

.coupon-card .value,
.remark-card  .value {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  font-size: 13px;
  color: #6B7280;
  text-align: right;
}

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  padding: 12px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
}

.price-info .label {
  font-size: 13px;
  color: #6B7280;
  margin-right: 8px;
}

.price-info .price {
  font-size: 22px;
  font-weight: 700;
  color: #EF4444;
}

.submit-btn {
  min-width: 140px;
  height: 44px;
  font-size: 15px;
}

/* 优惠券弹窗 */
.coupon-popup {
  padding: 20px;
}

.popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 17px;
  font-weight: 600;
  margin-bottom: 20px;
}

.coupon-list {
  max-height: 300px;
  overflow-y: auto;
}

.coupon-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border: 1px solid #E5E7EB;
  border-radius: 10px;
  margin-bottom: 12px;
  cursor: pointer;
}

.coupon-item.selected {
  border-color: #7C3AED;
  background: #F5F3FF;
}

.coupon-amount {
  font-size: 28px;
  font-weight: 700;
  color: #EF4444;
  min-width: 80px;
}

.coupon-info p:first-child {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

.expire {
  font-size: 12px;
  color: #9CA3AF;
  margin-top: 4px;
}

.confirm-coupon-btn {
  width: 100%;
  height: 44px;
  background: linear-gradient(135deg, #7C3AED 0%, #9333EA 100%);
  color: white;
  border: none;
  border-radius: 22px;
  font-size: 15px;
  font-weight: 500;
  margin-top: 20px;
}

.remark-textarea {
  width: 100%;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  padding: 12px;
  font-size: 14px;
  resize: none;
  outline: none;
}
.estimate-summary { margin: 0 16px 12px; padding: 12px 16px; background: #fff; border-radius: 10px; color: #6B7280; font-size: 13px; }
.estimate-summary > div { display: flex; justify-content: space-between; align-items: center; padding: 4px 0; }
.estimate-summary strong { color: #111827; font-weight: 600; }
.estimate-summary .discount-row strong { color: #059669; }
.estimate-summary .payable-row { margin-top: 6px; padding-top: 8px; border-top: 1px solid #F3F4F6; }
.estimate-summary .payable-row strong { color: #EF4444; font-size: 18px; }
</style>








