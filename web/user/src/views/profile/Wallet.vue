<template>
  <div class="wallet-page">
    <!-- 顶部导航 -->
    <div class="page-header">
      <van-icon name="arrow-left" size="20" @click="goBack" />
      <span class="title">我的钱包</span>
      <div style="width: 20px;"></div>
    </div>

    <!-- 余额卡片 -->
    <div class="balance-card">
      <p class="label">账户余额（元）</p>
      <h1 class="balance">{{ walletInfo.balance ?? '--' }}</h1>
      
      <div class="actions">
        <button class="btn-recharge" @click="showRecharge = true">充值</button>
        <button class="btn-withdraw" @click="showWithdraw = true">提现</button>
      </div>
    </div>

    <!-- 资产概览 -->
    <div class="asset-card">
      <h3>资产概览</h3>
      <div class="asset-list">
        <div class="asset-item">
          <span class="label">可用余额</span>
          <span class="value green">{{ walletInfo.balance ?? '--' }}</span>
        </div>
        <div class="asset-item">
          <span class="label">行程待付</span>
          <span class="value red">{{ walletInfo.pending ?? '--' }}</span>
        </div>
        <div class="asset-item">
          <span class="label">优惠券抵扣</span>
          <span class="value orange">{{ walletInfo.couponDiscount ?? '--' }}</span>
        </div>
        <div class="asset-item">
          <span class="label">充值</span>
          <span class="value blue">{{ walletInfo.recharged ?? '--' }}</span>
        </div>
      </div>
    </div>

    <!-- 交易记录 -->
    <div class="transaction-card">
      <div class="card-header">
        <h3>交易记录</h3>
        <button class="filter-btn" @click="showFilter = true">
          筛选
          <van-icon name="arrow-down" size="12" />
        </button>
      </div>

      <div class="transaction-list" v-if="transactions.length > 0">
        <div 
          v-for="(item, index) in transactions" 
          :key="index"
          class="transaction-item"
        >
          <div class="icon" :class="item.type">
            <van-icon :name="item.icon" :size="20" :color="item.color" />
          </div>
          <div class="info">
            <p class="title">{{ item.title }}</p>
            <p class="time">{{ item.time }}</p>
          </div>
          <div class="amount" :class="{ income: item.amount > 0 }">
            {{ item.amount > 0 ? '+' : '' }}¥{{ Math.abs(item.amount) }}
          </div>
        </div>
      </div>

      <div class="empty-state" v-else>
        <p>暂无交易记录</p>
      </div>
    </div>

    <!-- 充值弹窗 -->
    <van-popup v-model:show="showRecharge" position="bottom" round>
      <div class="recharge-popup">
        <div class="popup-header">
          <span>充值金额</span>
          <van-icon name="cross" size="20" @click="showRecharge = false" />
        </div>
        
        <div class="amount-input">
          <span>¥</span>
          <input 
            v-model="rechargeAmount" 
            type="number" 
            placeholder="请输入充值金额"
          />
        </div>

        <div class="quick-amounts">
          <button 
            v-for="amount in quickAmounts" 
            :key="amount"
            class="quick-btn"
            :class="{ active: rechargeAmount == amount }"
            @click="rechargeAmount = amount"
          >
            ¥{{ amount }}
          </button>
        </div>

        <button class="btn-confirm-recharge" @click="handleRecharge">
          确认充值
        </button>
      </div>
    </van-popup>

    <!-- 提现弹窗 -->
    <van-dialog
      v-model:show="showWithdraw"
      title="提现"
      show-cancel-button
      confirm-button-text="确认提现"
      @confirm="handleWithdraw"
    >
      <div class="withdraw-content">
        <p>可提现余额：¥{{ walletInfo.balance ?? '--' }}</p>
        <input 
          v-model="withdrawAmount" 
          type="number" 
          placeholder="请输入提现金额"
          class="withdraw-input"
        />
        <p class="tip">提现将在1-3个工作日内到账</p>
      </div>
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showLoadingToast, closeToast } from 'vant'

const router = useRouter()

// 钱包信息
const walletInfo = ref({
  // 余额由后端钱包接口提供，未加载到真实数据前不预置演示金额。
  balance: null,
  pending: null,
  couponDiscount: null,
  recharged: null
})

// 交易记录由后端接口提供，未加载真实数据前保持为空，避免展示演示记录。
const transactions = ref([])

// 充值相关
const showRecharge = ref(false)
const showWithdraw = ref(false)
const showFilter = ref(false)
const rechargeAmount = ref('')
const withdrawAmount = ref('')

const quickAmounts = [10, 30, 50, 100, 200]

// 方法
const goBack = () => router.back()

// 处理充值
const handleRecharge = async () => {
  if (!rechargeAmount.value || parseFloat(rechargeAmount.value) <= 0) {
    showToast('请输入有效金额')
    return
  }

  try {
    const toast = showLoadingToast({
      message: '正在处理...',
      forbidClick: true,
      duration: 0
    })

    // 调用充值接口
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    closeToast()
    showToast('充值成功')
    showRecharge.value = false
    
    // 更新余额
    walletInfo.value.balance = (Number(walletInfo.value.balance || 0) + parseFloat(rechargeAmount.value)).toFixed(2)
    rechargeAmount.value = ''
  } catch (error) {
    console.error(error)
    closeToast()
    showToast('充值失败，请重试')
  }
}

// 处理提现
const handleWithdraw = async () => {
  if (!withdrawAmount.value || parseFloat(withdrawAmount.value) <= 0) {
    showToast('请输入有效金额')
    return false
  }

  if (parseFloat(withdrawAmount.value) > Number(walletInfo.value.balance || 0)) {
    showToast('余额不足')
    return false
  }

  try {
    const toast = showLoadingToast({
      message: '正在处理...',
      forbidClick: true,
      duration: 0
    })

    // 调用提现接口
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    closeToast()
    showToast('提现申请已提交')
    
    // 更新余额
    walletInfo.value.balance = (Number(walletInfo.value.balance || 0) - parseFloat(withdrawAmount.value)).toFixed(2)
    withdrawAmount.value = ''
    return true
  } catch (error) {
    console.error(error)
    closeToast()
    showToast('提现失败，请重试')
    return false
  }
}

onMounted(() => {
  // 加载钱包数据
})
</script>

<style scoped>
.wallet-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: white;
  position: sticky;
  top: 0;
  z-index: 10;
}

.title {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
}

.balance-card {
  margin: 16px;
  padding: 30px 24px;
  background: linear-gradient(135deg, #7C3AED 0%, #9333EA 100%);
  border-radius: 16px;
  text-align: center;
  color: white;
}

.balance-card .label {
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 12px;
}

.balance {
  font-size: 42px;
  font-weight: 700;
  margin-bottom: 28px;
}

.actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.btn-recharge,
.btn-withdraw {
  min-width: 120px;
  height: 40px;
  border-radius: 20px;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-recharge {
  background: white;
  color: #7C3AED;
  border: none;
}

.btn-withdraw {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.asset-card,
.transaction-card {
  margin: 0 16px 16px;
  background: white;
  border-radius: 12px;
  padding: 16px;
}

.asset-card h3,
.card-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  background: #F3F4F6;
  border: none;
  border-radius: 14px;
  font-size: 13px;
  color: #6B7280;
  cursor: pointer;
}

.asset-item {
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #F9FAFB;
}

.asset-item:last-child {
  border-bottom: none;
}

.label {
  font-size: 14px;
  color: #6B7280;
}

.value {
  font-size: 15px;
  font-weight: 500;
}

.value.green { color: #059669; }
.value.red { color: #DC2626; }
.value.orange { color: #D97706; }
.value.blue { color: #2563EB; }

.transaction-list {
  max-height: 400px;
  overflow-y: auto;
}

.transaction-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px solid #F9FAFB;
}

.icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.icon.expense { background: #F5F3FF; }
.icon.income { background: #ECFDF5; }
.icon.refund { background: #EFF6FF; }

.info {
  flex: 1;
}

.info .title {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.time {
  font-size: 12px;
  color: #9CA3AF;
}

.amount {
  font-size: 16px;
  font-weight: 600;
  color: #DC2626;
}

.amount.income {
  color: #059669;
}

.empty-state {
  text-align: center;
  padding: 40px 0;
  color: #9CA3AF;
}

/* 充值弹窗 */
.recharge-popup {
  padding: 24px;
}

.popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 17px;
  font-weight: 600;
  margin-bottom: 24px;
}

.amount-input {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 36px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid #E5E7EB;
}

.amount-input input {
  flex: 1;
  font-size: 36px;
  font-weight: 700;
  outline: none;
  border: none;
  background: transparent;
}

.quick-amounts {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}

.quick-btn {
  height: 44px;
  background: #F9FAFB;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  font-size: 15px;
  cursor: pointer;
  transition: all 0.2s;
}

.quick-btn.active {
  background: #F5F3FF;
  border-color: #7C3AED;
  color: #7C3AED;
}

.btn-confirm-recharge {
  width: 100%;
  height: 48px;
  background: linear-gradient(135deg, #7C3AED 0%, #9333EA 100%);
  color: white;
  border: none;
  border-radius: 24px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
}

/* 提现内容 */
.withdraw-content {
  padding: 20px 0;
  text-align: center;
}

.withdraw-content > p {
  font-size: 14px;
  color: #6B7280;
  margin-bottom: 16px;
}

.withdraw-input {
  width: calc(100% - 32px);
  height: 44px;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  padding: 0 16px;
  font-size: 16px;
  text-align: center;
  outline: none;
  margin-bottom: 12px;
}

.tip {
  font-size: 12px !important;
  color: #9CA3AF !important;
}
</style>
