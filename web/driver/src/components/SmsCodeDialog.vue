<template>
  <van-dialog v-model:show="visible" title="������֤��" confirm-button-text="֪����">
    <div class="sms-code-dialog">
      <p class="sms-code-dialog__phone">{{ phoneText }}</p>
      <div class="sms-code-dialog__code">{{ codeText }}</div>
      <p class="sms-code-dialog__tip">��֤�����Ժ�˷��ؽ������¼ʱ����д����֤�������</p>
    </div>
  </van-dialog>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  phone: { type: String, default: '' },
  code: { type: String, default: '' }
})

const emit = defineEmits(['update:show'])

const visible = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

const phoneText = computed(() => {
  const phone = String(props.phone || '').trim()
  if (!phone) return '�ֻ��ţ�--'
    if (phone.length <= 7) return '\u624b\u673a\u53f7\uff1a' + phone
    return '\u624b\u673a\u53f7\uff1a' + phone.slice(0, 3) + '****' + phone.slice(-4)
})

const codeText = computed(() => String(props.code || '').trim() || '���δ������֤��')
</script>

<style scoped>
.sms-code-dialog { display: grid; gap: 12px; padding: 18px 18px 20px; text-align: center; }
.sms-code-dialog__phone { margin: 0; color: #667085; font-size: 13px; }
.sms-code-dialog__code { padding: 16px 12px; border-radius: 12px; background: #f5f3ff; color: #5b5cff; font-size: 28px; font-weight: 800; letter-spacing: 2px; }
.sms-code-dialog__tip { margin: 0; color: #98a2b3; font-size: 12px; line-height: 1.5; }
</style>
