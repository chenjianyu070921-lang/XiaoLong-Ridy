<template>
  <van-popup
    :show="visible"
    position="bottom"
    teleport="#driver-home-popups"
    class="home-dest-panel"
    :style="{ left: '0', right: '0', width: 'min(100vw, 390px)', height: '100%', margin: '0 auto', borderRadius: 0 }"
    @update:show="(v) => emit('update:visible', v)"
  >
    <div class="dest-header">
      <button type="button" class="back" @click="close">返回</button>
      <span class="dest-title">回家顺路目的地</span>
    </div>

    <div class="dest-body">
      <!-- 已开启：展示当前目的地，支持关闭回家模式 -->
      <template v-if="setting?.isHomeModeOpen">
        <div class="dest-card">
          <span class="dest-label">当前回家目的地</span>
          <strong class="dest-addr">{{ setting.homeAddr || '未命名地址' }}</strong>
          <span class="dest-coord">{{ coordText }}</span>
          <p class="dest-hint">已开启回家顺路模式：只推送与回家方向一致的订单，反向单不再打扰你。到达目的地后不会自动下线，需手动停止听单。</p>
          <button type="button" class="ghost-btn" @click="changeAddress">修改目的地</button>
          <button type="button" class="close-mode-btn" :disabled="submitting" @click="turnOff">关闭回家模式</button>
        </div>
      </template>

      <!-- 未开启/修改中：POI 联想选择地址 -->
      <template v-else>
        <div class="dest-card">
          <span class="dest-label">设置回家目的地</span>
          <input
            v-model="keyword"
            type="text"
            class="dest-input"
            placeholder="输入小区 / 地标 / 街道名称"
            @input="searchPoi"
          />
          <p v-if="poiError" class="dest-error">{{ poiError }}</p>

          <ul v-if="poiList.length" class="poi-list">
            <li
              v-for="(poi, index) in poiList"
              :key="poi.id || index"
              class="poi-item"
              :class="{ active: selected && selected.id === (poi.id || index) }"
              @click="selectPoi(poi, index)"
            >
              <div class="poi-main">
                <strong>{{ poi.name }}</strong>
                <span>{{ poi.district || poi.address || poi.city || '' }}</span>
              </div>
              <van-icon v-if="selected && selected.id === (poi.id || index)" name="success" />
            </li>
          </ul>
          <p v-else-if="keyword && !searching" class="dest-hint">没有匹配地点，换个关键词试试（请从候选列表点选，避免地理编码失败）。</p>

          <div class="dest-actions">
            <button type="button" class="primary-btn" :disabled="!canSubmit || submitting" @click="saveAndOpen">
              {{ submitting ? '保存中…' : '保存并开启回家模式' }}
            </button>
            <button v-if="setting?.hasSetting" type="button" class="ghost-btn" @click="cancelEdit">取消修改</button>
          </div>
          <p class="dest-hint">未开启回家模式时，你可以随时保存地址；开启后听单规则才切换为顺路过滤。</p>
        </div>
      </template>
    </div>
  </van-popup>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { showToast } from 'vant'
import { setHomeDestination, setHomeMode } from '@/api/driver'
import { loadDriverAmap } from '@/config/amap'

const props = defineProps({
  visible: { type: Boolean, default: false },
  // 当前设置：{ hasSetting, homeAddr, homeLng, homeLat, isHomeModeOpen, maxDetourRatio }
  setting: { type: Object, default: () => ({}) }
})
const emit = defineEmits(['update:visible', 'saved'])

const keyword = ref('')
const poiList = ref([])
const searching = ref(false)
const poiError = ref('')
const submitting = ref(false)
const selected = ref(null)
let autoComplete = null

const coordText = computed(() => {
  const lng = Number(props.setting?.homeLng || 0)
  const lat = Number(props.setting?.homeLat || 0)
  if (!lng || !lat) return ''
  return `${lng.toFixed(6)}, ${lat.toFixed(6)}`
})

const canSubmit = computed(() => !!selected.value?.lng && !!selected.value?.lat)

function close() {
  emit('update:visible', false)
}

function changeAddress() {
  keyword.value = ''
  poiList.value = []
  selected.value = null
  // 通过父级将 setting 切到"编辑态"：直接触发一次空保存不改开关即可进入输入界面
  emit('update:visible', true)
  editing.value = true
}

const editing = ref(false)

function cancelEdit() {
  editing.value = false
  keyword.value = ''
  poiList.value = []
  selected.value = null
}

async function ensureAutoComplete() {
  if (autoComplete) return autoComplete
  try {
    const AMap = await loadDriverAmap(['AMap.AutoComplete'])
    autoComplete = new AMap.AutoComplete({ city: '全国', citylimit: false })
    return autoComplete
  } catch (e) {
    poiError.value = '地图组件加载失败，请检查网络后重试'
    return null
  }
}

let searchTimer = null
function searchPoi() {
  if (searchTimer) clearTimeout(searchTimer)
  const text = keyword.value.trim()
  if (!text) {
    poiList.value = []
    return
  }
  searchTimer = setTimeout(() => doSearch(text), 250)
}

async function doSearch(text) {
  const ac = await ensureAutoComplete()
  if (!ac) return
  searching.value = true
  poiError.value = ''
  ac.search(text, (status, result) => {
    searching.value = false
    if (status !== 'complete' || !result || !result.tips) {
      poiList.value = []
      return
    }
    poiList.value = result.tips
      .filter((tip) => tip.location)
      .map((tip, index) => ({
        id: tip.id || `${tip.name}-${index}`,
        name: tip.name,
        district: tip.district,
        address: tip.address,
        lng: tip.location.lng,
        lat: tip.location.lat
      }))
  })
}

function selectPoi(poi, index) {
  selected.value = { ...poi, id: poi.id || index }
  keyword.value = poi.name
  poiList.value = []
}

async function saveAndOpen() {
  if (!canSubmit.value || submitting.value) return
  submitting.value = true
  try {
    const saved = await setHomeDestination({
      homeAddr: selected.value.name,
      homeLng: selected.value.lng,
      homeLat: selected.value.lat,
      open: true
    })
    editing.value = false
    showToast('已开启回家顺路模式')
    emit('saved', saved)
    close()
  } catch (e) {
    showToast(e?.response?.data?.message || '保存失败，请重试')
  } finally {
    submitting.value = false
  }
}

async function turnOff() {
  if (submitting.value) return
  submitting.value = true
  try {
    const saved = await setHomeMode({ open: false })
    showToast('已关闭回家模式，恢复全域听单')
    emit('saved', saved)
    close()
  } catch (e) {
    showToast(e?.response?.data?.message || '操作失败，请重试')
  } finally {
    submitting.value = false
  }
}

watch(
  () => props.visible,
  (v) => {
    if (v) {
      editing.value = !props.setting?.isHomeModeOpen
      keyword.value = ''
      poiList.value = []
      selected.value = null
      poiError.value = ''
    }
  }
)
</script>

<style scoped>
.home-dest-panel {
  display: flex;
  flex-direction: column;
  background: var(--driver-soft);
}
.dest-header {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--driver-card);
  border-bottom: 1px solid var(--driver-line);
  padding: 12px;
}
.dest-header .back { border: none; background: none; color: var(--driver-primary); font-size: 14px; padding: 4px 8px; }
.dest-title { font-size: 16px; font-weight: 700; color: var(--driver-ink); }
.dest-body { flex: 1; overflow-y: auto; padding: 12px; }
.dest-card {
  background: var(--driver-card);
  border: 1px solid var(--driver-line);
  border-radius: 12px;
  padding: 14px;
  display: grid;
  gap: 10px;
}
.dest-label { color: var(--driver-muted); font-size: 12px; }
.dest-addr { color: var(--driver-ink); font-size: 16px; }
.dest-coord { color: var(--driver-muted); font-size: 12px; }
.dest-hint { margin: 0; color: var(--driver-muted); font-size: 12px; line-height: 1.6; }
.dest-error { margin: 0; color: #DC2626; font-size: 13px; }
.dest-input {
  width: 100%;
  height: 42px;
  padding: 0 12px;
  border: 1px solid var(--driver-line);
  border-radius: 10px;
  background: var(--driver-bg);
  color: var(--driver-ink);
  font-size: 14px;
  outline: none;
}
.dest-input:focus { border-color: var(--driver-primary); }
.poi-list { margin: 0; padding: 0; list-style: none; display: grid; gap: 6px; max-height: 320px; overflow-y: auto; }
.poi-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px;
  border-radius: 8px;
  background: var(--driver-bg);
  color: var(--driver-ink);
}
.poi-item.active { background: var(--driver-soft); border: 1px solid var(--driver-primary); }
.poi-main { display: grid; gap: 2px; min-width: 0; }
.poi-main strong { font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.poi-main span { color: var(--driver-muted); font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dest-actions { display: grid; gap: 8px; }
.primary-btn {
  height: 44px; border: none; border-radius: 22px;
  background: var(--driver-primary); color: #fff; font-size: 15px; font-weight: 700;
}
.primary-btn:disabled { opacity: .55; }
.ghost-btn {
  height: 40px; border: 1px solid var(--driver-line); border-radius: 20px;
  background: transparent; color: var(--driver-ink); font-size: 14px;
}
.close-mode-btn {
  height: 40px; border: 1px solid var(--driver-primary); border-radius: 20px;
  background: transparent; color: var(--driver-primary); font-size: 14px; font-weight: 700;
}
</style>
