<template>
  <div class="coupon-page">
    <header class="page-header">
      <van-icon name="arrow-left" size="20" @click="goBack" />
      <h1>我的优惠券</h1>
      <div class="header-placeholder"></div>
    </header>

    <van-tabs v-model:active="activeStatus" color="#7C3AED" @change="loadCoupons">
      <van-tab v-for="tab in tabs" :key="tab.value" :title="tab.label" :name="tab.value" />
    </van-tabs>

    <main class="coupon-content">
      <van-loading v-if="loading" vertical>加载中...</van-loading>
      <div v-else-if="coupons.length" class="coupon-list">
        <article v-for="coupon in coupons" :key="coupon.userCouponId" class="coupon-item">
          <div class="coupon-value">
            <strong>¥{{ amount(coupon) }}</strong>
            <span>{{ coupon.type === 2 ? '折扣券' : '立减券' }}</span>
          </div>
          <div class="coupon-info">
            <h2>{{ coupon.name }}</h2>
            <p>满{{ money(coupon.thresholdCents) }}元可用</p>
            <p>有效期至 {{ expireDate(coupon.expireAt) }}</p>
          </div>
          <span class="coupon-status" :class="statusClass(coupon.status)">{{ statusText(coupon.status) }}</span>
        </article>
      </div>
      <van-empty v-else description="暂无优惠券" />
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getMyCoupons } from '@/api/order'
import { showToast } from 'vant'

const router = useRouter()
const activeStatus = ref(1)
const coupons = ref([])
const loading = ref(false)
const tabs = [
  { label: '可用', value: 1 },
  { label: '已使用', value: 2 },
  { label: '已过期', value: 3 },
  { label: '全部', value: 0 }
]

// 按用户券状态加载当前登录用户的真实优惠券列表。
const loadCoupons = async () => {
  loading.value = true
  try {
    const result = await getMyCoupons(activeStatus.value)
    coupons.value = Array.isArray(result?.list) ? result.list : []
  } catch (error) {
    coupons.value = []
    showToast('优惠券加载失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

const amount = coupon => (Number(coupon.faceValueCents || 0) / 100).toFixed(2)
const money = cents => (Number(cents || 0) / 100).toFixed(2)
const expireDate = timestamp => timestamp ? new Date(Number(timestamp) * 1000).toLocaleDateString() : '长期有效'
const statusText = status => ({ 1: '可使用', 2: '已使用', 3: '已过期', 4: '使用中' }[status] || '不可用')
const statusClass = status => `status-${status}`
const goBack = () => router.back()

onMounted(loadCoupons)
</script>

<style scoped>
.coupon-page { min-height: 100vh; background: #f5f5f5; }
.page-header { height: 52px; display: flex; align-items: center; justify-content: space-between; padding: 0 16px; background: #fff; position: sticky; top: 0; z-index: 2; }
.page-header h1 { margin: 0; color: #111827; font-size: 17px; font-weight: 600; }
.header-placeholder { width: 20px; }
.coupon-content { padding: 14px 16px 28px; }
.coupon-list { display: flex; flex-direction: column; gap: 12px; }
.coupon-item { min-height: 104px; display: flex; align-items: center; gap: 12px; padding: 14px; border-radius: 10px; background: #fff; box-shadow: 0 2px 8px rgba(15, 23, 42, .04); }
.coupon-value { width: 82px; flex-shrink: 0; text-align: center; color: #ef4444; }
.coupon-value strong { display: block; font-size: 24px; line-height: 1.2; }
.coupon-value span { font-size: 12px; color: #f97316; }
.coupon-info { min-width: 0; flex: 1; }
.coupon-info h2 { margin: 0 0 8px; overflow: hidden; color: #111827; font-size: 15px; text-overflow: ellipsis; white-space: nowrap; }
.coupon-info p { margin: 4px 0; color: #9ca3af; font-size: 12px; }
.coupon-status { flex-shrink: 0; font-size: 12px; }
.status-1 { color: #7c3aed; }
.status-2, .status-3, .status-4 { color: #9ca3af; }
</style>
