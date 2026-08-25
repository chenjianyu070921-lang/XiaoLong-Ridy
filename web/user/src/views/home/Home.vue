<template>
  <div class="home-page">
    <header class="home-header">
      <div class="brand-mark">
        <img src="/logo.png" alt="花小龙" />
        <div><strong>花小龙出行</strong><span>便捷 · 安心 · 省心</span></div>
      </div>
      <button class="city-switch" type="button" aria-label="选择城市" @click="router.push('/city/select')"><van-icon name="location" size="16" />{{ selectedCity?.name || currentCityName || '定位中' }}<van-icon name="arrow-down" size="12" /></button>
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

        <div v-if="searchMode === 'destination' && !keyword.trim() && recentDestinations.length" class="search-history">
          <div class="history-title">历史目的地</div>
          <div class="history-chips">
            <button v-for="item in recentDestinations.slice(0, 3)" :key="item.id || item.name" type="button" class="history-chip" @click="selectSearchResult(item)">
              <van-icon name="clock-o" /><span>{{ item.name }}</span>
            </button>
          </div>
        </div>

        <div class="result-section-title">{{ keyword.trim() ? '搜索结果' : searchMode === 'pickup' ? '附近地点' : '热门目的地' }}</div>
        <div v-if="searchLoading" class="search-status"><van-loading size="24px" vertical>正在搜索地点...</van-loading></div>
        <div v-else-if="searchResults.length" class="search-results">
          <button v-for="item in searchResults" :key="item.id" type="button" class="result-item" @click="selectSearchResult(item)">
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
import { showToast } from 'vant'
import { useOrderStore } from '@/stores/order'
import { getOrders } from '@/api/order'
import { readSelectedCity, saveSelectedCity } from '@/data/cities'


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
const currentPoint = ref({ lat: null, lng: null, address: '' })
// 当前定位所属行政区编码，优先使用 adcode 精确限制 POI 搜索范围。
const selectedCity = ref(localStorage.getItem('passenger-manual-city') === '1' ? readSelectedCity() : null)
// 历史版本可能保存了“当前位置”占位值，不能把它当作真实城市展示。
if (selectedCity.value?.name === '当前位置') {
  selectedCity.value = null
  localStorage.removeItem('passenger-selected-city')
  localStorage.removeItem('passenger-manual-city')
}
const currentCityName = ref('')
const currentCityCode = ref(selectedCity.value?.adcode || '')

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

// 从本地缓存读取最近目的地，异常数据会被过滤，避免影响地图搜索。
function loadRecentDestinations() {
  try {
    const list = JSON.parse(localStorage.getItem(recentDestinationKey) || '[]')
    return Array.isArray(list) ? list.filter(item => item?.name && Number.isFinite(item.lat) && Number.isFinite(item.lng)).slice(0, 10) : []
  } catch (error) {
    console.warn('读取历史目的地失败:', error)
    return []
  }
}

function rememberDestination(item) {
  const saved = { id: item.id || '', name: item.name, address: item.address || '', lat: item.lat, lng: item.lng }
  const rest = recentDestinations.value.filter(entry => entry.id !== saved.id && (entry.lat !== saved.lat || entry.lng !== saved.lng))
  recentDestinations.value = [saved, ...rest].slice(0, 10)
  localStorage.setItem(recentDestinationKey, JSON.stringify(recentDestinations.value))
}

// 初始化高德地图及定位、地点搜索和逆地理编码服务。
async function initMap() {
  const key = import.meta.env.VITE_AMAP_KEY || ''
  const securityCode = import.meta.env.VITE_AMAP_SECURITY_CODE || ''
  if (!key) {
    locationFailed.value = true
    showToast('未配置高德地图 Key')
    return
  }
  try {
    if (securityCode) window._AMapSecurityConfig = { securityJsCode: securityCode }
    const AMap = await AMapLoader.load({ key, version: '2.0', plugins: ['AMap.Geolocation', 'AMap.PlaceSearch', 'AMap.CitySearch', 'AMap.Geocoder'] })
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
    currentCityCode.value = result.adcode
    configurePlaceSearch(result.adcode)
  }
  return result
}

function updateCurrentMarker(lng, lat) {
  if (!mapInstance.value || !AMapSDK.value) return
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
  let resolvedAdcode = addressComponent?.adcode || ''
  let resolvedCityName = addressComponent?.city || addressComponent?.province || ''
  if (!text || ['当前位置', '我的位置', '当前城市'].includes(text)) {
    const reverseResult = await reverseGeocode(normalizedLng, normalizedLat)
    text = reverseResult.address
    resolvedAdcode = resolvedAdcode || reverseResult.adcode
    resolvedCityName = resolvedCityName || reverseResult.cityName
  }
  if (!text || ['当前位置', '我的位置', '当前城市'].includes(text)) {
    return markLocationFailure('无法解析当前位置详细地址，请重新定位')
  }
  locationFailed.value = false
  currentPoint.value = { lng: normalizedLng, lat: normalizedLat, address: text }
  // 定位结果只作为默认城市来源；手动选择城市后，不能再被设备定位覆盖。
  if (!selectedCity.value) {
    if (resolvedAdcode) {
      currentCityCode.value = resolvedAdcode
      currentCityName.value = resolvedCityName || '当前城市'
      configurePlaceSearch(currentCityCode.value)
    } else {
      // 浏览器定位或高德定位未返回 adcode 时，通过坐标逆地理编码补齐行政区编码。
      const cityResult = await resolveCityCode(normalizedLng, normalizedLat)
      currentCityName.value = cityResult.cityName || currentCityName.value
    }
  } else {
    currentCityCode.value = selectedCity.value.adcode
    currentCityName.value = selectedCity.value.name
    configurePlaceSearch(currentCityCode.value)
  }
  // 自动定位只更新地图当前位置；只有用户明确选择地点时，才写入订单上车点。
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
  currentCityCode.value = city.adcode
  configurePlaceSearch(city.adcode)
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
    await applyCurrentLocation(result, false)
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
  return { id: poi.id || `${lng},${lat}`, name: poi.name || '未命名地点', address: detailedAddress, displayAddress: detailedAddress, lng, lat, distanceText: Number.isFinite(distance) ? distance < 1000 ? `${Math.round(distance)}m` : `${(distance / 1000).toFixed(1)}km` : '' }
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

function searchByKeyword(value) {
  const sequence = ++searchSequence
  if (!placeSearch.value) return Object.assign(searchLoading, { value: false })
  searchLoading.value = true
  searchMessage.value = ''
  placeSearch.value.setType('')
  const normalizedKeyword = value.replace(/\s+/g, '').replace(/市$/, '')
  const cityName = (currentCityName.value || selectedCity.value?.name || '').replace(/市$/, '')
  const cityKeyword = cityName && !normalizedKeyword.startsWith(cityName) ? cityName + normalizedKeyword : value
  const isRelevant = item => {
    const text = (item.name + item.address).replace(/\s+/g, '')
    return text.includes(normalizedKeyword) || normalizedKeyword.includes(item.name.replace(/\s+/g, ''))
  }
  const finish = (status, result) => {
    const relevant = (result?.poiList?.pois || []).map(normalizePOI).filter(Boolean).filter(isRelevant)
    if (sequence !== searchSequence) return
    searchLoading.value = false
    searchResults.value = relevant
    searchMessage.value = relevant.length ? '' : '当前城市未找到匹配地点，请换个关键词试试'
  }
  // 优先将当前城市拼入关键词，避免高德把“宝龙广场”误匹配到全国其他城市的弱相关地点。
  const citySearch = new AMapSDK.value.PlaceSearch({ pageSize: 50, pageIndex: 1, extensions: 'base', city: currentCityCode.value || cityName || undefined, citylimit: Boolean(currentCityCode.value || cityName) })
  citySearch.setType('')
  citySearch.search(cityKeyword, (status, result) => {
    const pois = result?.poiList?.pois || []
    if (status === 'complete' && pois.length) {
      finish(status, result)
      return
    }
    placeSearch.value.search(value, finish)
  })
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
  destinationAddress.value = item.name
  orderStore.setOrderParams({ toAddress: item.name, toLat: item.lat, toLng: item.lng, cityCode: currentCityCode.value })
  rememberDestination(item)
  closeLocationSearch()
  if (mapInstance.value) {
    const position = [item.lng, item.lat]
    if (!destinationMarker.value && AMapSDK.value) {
      destinationMarker.value = new AMapSDK.value.Marker({ position })
      mapInstance.value.add(destinationMarker.value)
    } else destinationMarker.value?.setPosition(position)
    mapInstance.value.setCenter(position)
  }
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
.call-car-btn{width:100%;height:48px;font-size:16px;margin-top:16px}.search-popup{padding:16px;height:100%;overflow-y:auto}.search-mode-title{margin-bottom:14px;color:#111827;font-size:18px;font-weight:600;text-align:center}.result-section-title{padding:18px 0 8px;color:#374151;font-size:15px;font-weight:600}.history-chips{display:flex;gap:8px;overflow-x:auto}.history-chip{min-width:0;display:flex;align-items:center;gap:6px;padding:8px 12px;border:1px solid #E5E7EB;border-radius:6px;background:#fff;color:#374151;white-space:nowrap}.search-header{display:flex;align-items:center;gap:12px;padding-bottom:16px;border-bottom:1px solid #F3F4F6}.search-input{flex:1;height:40px;border:none;outline:none;font-size:15px;background:transparent}.cancel-btn{color:var(--primary-color);font-size:14px;cursor:pointer;white-space:nowrap}.search-results{padding-top:16px}.search-status{display:flex;justify-content:center;padding:56px 0;color:#6b7280}.result-item{width:100%;min-height:68px;display:flex;align-items:center;gap:12px;padding:12px 0;border:0;border-bottom:1px solid #F3F4F6;background:transparent;text-align:left;cursor:pointer}.result-info{min-width:0;display:flex;flex:1;flex-direction:column}.result-distance{flex-shrink:0;color:#9ca3af;font-size:12px}.result-info .name{font-size:15px;font-weight:500;color:var(--text-primary);margin-bottom:4px}.result-info .address{font-size:13px;color:#6b7280}.empty-state{text-align:center;padding:60px 0;color:#9ca3af}.coupon-ad-mask{position:fixed;top:0;right:0;bottom:0;left:0;z-index:30;display:flex;align-items:center;justify-content:center;padding:24px;background:#0f172a7a}.coupon-ad{position:relative;width:min(330px,100%);overflow:hidden;border-radius:16px;background:linear-gradient(145deg,#ef4444,#f97316);box-shadow:0 18px 42px #7f1d1d59}.coupon-ad-close{position:absolute;top:10px;right:10px;z-index:1;width:36px;height:36px;border:0;border-radius:50%;color:#fff;background:#7f1d1d47}.login-coupon-ad{width:min(390px,100%);background:#fff}.login-coupon-ad .coupon-ad-body{padding:0}.login-coupon-ad .coupon-gift-image{max-height:calc(100vh - 72px);width:100%;object-fit:contain}.login-coupon-close{top:auto;right:50%;bottom:-52px;transform:translateX(50%);background:#ffffff;color:#6b21a8;box-shadow:0 2px 10px #0003}.coupon-ad-body{display:flex;width:100%;min-height:300px;align-items:center;flex-direction:column;justify-content:center;gap:8px;padding:42px 28px 28px;border:0;color:#fff;background:transparent}.coupon-gift-image{display:block;width:100%;max-height:420px;object-fit:contain}.coupon-ad-kicker{font-size:14px;opacity:.9}.coupon-ad-title{font-size:26px;line-height:1.25}.coupon-ad-subtitle{font-size:15px;opacity:.92}.coupon-ad-amount{margin:8px 0 2px;font-size:52px;font-weight:800;line-height:1}.coupon-ad-action{margin-top:10px;min-width:150px;padding:11px 24px;border-radius:24px;color:#dc2626;background:#fff7ed;font-size:16px;font-weight:700}

</style>