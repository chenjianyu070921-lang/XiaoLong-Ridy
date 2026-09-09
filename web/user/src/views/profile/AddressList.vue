<template>
  <div class="address-page">
    <!-- 顶部导航：返回个人中心，并提供新增地址入口。 -->
    <div class="page-header">
      <van-icon name="arrow-left" size="20" @click="goBack" />
      <span class="title">常用地址</span>
      <button type="button" class="add-icon-btn" aria-label="新增地址" @click="openCreate">
        <van-icon name="plus" size="20" />
      </button>
    </div>

    <main class="address-content">
      <van-loading v-if="loading" class="loading" color="#7C3AED">加载中...</van-loading>
      <div v-else-if="addresses.length" class="address-list">
        <article v-for="item in addresses" :key="item.id" class="address-item">
          <div class="address-main">
            <div class="address-title-row">
              <strong>{{ item.address }}</strong>
              <span v-if="item.isDefault" class="default-tag">默认</span>
              <span class="type-tag">{{ tagText(item.tag) }}</span>
            </div>
            <p>{{ item.contactName }} {{ item.contactPhone }}</p>
          </div>
          <div class="address-actions">
            <button type="button" aria-label="编辑地址" @click="openEdit(item)">
              <van-icon name="edit" size="18" />
            </button>
            <button type="button" aria-label="删除地址" @click="removeAddress(item)">
              <van-icon name="delete-o" size="18" />
            </button>
          </div>
        </article>
      </div>
      <van-empty v-else description="暂无常用地址">
        <button type="button" class="empty-add-btn" @click="openCreate">新增地址</button>
      </van-empty>
    </main>

    <!-- 地址编辑弹窗：复用同一份表单完成新增和更新。 -->
    <van-popup v-model:show="showEditor" position="bottom" round class="address-popup">
      <section class="editor-panel">
        <div class="editor-head">
          <h2>{{ editingId ? '编辑地址' : '新增地址' }}</h2>
          <button type="button" aria-label="关闭" @click="closeEditor">
            <van-icon name="cross" size="20" />
          </button>
        </div>
        <van-field v-model="form.contactName" label="联系人" placeholder="请输入联系人姓名" maxlength="20" clearable />
        <van-field v-model="form.contactPhone" label="手机号" placeholder="请输入联系人手机号" maxlength="11" type="tel" clearable />
        <van-field
          v-model="poiKeyword"
          label="地址"
          placeholder="搜索小区、写字楼、商圈"
          maxlength="40"
          clearable
          @focus="ensureAMapReady"
        />
        <div v-if="selectedPoi.name" class="selected-poi">
          <strong>{{ selectedPoi.name }}</strong>
          <span>{{ selectedPoi.address }}</span>
        </div>
        <div v-if="poiLoading || poiResults.length || poiMessage" class="poi-result-box">
          <van-loading v-if="poiLoading" class="poi-loading" size="18">搜索中...</van-loading>
          <template v-else>
            <button
              v-for="item in poiResults"
              :key="item.id"
              type="button"
              class="poi-result"
              @click="selectPoi(item)"
            >
              <strong>{{ item.name }}</strong>
              <span>{{ item.address || '暂无详细地址' }}</span>
            </button>
          </template>
          <p v-if="!poiLoading && poiMessage" class="poi-message">{{ poiMessage }}</p>
        </div>
        <van-field v-model="form.detail" label="详细地址" placeholder="门牌号、楼层等补充信息" maxlength="60" clearable />
        <div class="tag-row">
          <span>标签</span>
          <van-radio-group v-model="form.tag" direction="horizontal">
            <van-radio name="home">家</van-radio>
            <van-radio name="work">公司</van-radio>
            <van-radio name="other">其他</van-radio>
          </van-radio-group>
        </div>
        <div class="switch-row">
          <span>设为默认地址</span>
          <van-switch v-model="form.isDefault" size="22" />
        </div>
        <button type="button" class="save-btn" :disabled="saving" @click="saveAddress">
          {{ saving ? '保存中...' : '保存地址' }}
        </button>
      </section>
    </van-popup>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showConfirmDialog, showToast } from 'vant'
import AMapLoader from '@amap/amap-jsapi-loader'
import { getAmapConfig } from '@/config/amap'
import { createAddress, deleteAddress, listAddresses, updateAddress } from '@/api/address'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const saving = ref(false)
const showEditor = ref(false)
const editingId = ref(0)
const addresses = ref([])
const AMapSDK = ref(null)
const poiKeyword = ref('')
const poiLoading = ref(false)
const poiResults = ref([])
const poiMessage = ref('')
const selectedPoi = ref({ name: '', address: '', lng: '', lat: '' })
let searchTimer = null
let searchSequence = 0

// defaultForm 创建地址表单默认值，避免新增和编辑之间残留上一次输入。
const defaultForm = () => ({
  contactName: '',
  contactPhone: '',
  tag: 'home',
  address: '',
  detail: '',
  longitude: '',
  latitude: '',
  isDefault: false
})

const form = ref(defaultForm())

const goBack = () => router.back()

// tagText 将后端地址标签转换为用户可读文案。
const tagText = (tag) => ({ home: '家', work: '公司', other: '其他' })[tag] || '其他'

// ensureAMapReady 初始化高德 POI 搜索能力，地址坐标必须来自用户选择的搜索结果。
const ensureAMapReady = async () => {
  if (AMapSDK.value) return true
  const { key, securityCode } = getAmapConfig()
  if (!key) {
    showToast('未配置高德地图 Key')
    return false
  }
  try {
    if (securityCode) window._AMapSecurityConfig = { securityJsCode: securityCode }
    // 高德 JS API 2.0 使用 AutoComplete 提供关键词提示，InputTips 不是可实例化构造器。
    AMapSDK.value = await AMapLoader.load({ key, version: '2.0', plugins: ['AMap.PlaceSearch', 'AMap.AutoComplete'] })
    return true
  } catch (error) {
    console.error('高德地址搜索初始化失败:', error)
    showToast('地址搜索加载失败')
    return false
  }
}

// readLngLat 兼容高德 SDK 返回的 LngLat 对象和普通经纬度对象。
const readLngLat = (location) => {
  const lng = typeof location?.getLng === 'function' ? location.getLng() : Number(location?.lng)
  const lat = typeof location?.getLat === 'function' ? location.getLat() : Number(location?.lat)
  return Number.isFinite(lng) && Number.isFinite(lat) ? { lng, lat } : null
}

// normalizePoi 将高德提示和 POI 结果统一成页面可选择的地址项。
const normalizePoi = (poi) => {
  const point = readLngLat(poi?.location)
  if (!point || !poi?.name) return null
  const region = [poi.pname, poi.cityname, poi.adname, poi.district].filter(Boolean).join('')
  const address = [region, poi.address].filter(Boolean).join('') || poi.name
  return {
    id: poi.id || `${point.lng},${point.lat}`,
    name: poi.name,
    address,
    lng: point.lng,
    lat: point.lat
  }
}

// searchPoi 调用高德 POI 搜索，避免让用户手动填写经纬度。
const searchPoi = async (keyword) => {
  const sequence = ++searchSequence
  poiLoading.value = true
  poiMessage.value = ''
  poiResults.value = []
  if (!await ensureAMapReady()) {
    poiLoading.value = false
    return
  }
  const addResults = (items = []) => {
    if (sequence !== searchSequence) return
    const map = new Map(poiResults.value.map(item => [item.id, item]))
    items.map(normalizePoi).filter(Boolean).forEach(item => {
      if (!map.has(item.id)) map.set(item.id, item)
    })
    poiResults.value = Array.from(map.values()).slice(0, 20)
    poiMessage.value = poiResults.value.length ? '' : '未找到相关地址'
  }
  let pending = 2
  const finish = () => {
    pending -= 1
    if (pending <= 0 && sequence === searchSequence) poiLoading.value = false
  }
  try {
    new AMapSDK.value.AutoComplete({ pageSize: 10 }).search(keyword, (status, result) => {
      if (status === 'complete' && Array.isArray(result?.tips)) addResults(result.tips)
      finish()
    })
  } catch (error) {
    console.warn('高德地址提示搜索失败:', error)
    finish()
  }
  try {
    new AMapSDK.value.PlaceSearch({ pageSize: 20, pageIndex: 1, extensions: 'all' }).search(keyword, (status, result) => {
      if (status === 'complete' && Array.isArray(result?.poiList?.pois)) addResults(result.poiList.pois)
      finish()
    })
  } catch (error) {
    console.warn('高德 POI 搜索失败:', error)
    finish()
  }
}

// selectPoi 保存用户点击的地址和对应坐标，后续提交给后端。
const selectPoi = (item) => {
  selectedPoi.value = item
  poiKeyword.value = item.name
  poiResults.value = []
  poiMessage.value = ''
  form.value.address = item.name
  form.value.longitude = String(item.lng)
  form.value.latitude = String(item.lat)
}

// loadAddresses 从后端读取当前乘客的常用地址，并按默认地址和排序稳定展示。
const loadAddresses = async () => {
  loading.value = true
  try {
    const res = await listAddresses()
    addresses.value = Array.isArray(res?.list) ? res.list : []
  } catch (error) {
    showToast(error?.response?.data?.message || '地址加载失败')
    addresses.value = []
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  editingId.value = 0
  form.value = defaultForm()
  poiKeyword.value = ''
  selectedPoi.value = { name: '', address: '', lng: '', lat: '' }
  poiResults.value = []
  poiMessage.value = ''
  showEditor.value = true
}

// openCreateFromRoute 根据首页快捷入口带来的 tag 参数打开新增地址弹窗。
const openCreateFromRoute = () => {
  const tag = ['home', 'work', 'other'].includes(route.query.tag) ? route.query.tag : 'home'
  openCreate()
  form.value.tag = tag
}

const openEdit = (item) => {
  editingId.value = Number(item.id || 0)
  form.value = {
    contactName: item.contactName || '',
    contactPhone: item.contactPhone || '',
    tag: item.tag || 'other',
    address: item.address || '',
    detail: '',
    longitude: String(item.longitude || ''),
    latitude: String(item.latitude || ''),
    isDefault: Boolean(item.isDefault)
  }
  poiKeyword.value = item.address || ''
  selectedPoi.value = {
    name: item.address || '',
    address: item.address || '',
    lng: item.longitude || '',
    lat: item.latitude || ''
  }
  poiResults.value = []
  poiMessage.value = ''
  showEditor.value = true
}

const closeEditor = () => {
  if (!saving.value) showEditor.value = false
}

// validateForm 校验地址表单，后端要求联系人、手机号、地址和经纬度都必须有效。
const validateForm = () => {
  const contactName = form.value.contactName.trim()
  const contactPhone = form.value.contactPhone.trim()
  const longitude = Number(form.value.longitude)
  const latitude = Number(form.value.latitude)
  if (!contactName) return showToast('请输入联系人'), false
  if (!/^1\d{10}$/.test(contactPhone)) return showToast('手机号格式不正确'), false
  if (!form.value.address.trim()) return showToast('请先选择地址'), false
  if (!Number.isFinite(longitude) || longitude < -180 || longitude > 180 || longitude === 0) return showToast('请从搜索结果选择地址'), false
  if (!Number.isFinite(latitude) || latitude < -90 || latitude > 90 || latitude === 0) return showToast('请从搜索结果选择地址'), false
  return true
}

// buildPayload 将表单字段转换为后端地址接口需要的结构。
const buildPayload = () => {
  const baseAddress = form.value.address.trim()
  const detail = form.value.detail.trim()
  return {
    contactName: form.value.contactName.trim(),
    contactPhone: form.value.contactPhone.trim(),
    tag: form.value.tag || 'other',
    address: detail ? `${baseAddress} ${detail}` : baseAddress,
    longitude: Number(form.value.longitude),
    latitude: Number(form.value.latitude),
    isDefault: Boolean(form.value.isDefault)
  }
}

const saveAddress = async () => {
  if (!validateForm()) return
  saving.value = true
  try {
    const payload = buildPayload()
    if (editingId.value) {
      await updateAddress({ id: editingId.value, ...payload })
      showToast('地址已更新')
    } else {
      await createAddress(payload)
      showToast('地址已新增')
    }
    showEditor.value = false
    await loadAddresses()
  } catch (error) {
    showToast(error?.response?.data?.message || '地址保存失败')
  } finally {
    saving.value = false
  }
}

const removeAddress = async (item) => {
  try {
    await showConfirmDialog({
      title: '删除地址',
      message: `确定删除“${item.address}”吗？`
    })
    await deleteAddress(item.id)
    showToast('地址已删除')
    await loadAddresses()
  } catch (error) {
    if (error) showToast(error?.response?.data?.message || '删除失败')
  }
}

onMounted(() => {
  loadAddresses()
  if (route.query.mode === 'create') openCreateFromRoute()
})

watch(poiKeyword, value => {
  clearTimeout(searchTimer)
  const keyword = value.trim()
  if (!keyword || keyword === selectedPoi.value.name) {
    poiResults.value = []
    poiMessage.value = ''
    poiLoading.value = false
    return
  }
  form.value.address = ''
  form.value.longitude = ''
  form.value.latitude = ''
  selectedPoi.value = { name: '', address: '', lng: '', lat: '' }
  searchTimer = setTimeout(() => searchPoi(keyword), 350)
})
</script>

<style scoped>
.address-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.page-header {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: #fff;
}

.title {
  color: var(--text-primary);
  font-size: 17px;
  font-weight: 600;
}

.add-icon-btn,
.editor-head button,
.address-actions button {
  display: flex;
  align-items: center;
  justify-content: center;
  border: 0;
  color: #374151;
  background: transparent;
}

.add-icon-btn {
  width: 32px;
  height: 32px;
}

.address-content {
  padding: 14px 12px 28px;
}

.loading {
  display: flex;
  justify-content: center;
  padding-top: 56px;
}

.address-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.address-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 14px;
  border-radius: 8px;
  background: #fff;
}

.address-main {
  min-width: 0;
  flex: 1;
}

.address-title-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
}

.address-title-row strong {
  min-width: 0;
  overflow: hidden;
  color: #111827;
  font-size: 15px;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.default-tag,
.type-tag {
  flex-shrink: 0;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
}

.default-tag {
  color: #059669;
  background: #ecfdf5;
}

.type-tag {
  color: #7c3aed;
  background: #f5f3ff;
}

.address-main p,
.address-main small {
  display: block;
  margin: 4px 0 0;
  color: #6b7280;
  font-size: 12px;
}

.address-actions {
  display: flex;
  flex-shrink: 0;
  gap: 4px;
}

.address-actions button {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: #f9fafb;
}

.empty-add-btn,
.save-btn {
  border: 0;
  color: #fff;
  background: #7c3aed;
  font-weight: 600;
}

.empty-add-btn {
  min-width: 112px;
  height: 38px;
  border-radius: 19px;
}

.address-popup {
  max-height: 86vh;
  overflow-y: auto;
}

.editor-panel {
  padding: 18px 16px 24px;
  background: #fff;
}

.editor-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.editor-head h2 {
  margin: 0;
  color: #111827;
  font-size: 18px;
}

.editor-head button {
  width: 32px;
  height: 32px;
}

.selected-poi {
  display: flex;
  flex-direction: column;
  gap: 3px;
  margin: 8px 16px 4px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #f5f3ff;
}

.selected-poi strong {
  color: #4c1d95;
  font-size: 14px;
}

.selected-poi span {
  color: #6b7280;
  font-size: 12px;
  line-height: 1.4;
}

.poi-result-box {
  margin: 6px 16px 10px;
  border: 1px solid #eef0f4;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
}

.poi-loading,
.poi-message {
  display: flex;
  justify-content: center;
  margin: 0;
  padding: 12px;
  color: #6b7280;
  font-size: 13px;
}

.poi-result {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
  padding: 11px 12px;
  border: 0;
  border-bottom: 1px solid #f3f4f6;
  text-align: left;
  background: #fff;
}

.poi-result:last-child {
  border-bottom: 0;
}

.poi-result strong {
  overflow: hidden;
  color: #111827;
  font-size: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.poi-result span {
  color: #6b7280;
  font-size: 12px;
  line-height: 1.35;
}

.tag-row,
.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 48px;
  padding: 0 16px;
  border-bottom: 1px solid #f3f4f6;
  color: #323233;
  font-size: 14px;
}

.save-btn {
  width: 100%;
  height: 44px;
  margin-top: 18px;
  border-radius: 22px;
  font-size: 15px;
}

.save-btn:disabled {
  opacity: .6;
}
</style>
