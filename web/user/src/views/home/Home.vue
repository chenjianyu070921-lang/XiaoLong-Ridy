<template>
  <div class="home-page">
    <header class="home-header">
      <div class="brand-mark">
        <img src="/logo.png" alt="花小龙" />
        <div><strong>花小龙出行</strong><span>便捷 · 安心 · 省心</span></div>
      </div>
      <button class="city-switch" type="button" aria-label="选择城市" @click="router.push('/city/select')"><van-icon name="location" size="16" />{{ selectedCity?.name || currentCityName || '暂未定位' }}<van-icon name="arrow-down" size="12" /></button>
      <div class="header-actions">
        <van-icon name="qr" size="22" @click="showToast('扫一扫功能开发中')" />
        <van-icon name="ellipsis" size="22" @click="showToast('更多功能开发中')" />
      </div>
    </header>

    <div id="map-container" class="map-container">
      <div v-if="!mapReady" class="fallback-map">
        <div class="fallback-grid"></div>
        <div class="fallback-road road-horizontal"></div>
        <div class="fallback-road road-diagonal"></div>
        <div class="fallback-road road-vertical"></div>
        <div class="current-location-marker"><span></span></div>
      </div>
      <button v-if="locationFailed" type="button" class="relocate-btn" :disabled="locating" @click="locateUser">
        <van-icon name="replay" />{{ locating ? '定位中...' : '重新定位' }}
      </button>
      <button v-else type="button" class="current-location-btn" aria-label="回到当前定位" :disabled="locating" @click="returnToCurrentLocation"><van-icon name="aim" size="20" /></button>
    </div>

    <div class="bottom-sheet">
      <div class="sheet-heading"><div><span class="eyebrow">准备好出发了吗</span><h1>你要去哪儿？</h1></div><button type="button" class="safety-entry" @click="router.push('/safety')"><van-icon name="shield-o" size="16" />安全中心</button></div>
      <div class="pickup-bar" @click="openLocationSearch('pickup')">
        <div class="pickup-dot"></div>
        <span class="pickup-label">现在</span>
        <span class="pickup-address">{{ pickupAddress || '请选择上车点' }}</span>
        <van-icon name="arrow" size="14" color="#9CA3AF" />
      </div>
      <div class="destination-box" @click="openLocationSearch('destination')">
        <van-icon name="search" size="18" color="#9CA3AF" />
        <span class="destination-placeholder">{{ destinationAddress || '输入目的地' }}</span>
      </div>
      <div v-if="activeOrder" class="active-order-link" @click="openActiveOrder">
        <span>您有正在进行的订单</span><span class="active-order-route">{{ activeOrder.fromAddress }} → {{ activeOrder.toAddress }}</span><van-icon name="arrow" size="14" />
      </div>
      <button class="btn-primary call-car-btn" :disabled="!canCallCar || Boolean(activeOrder)" @click="callCar">{{ activeOrder ? '已有进行中的订单' : '立即叫车' }}</button>
      <div class="service-promises"><span><van-icon name="checked" />实时计价</span><span><van-icon name="shield-o" />行程保障</span><span><van-icon name="service-o" />客服在线</span></div>
    </div>

    <van-tabbar v-model="activeTab" active-color="#7C3AED" inactive-color="#6B7280">
      <van-tabbar-item icon="home-o" name="home">首页</van-tabbar-item>
      <van-tabbar-item icon="orders-o" name="orders" @click="router.push('/orders')">订单</van-tabbar-item>
      <van-tabbar-item icon="user-o" name="profile" @click="router.push('/profile')">我的</van-tabbar-item>
    </van-tabbar>

    <van-popup v-model:show="searchVisible" position="top" :style="{ height: '82%' }" round>
      <div class="search-popup">
        <div class="search-mode-title">{{ searchMode === 'pickup' ? '选择上车点' : '选择目的地' }}</div>
        <div class="search-header">
          <van-icon name="arrow-left" size="20" @click="closeLocationSearch" />
          <input v-model="keyword" type="text" class="search-input" :placeholder="searchMode === 'pickup' ? '搜索上车点' : '搜索目的地'" autofocus />
          <span class="cancel-btn" @click="closeLocationSearch">取消</span>
        </div>

        <div v-if="searchMode === 'destination' && !keyword.trim()" class="quick-addresses">
          <button v-for="item in quickDestinationAddresses" :key="item.tag" type="button" class="quick-address-button" :class="{ 'is-empty': !item.address }" @click="selectQuickDestination(item)">
            <van-icon :name="item.icon" />
            <span>{{ item.label }}</span>
          </button>
        </div>

        <div v-if="searchMode === 'destination' && !keyword.trim() && groupedRecentDestinations.length" class="search-history">
          <div class="history-title-row">
            <span class="history-title">历史目的地</span>
            <button type="button" class="history-clear" @click="clearRecentDestinations">清空历史</button>
          </div>
          <section v-for="group in groupedRecentDestinations" :key="group.cityKey" class="history-city">
            <div class="history-city-title">{{ group.cityName }}</div>
            <div class="history-list">
              <button v-for="item in group.items" :key="item.id || `${item.name}-${item.lat}-${item.lng}`" type="button" class="history-row" @click="selectSearchResult(item)">
                <van-icon name="clock-o" />
                <span class="history-info">
                  <span class="history-name">{{ item.name }}</span>
                  <span class="history-address">{{ item.address || item.name }}</span>
                </span>
                <van-icon name="cross" class="history-delete" @click.stop="removeRecentDestination(item)" />
              </button>
            </div>
          </section>
        </div>

        <div class="result-section-title">{{ keyword.trim() ? '搜索结果' : searchMode === 'pickup' ? '附近地点' : '热门目的地' }}</div>
        <div v-if="searchLoading" class="search-status"><van-loading size="24px" vertical>正在搜索地点...</van-loading></div>
        <div v-else-if="visibleSearchResults.length" class="search-results">
          <button v-for="item in visibleSearchResults" :key="item.id" type="button" class="result-item" @click="selectSearchResult(item)">
            <van-icon name="location-o" size="20" color="#7C3AED" />
            <span class="result-info"><span class="name">{{ item.name }}</span><span class="address">{{ item.address }}</span></span>
            <span v-if="item.distanceText" class="result-distance">{{ item.distanceText }}</span>
          </button>
        </div>
        <div v-else class="empty-state"><p>{{ searchMessage || (currentPoint.lat == null ? '请先允许定位后查看附近地点' : '未找到相关地点') }}</p></div>
      </div>
    </van-popup>

    <div v-if="showLoginCoupon" class="coupon-ad-mask" @click.self="closeLoginCoupon">
      <section class="coupon-ad login-coupon-ad" role="dialog" aria-modal="true" aria-label="登录优惠券广告">
        <button type="button" class="coupon-ad-close login-coupon-close" aria-label="关闭登录优惠券广告" @click="closeLoginCoupon"><van-icon name="cross" size="20" /></button>
        <button type="button" class="coupon-ad-body" @click="useLoginCoupon">
          <img src="/login-coupon-ad.png" alt="登录送优惠券" class="coupon-gift-image" />
        </button>
      </section>
    </div>

    <div v-if="showWelcomeCoupon" class="coupon-ad-mask" @click.self="showWelcomeCoupon = false">
      <section class="coupon-ad" role="dialog" aria-modal="true" aria-label="新人优惠券">
        <button type="button" class="coupon-ad-close" aria-label="关闭优惠券广告" @click="showWelcomeCoupon = false"><van-icon name="cross" size="20" /></button>
        <button type="button" class="coupon-ad-body" @click="viewWelcomeCoupons">
          <img src="/new-user-coupon-gift.png" alt="新人优惠券礼包" class="coupon-gift-image" />
          <span class="coupon-ad-action">查看优惠券</span>
        </button>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import AMapLoader from '@amap/amap-jsapi-loader'
import { getAmapConfig } from '@/config/amap'
import { showToast } from 'vant'
import { useOrderStore } from '@/stores/order'
import { listAddresses } from '@/api/address'
import { getOrders } from '@/api/order'
import { normalizeCityCode, readCurrentCity, readSelectedCity, saveCurrentCity, saveSelectedCity } from '@/data/cities'


const router = useRouter()
const orderStore = useOrderStore()
const mapInstance = ref(null)
const mapReady = ref(false)
const locationFailed = ref(false)
const locating = ref(false)
const AMapSDK = ref(null)
const placeSearch = ref(null)
const geocoder = ref(null)
const geolocation = ref(null)
const currentMarker = ref(null)
const destinationMarker = ref(null)
const pickupAddress = ref('')
const destinationAddress = ref('')
const searchVisible = ref(false)
const showWelcomeCoupon = ref(false)
const showLoginCoupon = ref(false)
const searchMode = ref('destination')
const keyword = ref('')
const searchResults = ref([])
const searchLoading = ref(false)
const searchMessage = ref('')
const activeTab = ref('home')
const activeOrder = ref(null)
const commonAddresses = ref([])
const currentPoint = ref({ lat: null, lng: null, address: '' })
// 当前定位所属行政区编码，优先使用 adcode 精确限制 POI 搜索范围。
const selectedCity = ref(localStorage.getItem('passenger-manual-city') === '1' ? readSelectedCity() : null)
// 历史版本可能保存了"当前位置"占位值，不能把它当作真实城市展示。
if (selectedCity.value?.name === '当前位置') {
  selectedCity.value = null
  localStorage.removeItem('passenger-selected-city')
  localStorage.removeItem('passenger-manual-city')
}
const currentLocationCity = readCurrentCity()
// 先显示最近一次定位城市，新的定位完成后再用实时结果覆盖。
const currentCityName = ref(selectedCity.value?.name || currentLocationCity?.name || '')
const currentCityCode = ref(normalizeCityCode(selectedCity.value?.adcode || currentLocationCity?.adcode || ''))

// 根据当前城市编码重建 POI 搜索实例，确保 city 和 citylimit 同时生效。
function configurePlaceSearch(cityCode = '') {
  if (!AMapSDK.value) return
  placeSearch.value = new AMapSDK.value.PlaceSearch({
    pageSize: 50,
    pageIndex: 1,
    extensions: 'base',
    city: cityCode || undefined,
    citylimit: Boolean(cityCode)
  })
}
let searchTimer = null
let searchSequence = 0

const welcomeCouponKey = 'passenger-welcome-coupon-claimed'
const newUserPendingGiftKey = 'passenger-new-user-pending-gift'
const loginCouponPendingKey = 'passenger-login-coupon-pending'
const recentDestinationKey = 'passenger-recent-destinations'
const recentDestinations = ref(loadRecentDestinations())
const canCallCar = computed(() => Boolean(pickupAddress.value && destinationAddress.value))
// groupedRecentDestinations 按城市拆分历史目的地，每个城市只展示最近 5 条。
const groupedRecentDestinations = computed(() => {
  const groups = []
  const groupMap = new Map()
  for (const item of recentDestinations.value) {
    const cityKey = item.cityCode || item.cityName || 'unknown'
    const cityName = item.cityName || '未知城市'
    if (!groupMap.has(cityKey)) {
      const group = { cityKey, cityName, items: [] }
      groupMap.set(cityKey, group)
      groups.push(group)
    }
    const group = groupMap.get(cityKey)
    if (group.items.length < 5) group.items.push(item)
  }
  return groups
})
// visibleSearchResults 控制选址列表展示数量，热门目的地最多保留 20 条，搜索和上车点列表保持完整。
const visibleSearchResults = computed(() => {
  if (searchMode.value === 'destination' && !keyword.value.trim()) return searchResults.value.slice(0, 20)
  return searchResults.value
})
const quickDestinationAddresses = computed(() => [
  { tag: 'home', label: '家', icon: 'wap-home-o', address: findCommonAddress('home') },
  { tag: 'work', label: '公司', icon: 'office-o', address: findCommonAddress('work') }
])

// findCommonAddress 从常用地址列表中取出指定标签的地址，用于终点快捷选择。
function findCommonAddress(tag) {
  return commonAddresses.value.find(item => item.tag === tag && Number.isFinite(item.lat) && Number.isFinite(item.lng)) || null
}

// 从本地缓存读取最近目的地，异常数据会被过滤，避免影响地图搜索。
function loadRecentDestinations() {
  try {
    const list = JSON.parse(localStorage.getItem(recentDestinationKey) || '[]')
    return Array.isArray(list) ? list.map(normalizeRecentDestination).filter(Boolean).slice(0, 10) : []
  } catch (error) {
    console.warn('读取历史目的地失败:', error)
    return []
  }
}

// normalizeRecentDestination 兼容旧缓存，并补齐城市分类字段，避免不同城市历史混到同一组。
function normalizeRecentDestination(item) {
  if (!item?.name || !Number.isFinite(Number(item.lat)) || !Number.isFinite(Number(item.lng))) return null
  const cityCode = item.cityCode || currentCityCode.value || ''
  const cityName = item.cityName || currentCityName.value || selectedCity.value?.name || '未知城市'
  return { ...item, lat: Number(item.lat), lng: Number(item.lng), cityCode, cityName }
}

function rememberDestination(item) {
  const saved = {
    id: item.id || '',
    name: item.name,
    address: item.address || '',
    lat: item.lat,
    lng: item.lng,
    cityCode: item.cityCode || currentCityCode.value || '',
    cityName: item.cityName || currentCityName.value || selectedCity.value?.name || '未知城市'
  }
  const rest = recentDestinations.value.filter(entry => entry.id !== saved.id && (entry.lat !== saved.lat || entry.lng !== saved.lng))
  recentDestinations.value = [saved, ...rest].slice(0, 10)
  localStorage.setItem(recentDestinationKey, JSON.stringify(recentDestinations.value))
}

// removeRecentDestination 删除单条历史目的地，保留其他历史记录继续用于快捷选址。
function removeRecentDestination(item) {
  recentDestinations.value = recentDestinations.value.filter(entry => entry.id !== item.id && (entry.lat !== item.lat || entry.lng !== item.lng)).slice(0, 10)
  localStorage.setItem(recentDestinationKey, JSON.stringify(recentDestinations.value))
}

// clearRecentDestinations 一键清空本机保存的全部历史目的地。
function clearRecentDestinations() {
  recentDestinations.value = []
  localStorage.removeItem(recentDestinationKey)
}

// 初始化高德地图及定位、地点搜索和逆地理编码服务。
async function initMap() {
  const { key, securityCode } = getAmapConfig()
  if (!key) {
    locationFailed.value = true
    showToast('未配置高德地图 Key')
    return
  }
  try {
    if (securityCode) window._AMapSecurityConfig = { securityJsCode: securityCode }
    const AMap = await AMapLoader.load({ key, version: '2.0', plugins: ['AMap.Geolocation', 'AMap.PlaceSearch', 'AMap.CitySearch', 'AMap.Geocoder', 'AMap.InputTips', 'AMap.AutoComplete'] })
    AMapSDK.value = AMap
    // 未获取到用户位置时不指定重庆等固定城市，保持地图为通用初始视图，等待定位结果后再居中。
    mapInstance.value = new AMap.Map('map-container', { zoom: 4, viewMode: '2D' })
    configurePlaceSearch()
    geocoder.value = new AMap.Geocoder({ radius: 1000, extensions: 'base' })
    geolocation.value = new AMap.Geolocation({ enableHighAccuracy: true, timeout: 10000, zoomToAccuracy: true, position: 'RB', offset: [16, 108] })
    mapInstance.value.addControl(geolocation.value)
    mapReady.value = true
    await locateUser()
    // 用户手动选择城市后，定位完成也要立即切换到所选城市。
    if (selectedCity.value) await applySelectedCity(selectedCity.value)
  } catch (error) {
    console.error('地图初始化失败:', error)
    locationFailed.value = true
    showToast('地图加载失败，请检查地图配置')
  }
}

function formatDetailedAddress(addressComponent = {}, fallback = '') {
  // 将高德地址组件拼成可直接展示和下单保存的完整地址。
  const region = [addressComponent.province, addressComponent.city, addressComponent.district].filter(Boolean).join('')
  const detail = [addressComponent.township, addressComponent.street, addressComponent.streetNumber].filter(Boolean).join('')
  return [region, detail].filter(Boolean).join('') || fallback || '当前位置'
}

function reverseGeocode(lng, lat) {
  return new Promise(resolve => {
    if (!geocoder.value) return resolve({ address: '当前位置', cityName: '', adcode: '' })
    geocoder.value.getAddress([Number(lng), Number(lat)], (status, result) => {
      if (status !== 'complete' || !result?.regeocode) return resolve({ address: '当前位置', cityName: '', adcode: '' })
      resolve({
        address: result.regeocode.formattedAddress || formatDetailedAddress(result.regeocode.addressComponent),
        cityName: result.regeocode.addressComponent?.city || result.regeocode.addressComponent?.province || '',
        adcode: result.regeocode.addressComponent?.adcode || ''
      })
    })
  })
}

// 通过当前坐标反查 adcode，并立即刷新 POI 搜索城市范围。
async function resolveCityCode(lng, lat) {
  const result = await reverseGeocode(lng, lat)
  if (result.adcode) {
    // 逆地理返回的可能是区县编码，必须先归一化到地级市再限制 POI 搜索。
    result.adcode = normalizeCityCode(result.adcode)
    currentCityCode.value = result.adcode
    configurePlaceSearch(result.adcode)
  }
  return result
}

// loadCommonAddresses 加载家/公司常用地址，接口失败时不阻断正常 POI 搜索。
async function loadCommonAddresses() {
  try {
    const result = await listAddresses()
    const list = Array.isArray(result?.list) ? result.list : []
    commonAddresses.value = list.map(item => ({
      id: item.id,
      tag: item.tag,
      name: item.tag === 'home' ? '家' : item.tag === 'work' ? '公司' : item.address,
      address: item.address,
      lat: Number(item.latitude),
      lng: Number(item.longitude)
    })).filter(item => item.address && Number.isFinite(item.lat) && Number.isFinite(item.lng))
  } catch (error) {
    commonAddresses.value = []
    console.warn('查询常用地址失败:', error)
  }
}
function updateCurrentMarker(lng, lat) {
  // 地图异步初始化或页面销毁期间不执行覆盖物操作，避免调用未就绪实例的 add。
  if (!mapReady.value || !mapInstance.value || typeof mapInstance.value.add !== 'function' || !AMapSDK.value) return
  const position = [Number(lng), Number(lat)]
  if (currentMarker.value) currentMarker.value.setPosition(position)
  else {
    currentMarker.value = new AMapSDK.value.Marker({ position, title: '当前位置', anchor: 'center', zIndex: 120 })
    mapInstance.value.add(currentMarker.value)
  }
  mapInstance.value.setZoomAndCenter(16, position)
}

async function applyCurrentLocation({ lng, lat, address, name, addressComponent }, useAsPickup = true) {
  const normalizedLng = Number(lng)
  const normalizedLat = Number(lat)
  if (!Number.isFinite(normalizedLng) || !Number.isFinite(normalizedLat)) return markLocationFailure('定位坐标异常，请重新定位')
  // 定位结果可能只有坐标或占位文字，优先通过逆地理获取详细地址。
  let text = address || ''
  let resolvedAdcode = normalizeCityCode(addressComponent?.adcode || '')
  let resolvedCityName = addressComponent?.city || addressComponent?.province || ''
  // 高德精确定位常常只返回 formattedAddress，未同时返回 city/adcode。
  // 即使详细地址已经存在，只要城市名称或编码缺失，也必须逆地理补齐，不能直接显示“当前城市”。
  const shouldResolveCity = !resolvedCityName || !resolvedAdcode
  if (!text || ['当前位置', '我的位置', '当前城市'].includes(text) || shouldResolveCity) {
    const reverseResult = await reverseGeocode(normalizedLng, normalizedLat)
    text = text && !['当前位置', '我的位置', '当前城市'].includes(text) ? text : reverseResult.address
    resolvedAdcode = resolvedAdcode || normalizeCityCode(reverseResult.adcode)
    resolvedCityName = resolvedCityName || reverseResult.cityName
  }
  locationFailed.value = false
  currentPoint.value = { lng: normalizedLng, lat: normalizedLat, address: text }
  // 定位结果只作为默认城市来源；手动选择城市后，不能再被设备定位覆盖。
  if (!selectedCity.value) {
    if (resolvedAdcode) {
      currentCityCode.value = resolvedAdcode
      currentCityName.value = resolvedCityName || '当前城市'
      // 保存真实定位城市，供城市选择页展示，并让后续页面复用同一城市上下文。
      saveCurrentCity({ name: currentCityName.value, adcode: currentCityCode.value })
      configurePlaceSearch(currentCityCode.value)
    } else {
      // 浏览器定位或高德定位未返回 adcode 时，通过坐标逆地理编码补齐行政区编码。
      const cityResult = await resolveCityCode(normalizedLng, normalizedLat)
      currentCityName.value = cityResult.cityName || currentCityName.value
      if (cityResult.cityName && cityResult.adcode) {
        saveCurrentCity({ name: cityResult.cityName, adcode: cityResult.adcode })
      }
    }
  } else {
    currentCityCode.value = normalizeCityCode(selectedCity.value.adcode)
    currentCityName.value = selectedCity.value.name
    configurePlaceSearch(currentCityCode.value)
  }
  // 定位成功后默认写入上车点，用户也可以在选址弹层中手动改成其他地点。
  if (useAsPickup) {
    const pickupName = name || text
    pickupAddress.value = pickupName
    orderStore.setOrderParams({ fromLng: normalizedLng, fromLat: normalizedLat, fromAddress: pickupName, cityCode: currentCityCode.value })
  }
  updateCurrentMarker(normalizedLng, normalizedLat)
  if (searchVisible.value && !keyword.value.trim()) searchNearby()
}

// 将手动选择的城市切换为打车页面的工作城市，并把地图移动到该城市中心。
async function applySelectedCity(city) {
  if (!city?.name || !city?.adcode || !geocoder.value || !mapInstance.value) return
  currentCityName.value = city.name
  currentCityCode.value = normalizeCityCode(city.adcode)
  configurePlaceSearch(currentCityCode.value)
  await new Promise(resolve => {
    geocoder.value.getLocation(city.name, (status, result) => {
      const location = result?.geocodes?.[0]?.location
      if (status === 'complete' && location) {
        const lng = typeof location.getLng === 'function' ? location.getLng() : Number(location.lng)
        const lat = typeof location.getLat === 'function' ? location.getLat() : Number(location.lat)
        if (Number.isFinite(lng) && Number.isFinite(lat)) {
          currentPoint.value = { lng, lat, address: city.name }
          pickupAddress.value = city.name
          orderStore.setOrderParams({ fromLng: lng, fromLat: lat, fromAddress: city.name, cityCode: city.adcode })
          updateCurrentMarker(lng, lat)
        }
      }
      resolve()
    })
  })
}

function locateByAMap() {
  return new Promise((resolve, reject) => geolocation.value?.getCurrentPosition((status, result) => status === 'complete' && result?.position ? resolve({ lng: result.position.lng, lat: result.position.lat, address: result.formattedAddress || formatDetailedAddress(result.addressComponent, '当前位置'), addressComponent: result.addressComponent, adcode: result.addressComponent?.adcode || '' }) : reject(result || new Error('高德定位失败'))))
}

function locateByBrowser() {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) return reject(new Error('当前浏览器不支持定位'))
    navigator.geolocation.getCurrentPosition(async result => {
      const location = await reverseGeocode(result.coords.longitude, result.coords.latitude)
      resolve({ lng: result.coords.longitude, lat: result.coords.latitude, address: location.address, addressComponent: { adcode: location.adcode } })
    }, reject, { enableHighAccuracy: true, timeout: 12000, maximumAge: 15000 })
  })
}

function locateByCity() {
  return new Promise((resolve, reject) => {
    if (!AMapSDK.value?.CitySearch) return reject(new Error('城市定位服务未初始化'))
    new AMapSDK.value.CitySearch().getCity((status, result) => status === 'complete' && result?.center ? resolve({ lng: Array.isArray(result.center) ? result.center[0] : result.center.lng, lat: Array.isArray(result.center) ? result.center[1] : result.center.lat, address: result.city || '当前城市' }) : reject(result || new Error('城市定位失败')))
  })
}

function markLocationFailure(message = '定位失败，请检查定位权限或点击重新定位') {
  locationFailed.value = true
  pickupAddress.value = '定位失败，请重新定位'
  showToast(message)
}

// 定位按高德精确定位、浏览器定位、城市中心三级降级，保证页面可继续使用。
async function locateUser() {
  if (locating.value) return
  locating.value = true
  locationFailed.value = false
  try {
    let result
    try { result = await locateByAMap() } catch (firstError) {
      console.warn('高德精确定位失败，尝试浏览器定位:', firstError)
      try { result = await locateByBrowser() } catch (secondError) {
        console.warn('浏览器定位失败，尝试城市定位:', secondError)
        result = await locateByCity()
      }
    }
    await applyCurrentLocation(result)
  } catch (error) {
    console.error('所有定位方式均失败:', error)
    markLocationFailure()
  } finally { locating.value = false }
}

// 回到设备当前定位，并清除手动城市选择，让打车页恢复真实定位城市。
async function returnToCurrentLocation() {
  if (locating.value) return
  selectedCity.value = null
  currentCityName.value = ''
  currentCityCode.value = ''
  localStorage.removeItem('passenger-selected-city')
  localStorage.removeItem('passenger-manual-city')
  configurePlaceSearch()
  await locateUser()
}

function normalizePOI(poi) {
  if (!poi?.location) return null
  const lng = typeof poi.location.getLng === 'function' ? poi.location.getLng() : Number(poi.location.lng)
  const lat = typeof poi.location.getLat === 'function' ? poi.location.getLat() : Number(poi.location.lat)
  if (!Number.isFinite(lng) || !Number.isFinite(lat)) return null
  const prefix = [poi.pname, poi.cityname, poi.adname].filter(Boolean).join('')
  const distance = Number(poi.distance)
  // 同时保留 POI 名称和详细地址：名称用于订单展示，详细地址用于搜索结果辅助识别。
  const detailedAddress = [prefix, poi.address].filter(Boolean).join('') || poi.name || '暂无详细地址'
  const rawCityName = Array.isArray(poi.cityname) ? '' : poi.cityname
  const rawAdcode = Array.isArray(poi.adcode) ? '' : poi.adcode
  return { id: poi.id || `${lng},${lat}`, name: poi.name || '未命名地点', address: detailedAddress, displayAddress: detailedAddress, lng, lat, cityName: rawCityName || poi.pname || '', cityCode: normalizeCityCode(rawAdcode || ''), distanceText: Number.isFinite(distance) ? distance < 1000 ? `${Math.round(distance)}m` : `${(distance / 1000).toFixed(1)}km` : '' }
}

// resolveDestinationCity 补齐目的地所属城市，历史记录必须按目的地城市归类。
async function resolveDestinationCity(item) {
  if (item.cityName || item.cityCode) return item
  const result = await reverseGeocode(item.lng, item.lat)
  return {
    ...item,
    cityName: result.cityName || '未知城市',
    cityCode: normalizeCityCode(result.adcode || '')
  }
}

function handleSearchResult(sequence, status, result) {
  if (sequence !== searchSequence) return
  searchLoading.value = false
  if (status !== 'complete') {
    searchResults.value = []
    searchMessage.value = status === 'no_data' ? '未找到相关地点' : '地点搜索失败，请稍后重试'
    return
  }
  searchResults.value = (result.poiList?.pois || []).map(normalizePOI).filter(Boolean)
}

// 目的地关键词搜索改为后端代理：passenger-api → locationsvc.POISearch → AMap(后端 key)。
// 不再走 AMap JS InputTips/PlaceSearch，避开浏览器侧 key 类型不匹配导致全部失败的问题。
// 邻近浏览 searchNearby依赖 AMap JS（地图附近特殊浏览依赖 JS key，后续单独处理）。
async function searchByKeyword(value) {
  const sequence = ++searchSequence
  searchLoading.value = true
  searchMessage.value = ''
  const cityCode = currentCityCode.value || selectedCity.value?.adcode || ''
  const lat = Number(currentPoint.value.lat) || 0
  const lng = Number(currentPoint.value.lng) || 0
  try {
    const { poiSearch } = await import('@/api/location')
    const data = await poiSearch({
      keyword: value,
      city: cityCode,
      lat,
      lng,
      size: 20
    })
    if (sequence !== searchSequence) return
    const rawItems = Array.isArray(data?.items) ? data.items : []
    const items = rawItems.map(p => ({
      id: `${p.lng},${p.lat}`,
      name: p.name || '未命名地点',
      address: p.address || p.name || '',
      displayAddress: p.address || p.name || '',
      lng: p.lng,
      lat: p.lat,
      cityName: '',
      cityCode: '',
      distanceText: ''
    }))
    searchResults.value = items
    searchMessage.value = items.length ? '' : '未找到相关地点，请换个关键词试试'
  } catch (e) {
    if (sequence !== searchSequence) return
    console.warn('POI 搜索失败:', e)
    searchResults.value = []
    searchMessage.value = '地点搜索失败，请稍后重试'
  } finally {
    if (sequence === searchSequence) searchLoading.value = false
  }
}
function searchNearby() {
  if (!placeSearch.value || currentPoint.value.lng == null || currentPoint.value.lat == null) {
    searchResults.value = []
    searchLoading.value = false
    return
  }
  const sequence = ++searchSequence
  searchLoading.value = true
  searchMessage.value = ''
  placeSearch.value.setType(searchMode.value === 'pickup' ? '' : '购物服务|交通设施服务|医疗保健服务|风景名胜')
  placeSearch.value.searchNearBy('', [currentPoint.value.lng, currentPoint.value.lat], searchMode.value === 'pickup' ? 3000 : 10000, (status, result) => handleSearchResult(sequence, status, result))
}

watch(keyword, value => {
  clearTimeout(searchTimer)
  searchMessage.value = ''
  const text = value.trim()
  if (!text) { searchSequence += 1; searchNearby(); return }
  searchTimer = setTimeout(() => searchByKeyword(text), 350)
})

function openLocationSearch(mode) {
  searchMode.value = mode
  keyword.value = ''
  searchResults.value = []
  searchMessage.value = ''
  searchVisible.value = true
  searchNearby()
}

function closeLocationSearch() {
  searchVisible.value = false
  keyword.value = ''
  searchResults.value = []
  searchSequence += 1
}

async function selectSearchResult(item) {
  if (searchMode.value === 'pickup') {
    await applyCurrentLocation({ lng: item.lng, lat: item.lat, name: item.name, address: item.displayAddress || item.address || item.name, addressComponent: { adcode: currentCityCode.value } })
    closeLocationSearch()
    return
  }
  const resolvedDestination = await resolveDestinationCity(item)
  destinationAddress.value = resolvedDestination.name
  orderStore.setOrderParams({ toAddress: resolvedDestination.name, toLat: resolvedDestination.lat, toLng: resolvedDestination.lng, cityCode: resolvedDestination.cityCode || currentCityCode.value })
  rememberDestination(resolvedDestination)
  closeLocationSearch()
  if (mapInstance.value) {
    const position = [resolvedDestination.lng, resolvedDestination.lat]
    if (!destinationMarker.value && AMapSDK.value) {
      destinationMarker.value = new AMapSDK.value.Marker({ position })
      if (typeof mapInstance.value.add === 'function') mapInstance.value.add(destinationMarker.value)
    } else destinationMarker.value?.setPosition(position)
    mapInstance.value.setCenter(position)
  }
}

// selectQuickDestination 将家/公司常用地址复用为目的地，保证下单参数和地图标记走统一入口。
async function selectQuickDestination(item) {
  if (!item.address) {
    closeLocationSearch()
    router.push({ path: '/addresses', query: { mode: 'create', tag: item.tag } })
    return
  }
  await selectSearchResult({
    id: item.address.id || item.tag,
    name: item.address.address,
    address: item.label,
    displayAddress: item.address.address,
    lat: item.address.lat,
    lng: item.address.lng
  })
}

function callCar() {
  if (!canCallCar.value) return
  const selectedCar = orderStore.carTypes.find(car => car.selected)
  orderStore.setOrderParams({ carType: orderStore.orderParams.carType || selectedCar?.type || 2, fromAddress: pickupAddress.value })
  router.push('/order/create')
}

function closeLoginCoupon() {
  localStorage.removeItem(loginCouponPendingKey)
  showLoginCoupon.value = false
}

function useLoginCoupon() {
  closeLoginCoupon()
  router.push('/order/create')
}

function viewWelcomeCoupons() {
  // 注册时优惠券已经由 usersvc 自动写入 user_coupon，这里只负责跳转查看，避免重复领取。
  localStorage.setItem(welcomeCouponKey, '1')
  localStorage.removeItem(newUserPendingGiftKey)
  showWelcomeCoupon.value = false
  router.push('/coupons')
}

async function loadActiveOrder() {
  try {
    const result = await getOrders({ page: 1, pageSize: 50, status: 0 })
    const list = Array.isArray(result?.list) ? result.list : []
    activeOrder.value = list.find(item => [1, 2, 3].includes(Number(item.status))) || null
  } catch (error) {
    console.warn('查询进行中订单失败:', error)
  }
}

function openActiveOrder() {
  if (!activeOrder.value?.orderId) return
  router.push(/orders/)
}

onMounted(() => {
  initMap()
  loadCommonAddresses()
  loadActiveOrder()
  if (localStorage.getItem(newUserPendingGiftKey) === '1' && !localStorage.getItem(welcomeCouponKey)) {
    showWelcomeCoupon.value = true
  } else if (localStorage.getItem(loginCouponPendingKey) === '1') {
    showLoginCoupon.value = true
  }
})

onBeforeUnmount(() => {
  clearTimeout(searchTimer)
  searchSequence += 1
  mapInstance.value?.destroy()
})
</script>

<style scoped>
/* 首页样式从最近一次可用构建产物恢复，保持地图和底部操作区原有布局。 */
.home-page{min-height:100vh;background:#eef2f7;position:relative}.city-switch{display:inline-flex;align-items:center;gap:4px;min-width:0;margin-left:2px;padding:8px 10px;border:0;border-radius:18px;background:#ffffffe6;color:#374151;font-size:13px;box-shadow:0 2px 8px #0f172a1f}.brand-mark{display:flex;align-items:center;gap:9px;min-width:0}.brand-mark img{width:34px;height:34px;border-radius:11px;box-shadow:0 4px 12px #7c3aed35}.brand-mark div{display:flex;flex-direction:column;gap:2px}.brand-mark strong{font-size:14px}.brand-mark span{font-size:10px;color:#64748b}.sheet-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:12px;margin-bottom:12px}.eyebrow{display:block;margin-bottom:4px;color:#7c3aed;font-size:11px;font-weight:700;letter-spacing:1px}.sheet-heading h1{font-size:23px;line-height:1.2;font-weight:800;color:#111827}.safety-entry{display:inline-flex;align-items:center;gap:5px;padding:8px 10px;border:1px solid #ede9fe;border-radius:10px;background:#faf5ff;color:#6d28d9;font-size:12px;white-space:nowrap}.service-promises{display:flex;justify-content:space-between;padding:12px 2px 0;color:#94a3b8;font-size:11px}.service-promises span{display:inline-flex;align-items:center;gap:4px}.service-promises .van-icon{color:#10b981}.map-container{position:fixed;top:0;left:0;width:100%;height:100vh;background:#dde7f0;overflow:hidden}.home-header{position:fixed;top:0;left:0;right:0;z-index:30;display:flex;justify-content:flex-start;align-items:center;gap:8px;padding:calc(env(safe-area-inset-top) + 12px) 16px 12px;background:linear-gradient(180deg,#ffffffeb,#fff0);color:#1f2937}.home-header .brand-mark{flex-shrink:1}.home-header .city-switch{flex-shrink:0}.header-actions{display:flex;align-items:center;gap:16px;margin-left:auto}.header-actions .van-icon{width:38px;height:38px;display:flex;align-items:center;justify-content:center;border-radius:50%;background:#ffffffe6;box-shadow:0 2px 8px #0f172a1f;cursor:pointer}.fallback-map{position:absolute;top:0;right:0;bottom:0;left:0;background:radial-gradient(circle at 28% 26%,rgba(16,185,129,.18),transparent 18%),radial-gradient(circle at 74% 50%,rgba(245,158,11,.16),transparent 20%),linear-gradient(135deg,#e8f3ff,#f8fafc 54%,#ecfdf5)}.fallback-grid{position:absolute;top:0;right:0;bottom:0;left:0;background-image:linear-gradient(rgba(148,163,184,.16) 1px,transparent 1px),linear-gradient(90deg,rgba(148,163,184,.16) 1px,transparent 1px);background-size:36px 36px}.fallback-road{position:absolute;background:#ffffffc7;box-shadow:0 0 0 1px #cbd5e1a3}.road-horizontal{width:120%;height:34px;left:-10%;top:42%;transform:rotate(-8deg)}.road-diagonal{width:34px;height:120%;left:58%;top:-12%;transform:rotate(24deg)}.road-vertical{width:28px;height:80%;left:24%;top:6%;transform:rotate(-14deg)}.current-location-marker{position:absolute;left:50%;top:36%;width:28px;height:28px;display:flex;align-items:center;justify-content:center;background:#7c3aed29;border-radius:50%;transform:translate(-50%,-50%)}.current-location-marker span{width:12px;height:12px;background:#7c3aed;border:3px solid #FFFFFF;border-radius:50%;box-shadow:0 4px 12px #7c3aed52}.relocate-btn{position:absolute;left:50%;top:66%;display:inline-flex;align-items:center;gap:6px;padding:8px 14px;transform:translate(-50%,-50%);border:0;border-radius:22px;background:#ffffffeb;color:#dc2626;font-size:14px;box-shadow:0 6px 18px #0f172a2e;cursor:pointer}.current-location-btn{position:absolute;right:16px;top:calc(60% - 10px);bottom:auto;z-index:26;width:44px;height:44px;display:flex;align-items:center;justify-content:center;border:0;border-radius:50%;background:#fffffff2;color:#2563eb;box-shadow:0 4px 14px #0f172a2e;cursor:pointer}.current-location-btn:disabled{color:#94a3b8;cursor:not-allowed;opacity:.75}.relocate-btn:disabled{color:#6b7280;cursor:not-allowed;opacity:.78}.bottom-sheet{position:fixed;right:0;bottom:calc(50px + env(safe-area-inset-bottom));left:0;z-index:25;padding:18px 16px 14px;border-radius:18px 18px 0 0;background:#fff;box-shadow:0 -6px 24px #0f172a24}.pickup-bar{display:flex;align-items:center;gap:10px;padding:10px 0}.pickup-dot{width:10px;height:10px;border-radius:50%;background:#10b981;flex-shrink:0}.pickup-label{color:#6b7280;font-size:13px;flex-shrink:0}.pickup-address{flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--text-primary);font-size:15px;font-weight:500}.destination-box{display:flex;align-items:center;gap:8px;margin:6px 0 14px;padding:14px 16px;border-radius:12px;background:#f3f4f6;cursor:text}.destination-placeholder{color:#6b7280;font-size:15px}.search-history{padding-top:16px}.history-title{margin-bottom:8px;color:#6b7280;font-size:13px}.history-item{width:100%;min-height:58px;display:flex;align-items:center;gap:12px;padding:10px 0;border:0;border-bottom:1px solid #F3F4F6;background:transparent;text-align:left;cursor:pointer}.history-info{min-width:0;display:flex;flex:1;flex-direction:column;gap:4px}.history-name{color:var(--text-primary);font-size:15px}.history-address{overflow:hidden;color:#6b7280;font-size:13px;text-overflow:ellipsis;white-space:nowrap}.active-order-link { display: flex; align-items: center; gap: 8px; margin: 12px 0 8px; padding: 12px 14px; border: 1px solid #DDD6FE; border-radius: 10px; background: #F5F3FF; color: #6D28D9; font-size: 13px; cursor: pointer; }
.active-order-link .active-order-route { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #4C1D95; }
.call-car-btn{width:100%;height:48px;font-size:16px;margin-top:16px}.search-popup{padding:16px;height:100%;overflow-y:auto}.search-mode-title{margin-bottom:14px;color:#111827;font-size:18px;font-weight:600;text-align:center}.result-section-title{padding:18px 0 8px;color:#374151;font-size:15px;font-weight:600}.history-chips{display:flex;gap:8px;overflow-x:auto}.history-chip{min-width:0;display:flex;align-items:center;gap:6px;padding:8px 10px 8px 12px;border:1px solid #E5E7EB;border-radius:6px;background:#fff;color:#374151;white-space:nowrap}.history-chip span{max-width:112px;overflow:hidden;text-overflow:ellipsis}.history-delete{flex-shrink:0;margin-left:2px;padding:2px;color:#9CA3AF}.quick-addresses{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:10px;padding-top:14px}.quick-address-button{height:42px;display:flex;align-items:center;justify-content:center;gap:6px;border:1px solid #E5E7EB;border-radius:6px;background:#fff;color:#374151;font-size:14px;font-weight:500}.quick-address-button.is-empty{color:#7C3AED;background:#F5F3FF;border-color:#DDD6FE}.search-header{display:flex;align-items:center;gap:12px;padding-bottom:16px;border-bottom:1px solid #F3F4F6}.search-input{flex:1;height:40px;border:none;outline:none;font-size:15px;background:transparent}.cancel-btn{color:var(--primary-color);font-size:14px;cursor:pointer;white-space:nowrap}.search-results{padding-top:16px}.search-status{display:flex;justify-content:center;padding:56px 0;color:#6b7280}.result-item{width:100%;min-height:68px;display:flex;align-items:center;gap:12px;padding:12px 0;border:0;border-bottom:1px solid #F3F4F6;background:transparent;text-align:left;cursor:pointer}.result-info{min-width:0;display:flex;flex:1;flex-direction:column}.result-distance{flex-shrink:0;color:#9ca3af;font-size:12px}.result-info .name{font-size:15px;font-weight:500;color:var(--text-primary);margin-bottom:4px}.result-info .address{font-size:13px;color:#6b7280}.empty-state{text-align:center;padding:60px 0;color:#9ca3af}.coupon-ad-mask{position:fixed;top:0;right:0;bottom:0;left:0;z-index:30;display:flex;align-items:center;justify-content:center;padding:24px;background:#0f172a7a}.coupon-ad{position:relative;width:min(330px,100%);overflow:hidden;border-radius:16px;background:linear-gradient(145deg,#ef4444,#f97316);box-shadow:0 18px 42px #7f1d1d59}.coupon-ad-close{position:absolute;top:10px;right:10px;z-index:1;width:36px;height:36px;border:0;border-radius:50%;color:#fff;background:#7f1d1d47}.login-coupon-ad{width:min(390px,100%);background:#fff}.login-coupon-ad .coupon-ad-body{padding:0}.login-coupon-ad .coupon-gift-image{max-height:calc(100vh - 72px);width:100%;object-fit:contain}.login-coupon-close{top:auto;right:50%;bottom:-52px;transform:translateX(50%);background:#ffffff;color:#6b21a8;box-shadow:0 2px 10px #0003}.coupon-ad-body{display:flex;width:100%;min-height:300px;align-items:center;flex-direction:column;justify-content:center;gap:8px;padding:42px 28px 28px;border:0;color:#fff;background:transparent}.coupon-gift-image{display:block;width:100%;max-height:420px;object-fit:contain}.coupon-ad-kicker{font-size:14px;opacity:.9}.coupon-ad-title{font-size:26px;line-height:1.25}.coupon-ad-subtitle{font-size:15px;opacity:.92}.coupon-ad-amount{margin:8px 0 2px;font-size:52px;font-weight:800;line-height:1}.coupon-ad-action{margin-top:10px;min-width:150px;padding:11px 24px;border-radius:24px;color:#dc2626;background:#fff7ed;font-size:16px;font-weight:700}

/* 历史目的地按城市分组展示，位于家/公司快捷入口下方。 */
.search-history{padding-top:14px}
.history-title-row{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-bottom:10px}
.history-title{margin:0;color:#374151;font-size:14px;font-weight:600}
.history-clear{border:0;background:transparent;color:#7C3AED;font-size:13px;white-space:nowrap}
.history-city{padding:2px 0 8px}
.history-city-title{padding:6px 0;color:#111827;font-size:14px;font-weight:700}
.history-list{display:flex;flex-direction:column}
.history-row{width:100%;min-height:58px;display:flex;align-items:center;gap:10px;padding:10px 0;border:0;border-bottom:1px solid #F3F4F6;background:transparent;text-align:left}
.history-row>.van-icon:first-child{flex-shrink:0;color:#9CA3AF}
.history-row .history-info{min-width:0;display:flex;flex:1;flex-direction:column;gap:4px}
.history-row .history-name{max-width:100%;overflow:hidden;color:#111827;font-size:15px;font-weight:500;text-overflow:ellipsis;white-space:nowrap}
.history-row .history-address{max-width:100%;overflow:hidden;color:#6B7280;font-size:13px;text-overflow:ellipsis;white-space:nowrap}
.history-row .history-delete{width:32px;height:32px;display:flex;align-items:center;justify-content:center;margin-left:4px;padding:0;color:#9CA3AF}

</style>
