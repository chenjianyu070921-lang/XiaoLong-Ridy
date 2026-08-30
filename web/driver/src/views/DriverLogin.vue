<template>
  <main class="driver-login-page">
    <section class="login-hero">
      <img src="/logo.png" alt="��С��" />
      <div>
        <p>��С��˾����</p>
        <h1>�ӵ�����̨</h1>
      </div>
    </section>

    <section class="login-card">
      <div class="auth-mode-tabs">
        <button
          v-for="item in authTabs"
          :key="item.value"
          type="button"
          :class="{ active: activeTab === item.value }"
          @click="activeTab = item.value"
        >
          {{ item.label }}
        </button>
      </div>

      <van-form v-if="activeTab === 0" class="auth-form" @submit="handlePasswordLogin">
        <van-field v-model="passwordForm.phone" name="phone" type="tel" label="�ֻ���" placeholder="�������ֻ���" clearable />
        <van-field v-model="passwordForm.password" name="password" type="password" label="����" placeholder="����������" clearable />
        <button class="primary-action" :disabled="loading" type="submit">
          {{ loading ? '��¼��...' : '�����¼' }}
        </button>
      </van-form>

      <van-form v-else-if="activeTab === 1" class="auth-form" @submit="handleSMSLogin">
        <van-field v-model="smsForm.phone" name="phone" type="tel" label="�ֻ���" placeholder="�������ֻ���" clearable>
          <template #button>
            <van-button size="small" type="primary" native-type="button" :disabled="smsCountdown > 0" @click="sendCode">
              {{ smsCountdown > 0 ? smsCountdown + 's' : '����' }}
            </van-button>
          </template>
        </van-field>
        <van-field v-model="smsForm.code" name="code" type="tel" label="��֤��" placeholder="��������֤��" clearable />
        <button class="primary-action" :disabled="loading" type="submit">
          {{ loading ? '��¼��...' : '��֤���¼' }}
        </button>
      </van-form>

      <van-form v-else class="auth-form" @submit="handleRegister">
        <van-field v-model="registerForm.phone" name="phone" type="tel" label="�ֻ���" placeholder="�������ֻ���" clearable />
        <van-field v-model="registerForm.realName" name="realName" label="����" placeholder="��������ʵ����" clearable />
        <van-field v-model="registerForm.idCardNo" name="idCardNo" label="����֤��" placeholder="����������֤��" clearable />
        <van-field v-model="registerForm.driverLicenseNo" name="driverLicenseNo" label="��ʻ֤��" placeholder="�������ʻ֤��" clearable />
        <van-field v-model="registerForm.avatarUrl" name="avatarUrl" label="ͷ���ַ" placeholder="��ѡ" clearable />
        <van-field v-model="registerForm.password" name="password" type="password" label="����" placeholder="���õ�¼����" clearable />
        <button class="primary-action" :disabled="loading" type="submit">
          {{ loading ? '�ύ��...' : '�ύע��' }}
        </button>
      </van-form>
    </section>

    <SmsCodeDialog v-model:show="smsCodeDialogVisible" :phone="smsForm.phone" :code="smsCodeValue" />
  </main>
</template>

<script setup>
import { onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { closeToast, showLoadingToast, showToast } from 'vant'
import { sendDriverSMSCode } from '@/api/driver'
import SmsCodeDialog from '@/components/SmsCodeDialog.vue'
import { useDriverStore } from '@/stores/driver'

const router = useRouter()
const driverStore = useDriverStore()

const authTabs = [
  { label: '����', value: 0 },
  { label: '��֤��', value: 1 },
  { label: 'ע��', value: 2 }
]

const activeTab = ref(0)
const loading = ref(false)
const smsCountdown = ref(0)
const smsCodeDialogVisible = ref(false)
const smsCodeValue = ref('')
let smsTimer = null

const driverPhoneRegexp = /^(?:1[3-9]\d{9}|\d{12,15})$/

const passwordForm = reactive({ phone: '', password: '' })
const smsForm = reactive({ phone: '', code: '' })
const registerForm = reactive({
  phone: '',
  realName: '',
  idCardNo: '',
  driverLicenseNo: '',
  avatarUrl: '',
  password: ''
})

onUnmounted(() => {
  if (smsTimer) window.clearInterval(smsTimer)
})

function normalizePhone(phone) {
  return String(phone || '').trim()
}

function validatePhone(phone) {
  return driverPhoneRegexp.test(normalizePhone(phone))
}

async function sendCode() {
  const phone = normalizePhone(smsForm.phone)
  if (!validatePhone(phone)) {
    showToast('��������ȷ���ֻ���')
    return
  }
  try {
    const res = await sendDriverSMSCode(phone, { silentError: true })
    smsCodeValue.value = String(res?.code || '').trim()
    smsCodeDialogVisible.value = true
    showToast('��֤���ѷ���')
  } catch (error) {
    showToast(apiErrorMessage(error, '��֤�뷢��ʧ��'))
    return
  }

  smsCountdown.value = 60
  if (smsTimer) window.clearInterval(smsTimer)
  smsTimer = window.setInterval(() => {
    smsCountdown.value -= 1
    if (smsCountdown.value <= 0) {
      window.clearInterval(smsTimer)
      smsTimer = null
    }
  }, 1000)
}

async function handlePasswordLogin() {
  const phone = normalizePhone(passwordForm.phone)
  if (!validatePhone(phone) || !passwordForm.password) {
    showToast('�������ֻ��ź�����')
    return
  }
  await submitLogin(() => driverStore.loginPassword(phone, passwordForm.password, { silentError: true }))
}

async function handleSMSLogin() {
  const phone = normalizePhone(smsForm.phone)
  if (!validatePhone(phone) || !smsForm.code) {
    showToast('�������ֻ��ź���֤��')
    return
  }
  await submitLogin(() => driverStore.loginSMS(phone, smsForm.code, { silentError: true }))
}

async function submitLogin(action) {
  try {
    loading.value = true
    showLoadingToast({ message: '��¼��...', forbidClick: true, duration: 0 })
    await action()
    closeToast()
    showToast('��¼�ɹ�')
    router.replace('/home')
  } catch (error) {
    closeToast()
    showToast(apiErrorMessage(error, '��¼ʧ��'))
  } finally {
    loading.value = false
  }
}

async function handleRegister() {
  const phone = normalizePhone(registerForm.phone)
  if (!validatePhone(phone) || !registerForm.realName || !registerForm.idCardNo || !registerForm.driverLicenseNo || !registerForm.password) {
    showToast('����д����ע����Ϣ')
    return
  }

  try {
    loading.value = true
    showLoadingToast({ message: '�ύ��...', forbidClick: true, duration: 0 })
    await driverStore.register({ ...registerForm, phone }, { silentError: true })
    closeToast()
    showToast('ע��ɹ�����ȴ��������')
    activeTab.value = 0
  } catch (error) {
    closeToast()
    showToast(apiErrorMessage(error, 'ע��ʧ��'))
  } finally {
    loading.value = false
  }
}

function apiErrorMessage(error, fallbackMessage) {
  return error?.response?.data?.message || error?.message || fallbackMessage
}
</script>

<style scoped>
.driver-login-page {
  min-height: 100vh;
  padding: 18px 12px 24px;
  background: #eef2f7;
}

.login-hero {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 148px;
  padding: 22px 16px;
  border-radius: 0 0 22px 22px;
  background: linear-gradient(135deg, #6d4aff, #4b2bc5);
  color: #fff;
}

.login-hero img {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  background: #fff;
  object-fit: cover;
}

.login-hero p {
  margin: 0 0 5px;
  color: rgba(255, 255, 255, .78);
  font-size: 13px;
}

.login-hero h1 {
  margin: 0;
  font-size: 25px;
  line-height: 1.15;
  letter-spacing: 0;
}

.login-card {
  margin-top: -22px;
  padding: 12px;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 14px 34px rgba(77, 48, 160, .16);
}

.auth-mode-tabs {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
  padding: 4px;
  border-radius: 10px;
  background: #f1f5fb;
}

.auth-mode-tabs button {
  min-height: 38px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: #667085;
  font-size: 14px;
  font-weight: 800;
}

.auth-mode-tabs button.active {
  background: #6d4aff;
  color: #fff;
  box-shadow: 0 6px 14px rgba(109, 74, 255, .22);
}

.auth-form {
  padding-top: 12px;
}

.driver-login-page :deep(.van-cell) {
  padding: 12px 4px;
}

.driver-login-page :deep(.van-field__label) {
  width: 72px;
  color: #667085;
  font-size: 13px;
}

.primary-action {
  width: 100%;
  min-height: 48px;
  margin-top: 16px;
  border: 0;
  border-radius: 10px;
  background: #6d4aff;
  color: #fff;
  font-size: 16px;
  font-weight: 800;
}

.primary-action:disabled {
  opacity: .62;
}
</style>
