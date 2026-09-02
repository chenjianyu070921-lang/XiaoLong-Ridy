<template>
  <main class="mine-page-shell">
    <header class="mine-page-header">
      <button type="button" class="page-back" aria-label="返回" @click="goHome">
        <van-icon name="arrow-left" />
      </button>
      <div class="page-heading">
        <p>我的</p>
        <h1>钱包提现</h1>
      </div>
      <button type="button" class="page-action" @click="loadIncome">
        <van-icon name="replay" />
      </button>
    </header>

    <DriverWalletPanel
      :income-summary="incomeSummary"
      :today-income="todayIncome"
      :week-income="weekIncome"
      :income-bills="incomeBills"
      :format-price="formatPrice"
      :format-time="formatTime"
      @load-income="loadIncome"
      @open-withdraw="openWithdraw"
    />

    <van-popup v-model:show="withdrawVisible" round position="bottom" teleport="body">
      <section class="page-sheet">
        <h2>申请提现</h2>
        <van-form @submit="submitWithdraw">
          <van-field v-model="withdrawForm.amount" type="number" label="提现金额" placeholder="请输入金额" />
          <van-field v-model="withdrawForm.payeeName" label="收款人" placeholder="请输入收款人姓名" />
          <van-field v-model="withdrawForm.payAccount" label="收款账号" placeholder="请输入收款账号" />
          <button class="primary-action" type="submit" :disabled="withdrawLoading">
            {{ withdrawLoading ? '提交中...' : '确认提现' }}
          </button>
        </van-form>
      </section>
    </van-popup>
  </main>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import DriverWalletPanel from '@/components/driver-home/DriverWalletPanel.vue'
import { useDriverAssets } from '@/composables/useDriverAssets'
import { formatPrice, formatTime } from '@/utils/driver-format'
import '@/styles/driver-home-panels.css'
import '@/styles/driver-mine-pages.css'

const router = useRouter()
const {
  incomeSummary,
  todayIncome,
  weekIncome,
  incomeBills,
  withdrawVisible,
  withdrawLoading,
  withdrawForm,
  loadIncome,
  openWithdraw,
  submitWithdraw
} = useDriverAssets()

onMounted(() => {
  void loadIncome({ silentError: true })
})

function goHome() {
  if (window.history.length > 1) router.back()
  else router.replace('/home')
}
</script>
