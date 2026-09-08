<template>
  <main class="mine-page-shell">
    <header class="mine-page-header">
      <button type="button" class="page-back" aria-label="返回" @click="goBack">
        <van-icon name="arrow-left" />
      </button>
      <div class="page-heading">
        <p>我的</p>
        <h1>绑定银行卡</h1>
      </div>
    </header>

    <section class="page-stack">
      <div class="compliance-banner">
        <van-icon name="warning-o" />
        <div>
          <strong>无需输入银行卡取款密码</strong>
          <p>绑卡全程不需要、也不能输入银行卡取款密码（取钱用的密码）。平台绝不会索要该密码，谨防诈骗。</p>
        </div>
      </div>

      <van-form class="bind-card-form" @submit="submitBind">
        <van-field v-model="form.bankName" label="开户行" placeholder="请输入开户行名称" />
        <van-field v-model="form.cardNo" type="tel" label="银行卡号" placeholder="请输入银行卡号" />
        <van-field v-model="form.holderName" label="持卡人姓名" placeholder="须与司机实名一致" />
        <van-field v-model="form.holderIdCard" label="身份证号" placeholder="须与司机实名一致" />
        <van-field v-model="form.reservedPhone" type="tel" label="预留手机号" placeholder="请输入银行预留手机号">
          <template #button>
            <van-button
              size="small"
              type="primary"
              native-type="button"
              :disabled="countdown > 0 || !validReservedPhone"
              @click="sendCode"
            >
              {{ countdown > 0 ? countdown + 's' : '获取验证码' }}
            </van-button>
          </template>
        </van-field>
        <van-field v-model="form.smsCode" type="tel" label="短信验证码" placeholder="请输入短信验证码" />
        <button class="primary-action" type="submit" :disabled="submitting">
          {{ submitting ? '提交中...' : '确认绑定' }}
        </button>
      </van-form>

      <p class="bind-tip">绑卡成功后，平台将自动生成 6 位提现密码并发送至您的注册手机号，用于后续提现操作确认。</p>
    </section>
  </main>
</template>

<script setup>
import { computed, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { bindBankCard, sendBankCardSmsCode } from '@/api/driver'
import { apiErrorMessage, safeApiCall } from '@/utils/safe-request'
import '@/styles/driver-mine-pages.css'

const router = useRouter()
const form = reactive({
  bankName: '',
  cardNo: '',
  holderName: '',
  holderIdCard: '',
  reservedPhone: '',
  smsCode: ''
})
const submitting = ref(false)
const countdown = ref(0)
let timer = null

const validReservedPhone = computed(() => /^1[3-9]\d{9}$/.test(String(form.reservedPhone || '').trim()))

onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})

async function sendCode() {
  if (!validReservedPhone.value) {
    showToast('请输入正确的银行预留手机号')
    return
  }
  try {
    await sendBankCardSmsCode(String(form.reservedPhone).trim(), { silentError: true })
    showToast('验证码已发送（联调验证码见服务端日志）')
  } catch (error) {
    showToast(apiErrorMessage(error, '验证码发送失败'))
    return
  }
  countdown.value = 60
  if (timer) window.clearInterval(timer)
  timer = window.setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) {
      window.clearInterval(timer)
      timer = null
    }
  }, 1000)
}

async function submitBind() {
  if (!form.bankName.trim() || !form.cardNo.trim() || !form.holderName.trim() || !form.holderIdCard.trim() || !form.reservedPhone.trim() || !form.smsCode.trim()) {
    showToast('请填写完整的绑卡信息')
    return
  }
  if (!validReservedPhone.value) {
    showToast('请输入正确的银行预留手机号')
    return
  }
  submitting.value = true
  try {
    const res = await safeApiCall(() => bindBankCard({
      bankName: form.bankName.trim(),
      cardNo: form.cardNo.trim(),
      holderName: form.holderName.trim(),
      holderIdCard: form.holderIdCard.trim(),
      reservedPhone: form.reservedPhone.trim(),
      smsCode: form.smsCode.trim()
    }))
    if (!res) return
    showToast('绑卡成功，提现密码已发送至注册手机号')
    router.replace('/mine/bank-cards')
  } finally {
    submitting.value = false
  }
}

function goBack() {
  router.back()
}
</script>

<style scoped>
.compliance-banner {
  display: flex;
  gap: 10px;
  padding: 12px;
  border-radius: 10px;
  background: #fff4e5;
  color: #92610b;
  font-size: 12px;
  line-height: 1.5;
}

.compliance-banner > .van-icon {
  flex: 0 0 auto;
  margin-top: 1px;
  font-size: 16px;
}

.compliance-banner strong {
  font-size: 13px;
}

.compliance-banner p {
  margin: 2px 0 0;
}

.bind-card-form {
  padding: 4px 0;
}

.bind-tip {
  margin: 8px 2px 0;
  color: var(--driver-faint);
  font-size: 12px;
  line-height: 1.6;
}
</style>
