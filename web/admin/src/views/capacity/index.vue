<script setup>
// 实时运力页面：使用后台聚合快照展示真实司机位置，位置缺失时保留状态列表。
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Location } from '@element-plus/icons-vue'
import { capacityApi } from '../../api/modules'
import { getAmapConfig } from '../../config/amap'

const loading = ref(false)
const error = ref('')
const snapshot = ref({ drivers: [] })
const status = ref(0)
const onlineStatus = ref(0)
const autoRefresh = ref(true)
let timer
let map
let markers = []

const drivers = computed(() => snapshot.value.drivers || [])
const statusText = (value) => ({ 1: '待审核', 2: '正常', 3: '冻结', 4: '已注销' }[value] || '-')
const onlineText = (value) => ({ 1: '在线', 2: '行程中', 0: '离线' }[value] || '-')

// loadSnapshot 读取聚合快照并刷新地图标记；失败时保留上次结果，避免页面闪烁。
const loadSnapshot = async () => {
  loading.value = true
  error.value = ''
  try {
    snapshot.value = await capacityApi.map({ status: status.value || undefined, online_status: onlineStatus.value || undefined, limit: 200 })
    renderMarkers()
  } catch (e) {
    error.value = e?.message || '运力数据加载失败'
  } finally {
    loading.value = false
  }
}

// ensureMap 延迟加载高德 SDK；SDK 不可用时页面仍显示司机状态表。
const ensureMap = () => new Promise((resolve, reject) => {
  if (window.AMap) return resolve(window.AMap)
  const config = getAmapConfig()
  const script = document.createElement('script')
  script.src = `https://webapi.amap.com/maps?v=2.0&key=${config.key}&securityJsCode=${config.securityCode}`
  script.onload = () => window.AMap ? resolve(window.AMap) : reject(new Error('地图 SDK 不可用'))
  script.onerror = () => reject(new Error('地图 SDK 加载失败'))
  document.head.appendChild(script)
})

const renderMarkers = () => {
  if (!map || !window.AMap) return
  markers.forEach((marker) => map.remove(marker))
  markers = drivers.value.map((item) => {
    const marker = new window.AMap.Marker({ position: [Number(item.lng), Number(item.lat)], title: `司机 ${item.driver_id}` })
    marker.setLabel({ content: `<span class="capacity-label">${item.driver_id} ${onlineText(item.online_status)}</span>`, direction: 'top' })
    marker.setMap(map)
    return marker
  })
}

const initMap = async () => {
  try {
    const AMap = await ensureMap()
    map = new AMap.Map('capacity-map', { zoom: 11, center: [116.397428, 39.90923] })
    renderMarkers()
  } catch (e) {
    error.value = e.message
  }
}

const refresh = async () => { await loadSnapshot(); ElMessage.success('运力数据已刷新') }
const resetTimer = () => { clearInterval(timer); if (autoRefresh.value) timer = setInterval(loadSnapshot, 15000) }
onMounted(async () => { await initMap(); await loadSnapshot(); resetTimer() })
onBeforeUnmount(() => { clearInterval(timer); markers.forEach((marker) => marker.setMap(null)); map?.destroy() })
</script>

<template>
  <section class="capacity" v-loading="loading">
    <div class="page-head"><div><span class="eyebrow">调度中心 / 实时运力</span><h1>实时运力地图</h1><p>位置来自司机最新上报，单点位置失败不会影响其他司机展示</p></div><div class="actions"><el-switch v-model="autoRefresh" active-text="自动刷新" @change="resetTimer" /><el-button :icon="Refresh" @click="refresh">刷新</el-button></div></div>
    <el-alert v-if="error" :title="error" type="warning" show-icon :closable="false" />
    <div class="filters"><el-select v-model="status" clearable placeholder="司机状态" @change="loadSnapshot"><el-option :value="2" label="正常"/><el-option :value="3" label="冻结"/></el-select><el-select v-model="onlineStatus" clearable placeholder="在线状态" @change="loadSnapshot"><el-option :value="1" label="在线"/><el-option :value="2" label="行程中"/></el-select></div>
    <div class="summary"><div><span>地图司机</span><strong>{{ drivers.length }}</strong></div><div><span>可接单</span><strong>{{ snapshot.available_count || 0 }}</strong></div><div><span>行程中</span><strong>{{ snapshot.in_trip_count || 0 }}</strong></div><div><span>位置失败</span><strong>{{ snapshot.position_failure_count || 0 }}</strong></div><div><span>数据时间</span><strong>{{ snapshot.generated_at || '-' }}</strong></div></div>
    <div class="map-layout"><div id="capacity-map" class="map"></div><div class="panel"><div class="section-title"><el-icon><Location /></el-icon>司机状态</div><el-table :data="drivers" empty-text="暂无可用位置"><el-table-column prop="driver_id" label="司机 ID" width="90"/><el-table-column label="状态" width="90"><template #default="scope">{{ onlineText(scope.row.online_status) }}</template></el-table-column><el-table-column label="司机状态" width="100"><template #default="scope">{{ statusText(scope.row.driver_status) }}</template></el-table-column><el-table-column prop="speed_kmh" label="速度" width="90"/><el-table-column prop="report_time" label="最后上报" min-width="165"/></el-table></div></div>
  </section>
</template>

<style scoped>
.capacity{color:var(--text-color,#2e2c4e)}.page-head{display:flex;align-items:flex-end;justify-content:space-between;margin-bottom:18px}.eyebrow{color:var(--brand,#6c5ce7);font-size:12px;font-weight:600}.page-head h1{margin:7px 0 4px;font-size:26px}.page-head p{margin:0;color:var(--muted-color,#8b88a3)}.actions{display:flex;gap:12px;align-items:center}.filters{display:flex;gap:10px;margin:16px 0}.summary{display:grid;grid-template-columns:repeat(5,1fr);gap:12px;margin-bottom:16px}.summary>div,.panel{padding:16px;background:var(--panel-bg,#fff);border:1px solid var(--border-color,#e5e4f0);border-radius:8px}.summary span{display:block;color:var(--muted-color,#8b88a3);font-size:12px}.summary strong{display:block;margin-top:7px;font-size:20px}.map-layout{display:grid;grid-template-columns:minmax(0,1.5fr) minmax(420px,1fr);gap:16px}.map{height:560px;background:#eef2f6;border:1px solid var(--border-color,#e5e4f0);border-radius:8px}.section-title{margin-bottom:12px;font-weight:700;display:flex;gap:6px;align-items:center}@media(max-width:1000px){.summary{grid-template-columns:repeat(2,1fr)}.map-layout{grid-template-columns:1fr}.map{height:420px}}@media(max-width:600px){.page-head{display:block}.actions{margin-top:14px}.summary{grid-template-columns:1fr}}
</style>
