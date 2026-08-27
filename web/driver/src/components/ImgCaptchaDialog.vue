<template>
  <van-popup
    v-model:show="visible"
    class="img-captcha-popup"
    round
    teleport="body"
    :close-on-click-overlay="false"
    @open="loadCaptcha"
  >
    <section class="captcha-card">
      <button class="captcha-close" type="button" aria-label="关闭" @click="handleClose">×</button>
      <h2>请输入下方图形验证码</h2>

      <div class="captcha-row">
        <van-field
          v-model.trim="userInputCode"
          class="captcha-input"
          placeholder="请输入验证码"
          maxlength="8"
          clearable
        />
        <div class="captcha-image-wrap">
          <img v-if="imgBase64" class="captcha-image" :src="captchaImageSrc" alt="图形验证码" />
          <div v-else class="captcha-image empty">{{ loading ? '加载中' : '加载失败' }}</div>
          <button class="refresh-link" type="button" :disabled="loading" @click="refreshCaptcha">看不清？</button>
        </div>
      </div>

      <button class="confirm-button" type="button" :disabled="submitting || loading" @click="handleConfirm">
        确定
      </button>
    </section>
  </van-popup>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { showToast } from 'vant'
import { getImgCaptcha, invalidateImgCaptcha } from '@/api/driver'

type CaptchaConfirmPayload = {
  phone: string
  uuid: string
  userInputCode: string
}

type CaptchaClosePayload = {
  phone: string
  uuid: string
}

const props = defineProps<{
  phone: string
}>()

const emit = defineEmits<{
  confirm: [payload: CaptchaConfirmPayload]
  close: [payload: CaptchaClosePayload]
}>()

const visible = defineModel<boolean>('show', { default: false })

const uuid = ref('')
const imgBase64 = ref('')
const userInputCode = ref('')
const loading = ref(false)
const submitting = ref(false)

const captchaImageSrc = computed(() => {
  if (!imgBase64.value) return ''
  return imgBase64.value.startsWith('data:image') ? imgBase64.value : `data:image/png;base64,${imgBase64.value}`
})

watch(visible, (next) => {
  if (next) {
    userInputCode.value = ''
  }
})

async function loadCaptcha() {
  if (!props.phone) {
    showToast('请先输入手机号')
    visible.value = false
    return
  }
  loading.value = true
  try {
    // 禁止前端生成验证码：验证码内容、干扰线和图片均由司机后端返回。
    const res = await getImgCaptcha(props.phone, { silentError: true })
    uuid.value = res.uuid || ''
    imgBase64.value = res.imgBase64 || ''
    if (!uuid.value || !imgBase64.value) throw new Error('图形验证码返回数据不完整')
  } catch (error) {
    uuid.value = ''
    imgBase64.value = ''
    showToast(apiErrorMessage(error, '图形验证码加载失败'))
  } finally {
    loading.value = false
  }
}

async function refreshCaptcha() {
  userInputCode.value = ''
  await loadCaptcha()
}

async function handleConfirm() {
  if (!uuid.value) {
    showToast('请刷新图形验证码')
    return
  }
  if (!userInputCode.value) {
    showToast('请输入验证码')
    return
  }
  submitting.value = true
  try {
    emit('confirm', { phone: props.phone, uuid: uuid.value, userInputCode: userInputCode.value })
  } finally {
    submitting.value = false
  }
}

async function handleClose() {
  const currentUuid = uuid.value
  if (currentUuid) {
    emit('close', { phone: props.phone, uuid: currentUuid })
    invalidateImgCaptcha({ phone: props.phone, uuid: currentUuid }, { silentError: true }).catch(() => {})
  }
  uuid.value = ''
  imgBase64.value = ''
  userInputCode.value = ''
  visible.value = false
}

function apiErrorMessage(error: unknown, fallbackMessage: string) {
  const err = error as { response?: { data?: { message?: string } }; message?: string }
  return err?.response?.data?.message || err?.message || fallbackMessage
}
</script>

<style scoped>
.img-captcha-popup {
  width: min(88vw, 340px);
  overflow: visible;
}

.captcha-card {
  position: relative;
  padding: 24px 18px 18px;
  background: #ffffff;
}

.captcha-close {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 30px;
  height: 30px;
  border: 0;
  border-radius: 50%;
  background: #f3f5f8;
  color: #5f6775;
  font-size: 22px;
  line-height: 30px;
}

.captcha-card h2 {
  margin: 4px 28px 18px 0;
  color: #172033;
  font-size: 18px;
  font-weight: 700;
  line-height: 1.35;
}

.captcha-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 118px;
  gap: 10px;
  align-items: stretch;
}

.captcha-input {
  min-width: 0;
  border: 1px solid #dde4ef;
  border-radius: 8px;
  background: #f8fafc;
}

.captcha-image-wrap {
  position: relative;
  min-height: 46px;
  border-radius: 8px;
  overflow: hidden;
  background: #eef3f8;
}

.captcha-image {
  display: block;
  width: 100%;
  height: 46px;
  object-fit: cover;
}

.captcha-image.empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #8792a2;
  font-size: 13px;
}

.refresh-link {
  position: absolute;
  right: 6px;
  bottom: 4px;
  border: 0;
  background: rgba(255, 255, 255, 0.88);
  color: #1677ff;
  font-size: 12px;
  line-height: 18px;
}

.confirm-button {
  width: 100%;
  height: 46px;
  margin-top: 20px;
  border: 0;
  border-radius: 8px;
  background: #1677ff;
  color: #ffffff;
  font-size: 16px;
  font-weight: 700;
}

.confirm-button:disabled,
.refresh-link:disabled {
  opacity: 0.55;
}
</style>
