<template>
  <main class="mine-page-shell">
    <header class="mine-page-header">
      <button type="button" class="page-back" aria-label="返回" @click="goBack">
        <van-icon name="arrow-left" />
      </button>
      <div class="page-heading">
        <p>我的</p>
        <h1>银行卡</h1>
      </div>
      <button type="button" class="page-action" @click="loadCards">
        <van-icon name="plus" />
      </button>
    </header>

    <section class="page-stack">
      <div class="compliance-banner">
        <van-icon name="warning-o" />
        <span>平台不会索要银行卡取款密码（取钱用的密码），如遇索要请勿提供</span>
      </div>

      <div v-if="cards.length === 0" class="empty-state bank-empty">
        <p>暂未绑定银行卡</p>
        <button type="button" class="primary-action bank-empty-action" @click="goBind">去绑定银行卡</button>
      </div>

      <article v-for="card in cards" :key="card.id" class="bank-card-item">
        <div class="bank-card-icon">
          <van-icon name="bank-o" />
        </div>
        <div class="bank-card-main">
          <div class="bank-card-name">{{ card.bankName }}</div>
          <div class="bank-card-no">{{ card.maskedCardNo }}</div>
          <div class="bank-card-meta">
            <span>持卡人 {{ card.holderName }}</span>
            <span>预留 {{ maskPhone(card.reservedPhone) }}</span>
          </div>
        </div>
        <button type="button" class="bank-card-delete" @click="removeCard(card)">删除</button>
      </article>

      <button v-if="cards.length > 0 && cards.length < 5" type="button" class="bank-add-btn" @click="goBind">
        <van-icon name="plus" />
        <span>添加银行卡（{{ cards.length }}/5）</span>
      </button>
      <p v-if="cards.length >= 5" class="bank-limit-tip">最多绑定 5 张银行卡，如需绑定新卡请先删除已有卡片</p>
    </section>
  </main>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showConfirmDialog, showToast } from 'vant'
import { deleteBankCard, listBankCards } from '@/api/driver'
import { apiErrorMessage, safeApiCall } from '@/utils/safe-request'
import '@/styles/driver-mine-pages.css'

const router = useRouter()
const cards = ref([])

onMounted(() => {
  void loadCards({ silentError: true })
})

async function loadCards(config = {}) {
  const res = await safeApiCall(() => listBankCards(config))
  if (res && Array.isArray(res.cards)) {
    cards.value = res.cards
  }
  return res
}

async function removeCard(card) {
  try {
    await showConfirmDialog({
      title: '删除银行卡',
      message: `确认删除 ${card.bankName}（${card.maskedCardNo}）？`
    })
  } catch {
    return
  }
  const res = await safeApiCall(() => deleteBankCard(card.id))
  if (!res) return
  showToast('银行卡已删除')
  await loadCards({ silentError: true })
}

function goBack() {
  router.back()
}

function goBind() {
  router.push('/mine/bank-cards/bind')
}

function maskPhone(phone) {
  const value = String(phone || '')
  if (value.length < 7) return value
  return value.slice(0, 3) + '****' + value.slice(-4)
}
</script>

<style scoped>
.compliance-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 10px;
  background: var(--driver-soft);
  color: var(--driver-muted);
  font-size: 12px;
  line-height: 1.5;
}

.bank-empty {
  padding: 48px 0;
  text-align: center;
}

.bank-empty-action {
  width: auto;
  min-width: 180px;
  margin: 16px auto 0;
}

.bank-card-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border-radius: 12px;
  background: var(--driver-card);
  box-shadow: 0 3px 12px rgba(15, 23, 42, .05);
}

.bank-card-icon {
  display: grid;
  width: 40px;
  height: 40px;
  flex: 0 0 40px;
  place-items: center;
  border-radius: 10px;
  background: var(--driver-soft);
  color: var(--driver-primary);
  font-size: 20px;
}

.bank-card-main {
  min-width: 0;
  flex: 1;
}

.bank-card-name {
  color: var(--driver-ink);
  font-size: 15px;
  font-weight: 700;
}

.bank-card-no {
  margin-top: 2px;
  color: var(--driver-muted);
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 1px;
}

.bank-card-meta {
  display: flex;
  gap: 12px;
  margin-top: 2px;
  color: var(--driver-faint);
  font-size: 12px;
}

.bank-card-delete {
  border: 0;
  padding: 6px 10px;
  border-radius: 8px;
  background: transparent;
  color: #ef4444;
  font-size: 13px;
}

.bank-add-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100%;
  min-height: 46px;
  border: 1px dashed var(--driver-line);
  border-radius: 12px;
  background: var(--driver-card);
  color: var(--driver-primary);
  font-size: 14px;
}

.bank-limit-tip {
  margin: 0;
  color: var(--driver-faint);
  font-size: 12px;
  text-align: center;
}
</style>
