<template>
  <div class="city-page">
    <header class="city-header"><button class="icon-btn" type="button" @click="router.back()"><van-icon name="arrow-left" /></button><h1>选择城市</h1><button class="cancel-btn" type="button" @click="router.back()">取消</button></header>
    <div class="city-search"><van-icon name="search" color="#9CA3AF" /><input v-model="keyword" placeholder="城市中文名或拼音" /></div>
    <div class="current-city">当前城市：<strong>{{ selectedCity?.name || '暂未获取定位城市' }}</strong></div>
    <main class="city-content"><section v-for="group in filteredGroups" :id="'city-' + group.initial" :key="group.initial" class="city-group"><h2>{{ group.initial }}</h2><button v-for="city in group.items" :key="city.adcode" class="city-item" type="button" @click="selectCity(city)">{{ city.name }}</button></section><div v-if="!filteredGroups.length" class="empty">未找到相关城市</div></main>
    <nav class="letter-index"><button v-for="letter in letters" :key="letter" type="button" @click="scrollTo(letter)">{{ letter }}</button></nav>
  </div>
</template>
<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { cities, CURRENT_CITY_CHANGED_EVENT, readCurrentCity, readSelectedCity, saveSelectedCity } from '@/data/cities'
const router = useRouter()
const keyword = ref('')
// selectedCity 优先展示用户手动选择的城市，否则展示最近一次自动定位城市。
const selectedCity = ref(readSelectedCity() || readCurrentCity())
const letters = [...new Set(cities.map(city => city.initial))]

// refreshCurrentCity 接收首页定位完成事件，避免城市页只在首次打开时读取一次旧缓存。
function refreshCurrentCity(event) {
  if (localStorage.getItem('passenger-manual-city') === '1') return
  selectedCity.value = event?.detail || readCurrentCity()
}

onMounted(() => {
  refreshCurrentCity()
  window.addEventListener(CURRENT_CITY_CHANGED_EVENT, refreshCurrentCity)
})

onBeforeUnmount(() => {
  window.removeEventListener(CURRENT_CITY_CHANGED_EVENT, refreshCurrentCity)
})
const filteredGroups = computed(() => { const text = keyword.value.trim().toLowerCase(); const filtered = text ? cities.filter(city => city.name.includes(text) || city.initial.toLowerCase() === text) : cities; return letters.map(initial => ({ initial, items: filtered.filter(city => city.initial === initial) })).filter(group => group.items.length) })
function selectCity(city) { selectedCity.value = city; localStorage.setItem('passenger-manual-city', '1'); saveSelectedCity(city); router.replace({ path: '/home', query: { city: city.adcode } }) }
function scrollTo(letter) { document.getElementById('city-' + letter)?.scrollIntoView({ behavior: 'smooth', block: 'start' }) }
</script>
<style scoped>
.city-page{min-height:100vh;background:#f7f8fa;color:#1f2937}.city-header{position:sticky;top:0;z-index:2;display:flex;align-items:center;justify-content:space-between;height:56px;padding:0 14px;background:#fff;border-bottom:1px solid #eee}.city-header h1{margin:0;font-size:18px}.icon-btn,.cancel-btn{border:0;background:none}.icon-btn{font-size:22px}.cancel-btn{color:#7c3aed;font-size:14px}.city-search{display:flex;align-items:center;gap:8px;margin:12px;padding:0 12px;height:42px;border-radius:6px;background:#fff}.city-search input{flex:1;border:0;outline:0;font-size:14px}.current-city{padding:16px;background:#fff;border-top:1px solid #f1f1f1}.city-content{padding-bottom:24px}.city-group h2{margin:0;padding:10px 14px;color:#9ca3af;font-size:13px;font-weight:500;background:#f2f3f5}.city-item{display:block;width:100%;padding:14px;border:0;border-bottom:1px solid #eee;background:#fff;text-align:left;font-size:15px}.empty{text-align:center;padding:50px;color:#9ca3af}.letter-index{position:fixed;right:5px;top:50%;display:flex;flex-direction:column;transform:translateY(-50%)}.letter-index button{padding:2px;border:0;background:none;color:#6b7280;font-size:11px}
</style>
