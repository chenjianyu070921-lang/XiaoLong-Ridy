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

    <section class="page-stack bank-entry-section">
      <button type="button" class="bank-entry" @click="openBankCards">
        <div class="bank-entry-icon"><van-icon name="bank-o" /></div>
        <div class="bank-entry-text">
          <strong>银行卡管理</strong>
          <span>最多绑定 5 张，提现使用平台提现密码确认</span>
        </div>
        <van-icon name="arrow" class="bank-entry-arrow" />
      </button>
    </section>

    <section class="page-stack withdraw-record-section">
      <div class="detail-section-head">
        <h2>提现记录</h2>
        <span>{{ withdrawRecords.length }} 条</span>
      </div>
      <div v-if="withdrawRecords.length === 0" class="empty-state">暂无提现记录</div>
      <article v-for="record in withdrawRecords" :key="record.id || record.withdrawNo" class="record-card withdraw-record-card">
        <div class="record-head">
          <strong>{{ record.withdrawNo || '提现申请' }}</strong>
          <span class="withdraw-status" :class="'status-' + Number(record.status || 0)">{{ formatWithdrawStatus(record.status) }}</span>
        </div>
        <div class="record-meta">
          <span>{{ formatPrice(record.amountCents) }}</span>
          <span>{{ formatTime(record.appliedAt || record.createdAt) }}</span>
        </div>
        <p v-if="record.remark" class="withdraw-remark">{{ record.remark }}</p>
      </article>
    </section>

    <van-popup v-model:show="withdrawVisible" round position="bottom" teleport="body">
      <section class="page-sheet">
        <h2>申请提现</h2>
        <van-form @submit="submitWithdraw">
          <van-field v-model="withdrawForm.amount" type="number" label="提现金额" placeholder="请输入金额" />
          <van-field
            v-model="withdrawForm.bankCardId"
            is-link
            readonly
            label="收款银行卡"
            :placeholder="bankCards.length ? '请选择银行卡' : '暂未绑定银行卡'"
            @click="pickBankCard"
          >
            <template v-if="selectedBankCard" #input>
              <span class="bank-card-selected">{{ selectedBankCard.bankName }}（{{ selectedBankCard.maskedCardNo }}）</span>
            </template>
          </van-field>
          <van-field v-model="withdrawForm.withdrawPassword" type="password" label="平台提现密码" placeholder="请输入6位提现密码" />
          <div class="withdraw-extra">
            <button type="button" class="link-btn" @click="openBankCards">银行卡管理</button>
            <button type="button" class="link-btn" @click="openResetPassword">忘记提现密码？</button>
          </div>
          <button class="primary-action" type="submit" :disabled="withdrawLoading">
            {{ withdrawLoading ? '提交中...' : '确认提现' }}
          </button>
        </van-form>
      </section>
    </van-popup>

    <!-- 忘记提现密码：手机号 + 身份证 + 真实姓名 三重校验找回 -->
    <van-popup v-model:show="resetVisible" round position="bottom" teleport="body">
      <section class="page-sheet">
        <h2>找回提现密码</h2>
        <p class="sheet-tip">验证通过后将重置新的提现密码并发送至注册手机号</p>
        <van-form @submit="submitResetPassword">
          <van-field v-model="resetForm.phone" type="tel" label="手机号" placeholder="请输入注册手机号" />
          <van-field v-model="resetForm.idCardNo" label="身份证号" placeholder="请输入身份证号" />
          <van-field v-model="resetForm.realName" label="真实姓名" placeholder="请输入真实姓名" />
          <button class="primary-action" type="submit" :disabled="resetLoading">
            {{ resetLoading ? '验证中...' : '验证并重置' }}
          </button>
        </van-form>
      </section>
    </van-popup>

    <!-- 选择银行卡 -->
    <van-action-sheet v-model:show="pickVisible" :actions="bankCardActions" cancel-text="取消" @select="onPickCard" />
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import DriverWalletPanel from '@/components/driver-home/DriverWalletPanel.vue'
import { resetWithdrawPassword } from '@/api/driver'
import { useDriverAssets } from '@/composables/useDriverAssets'
import { apiErrorMessage, safeApiCall } from '@/utils/safe-request'
import { formatPrice, formatTime } from '@/utils/driver-format'
import '@/styles/driver-home-panels.css'
import '@/styles/driver-mine-pages.css'

const router = useRouter()
const {
  incomeSummary,
  todayIncome,
  weekIncome,
  incomeBills,
  withdrawRecords,
  withdrawVisible,
  withdrawLoading,
  withdrawForm,
  bankCards,
  loadBankCards,
  loadIncome,
  openWithdraw,
  submitWithdraw
} = useDriverAssets()

const resetVisible = ref(false)
const resetLoading = ref(false)
const resetForm = reactive({ phone: '', idCardNo: '', realName: '' })
const pickVisible = ref(false)

const selectedBankCard = computed(() => bankCards.value.find((card) => String(card.id) === String(withdrawForm.bankCardId)) || null)

const bankCardActions = computed(() =>
  bankCards.value.map((card) => ({
    name: `${card.bankName}（${card.maskedCardNo}）`,
    id: card.id
  }))
)

onMounted(() => {
  void loadIncome({ silentError: true })
  void loadBankCards({ silentError: true })
})

function goHome() {
  router.back()
}

function pickBankCard() {
  if (bankCards.value.length === 0) {
    openBankCards()
    return
  }
  pickVisible.value = true
}

function onPickCard(action) {
  withdrawForm.bankCardId = String(action.id)
  pickVisible.value = false
}

function openBankCards() {
  router.push('/mine/bank-cards')
}

async function openResetPassword() {
  resetVisible.value = true
}

async function submitResetPassword() {
  if (!resetForm.phone.trim() || !resetForm.idCardNo.trim() || !resetForm.realName.trim()) {
    showToast('请填写完整的验证信息')
    return
  }
  resetLoading.value = true
  try {
    const res = await safeApiCall(() => resetWithdrawPassword({
      phone: resetForm.phone.trim(),
      idCardNo: resetForm.idCardNo.trim(),
      realName: resetForm.realName.trim()
    }))
    if (!res) return
    resetVisible.value = false
    Object.assign(resetForm, { phone: '', idCardNo: '', realName: '' })
    showToast(res.message || '新提现密码已发送至注册手机号')
  } finally {
    resetLoading.value = false
  }
}

// 提现金额单位（后端为「元」）已在 useDriverAssets 的数据入口归一化为 amountCents，此处直接按「分」展示。
function formatWithdrawStatus(status) {
  return {
    1: '申请中',
    2: '打款成功',
    3: '打款失败'
  }[Number(status || 0)] || '--'
}
</script>

<style scoped>
.bank-card-selected {
  color: var(--driver-ink);
  font-size: 14px;
}

.withdraw-extra {
  display: flex;
  justify-content: space-between;
  padding: 6px 2px 0;
}

.link-btn {
  border: 0;
  background: transparent;
  padding: 4px 0;
  color: var(--driver-primary);
  font-size: 13px;
}

.sheet-tip {
  margin: 0 0 12px;
  color: var(--driver-muted);
  font-size: 12px;
}

.bank-entry-section {
  margin-top: 0;
}

.bank-entry {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 14px;
  border: 0;
  border-radius: 12px;
  background: var(--driver-card);
  box-shadow: 0 3px 12px rgba(15, 23, 42, .05);
  text-align: left;
}

.bank-entry-icon {
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

.bank-entry-text {
  display: grid;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.bank-entry-text strong {
  color: var(--driver-ink);
  font-size: 15px;
}

.bank-entry-text span {
  color: var(--driver-faint);
  font-size: 12px;
}

.bank-entry-arrow {
  color: var(--driver-faint);
  font-size: 14px;
}
</style>
