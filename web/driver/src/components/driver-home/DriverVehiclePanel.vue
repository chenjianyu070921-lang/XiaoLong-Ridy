<template>
  <section class="vehicle-manage">
    <div class="vm-head">
      <h2>车辆管理</h2>
      <span class="vm-count">已绑定 {{ vehicles.length }}/3</span>
      <button type="button" class="vm-add" :disabled="!canAdd" @click="openAdd">+ 添加车辆</button>
    </div>

    <p v-if="!canAdd" class="vm-hint">已达上限（最多 3 辆），请删除旧车辆后再添加新车。</p>

    <ul v-if="vehicles.length" class="vm-list">
      <li v-for="v in vehicles" :key="v.id" class="vm-card">
        <div class="vm-card-main">
          <p class="vm-plate">{{ v.plateNo || '未填车牌' }}</p>
          <p class="vm-sub">{{ v.brand || '未知品牌' }} {{ v.model || '' }} · {{ v.color || '未填颜色' }}</p>
          <p class="vm-meta">类型：{{ vehicleTypeLabel(v.vehicleType) }} · 状态：{{ formatVehicleStatus(v.status) }}</p>
        </div>
        <div class="vm-card-actions">
          <button type="button" class="vm-edit" @click="openEdit(v)">修改</button>
          <button type="button" class="vm-del" @click="onDelete(v)">删除</button>
        </div>
      </li>
    </ul>

    <div v-else class="vm-empty">
      <p>暂无绑定车辆</p>
      <button type="button" class="vm-add" :disabled="!canAdd" @click="openAdd">+ 添加车辆</button>
    </div>

    <!-- 添加/修改弹窗：居中、淡紫背景 -->
    <van-dialog
      v-model:show="formVisible"
      :title="editingId ? '修改车辆' : '添加车辆'"
      class="vehicle-modal"
      :show-confirm-button="false"
      :show-cancel-button="false"
      @closed="onClosed"
    >
      <div class="vm-form">
        <van-field v-model="form.plateNo" label="车牌号" placeholder="粤B12345" />
        <van-field v-model="form.brand" label="品牌" placeholder="BYD" />
        <van-field v-model="form.model" label="型号" placeholder="Han" />
        <van-field v-model="form.color" label="颜色" placeholder="黑色" />
        <van-field
          v-model="typeLabel"
          label="车辆类型"
          readonly
          placeholder="请选择"
          @click="typePickerVisible = true"
        />
        <van-field v-model="form.registrationDate" type="date" label="注册日期" />
        <van-field v-model="form.insuranceNo" label="保险单号" placeholder="INS-001" />
        <van-field v-model="form.insuranceExpireAt" type="date" label="保险到期" />
        <div class="vm-form-actions">
          <button type="button" class="vm-cancel" @click="formVisible = false">取消</button>
          <button type="button" class="vm-submit" :disabled="submitting" @click="onSubmit">
            {{ submitting ? '提交中…' : (editingId ? '保存' : '添加') }}
          </button>
        </div>
      </div>
    </van-dialog>

    <van-popup v-model:show="typePickerVisible" position="bottom" round>
      <van-picker
        :columns="typeColumns"
        @confirm="onTypeConfirm"
        @cancel="typePickerVisible = false"
      />
    </van-popup>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { showConfirmDialog, showToast } from 'vant'
import { createVehicle, deleteVehicle, updateVehicle } from '@/api/driver'
import { useDriverStore } from '@/stores/driver'
import { compact, dateToUnixSeconds, formatVehicleStatus, unixSecondsToDateInput } from '@/utils/driver-format'
import { safeApiCall } from '@/utils/safe-request'

const MAX_VEHICLES = 3
const driverStore = useDriverStore()
const vehicles = computed(() => driverStore.vehicles || [])
const canAdd = computed(() => vehicles.value.length < MAX_VEHICLES)

const typeColumns = [
  { text: '特惠快车', value: 1 },
  { text: '快车', value: 2 },
  { text: '拼车', value: 3 }
]
function vehicleTypeLabel(type) {
  const found = typeColumns.find((t) => t.value === Number(type))
  return found ? found.text : '未知'
}

const formVisible = ref(false)
const typePickerVisible = ref(false)
const editingId = ref(null)
const submitting = ref(false)
const form = reactive({
  plateNo: '', brand: '', model: '', color: '', vehicleType: 1,
  registrationDate: '', insuranceNo: '', insuranceExpireAt: ''
})
const typeLabel = computed(() => vehicleTypeLabel(form.vehicleType))

function resetForm() {
  Object.assign(form, {
    plateNo: '', brand: '', model: '', color: '', vehicleType: 1,
    registrationDate: '', insuranceNo: '', insuranceExpireAt: ''
  })
}
function openAdd() {
  if (!canAdd.value) {
    showToast('最多绑定3辆车，请先删除旧车辆')
    return
  }
  editingId.value = null
  resetForm()
  formVisible.value = true
}
function openEdit(v) {
  editingId.value = v.id
  Object.assign(form, {
    plateNo: v.plateNo || '', brand: v.brand || '', model: v.model || '', color: v.color || '',
    vehicleType: Number(v.vehicleType || 1),
    registrationDate: unixSecondsToDateInput(v.registrationDate),
    insuranceNo: v.insuranceNo || '',
    insuranceExpireAt: unixSecondsToDateInput(v.insuranceExpireAt)
  })
  formVisible.value = true
}
function onTypeConfirm(payload) {
  const value = payload?.selectedValues?.[0] ?? payload?.selectedOptions?.[0]?.value
  if (value != null) form.vehicleType = Number(value)
  typePickerVisible.value = false
}
function onClosed() {
  resetForm()
}

function vehiclePayload() {
  return compact({
    plateNo: form.plateNo.trim(),
    brand: form.brand.trim(),
    model: form.model.trim(),
    color: form.color.trim(),
    vehicleType: Number(form.vehicleType || 0),
    registrationDate: dateToUnixSeconds(form.registrationDate),
    insuranceNo: form.insuranceNo.trim(),
    insuranceExpireAt: dateToUnixSeconds(form.insuranceExpireAt)
  })
}

async function onSubmit() {
  if (!form.plateNo.trim()) {
    showToast('请输入车牌号')
    return
  }
  submitting.value = true
  const payload = vehiclePayload()
  try {
    const res = editingId.value
      ? await safeApiCall(() => updateVehicle({ ...payload, id: editingId.value }))
      : await safeApiCall(() => createVehicle(payload))
    if (!res) return
    showToast(editingId.value ? '车辆已更新' : '车辆已添加')
    formVisible.value = false
    await driverStore.loadVehicles({ silentError: true })
  } finally {
    submitting.value = false
  }
}

async function onDelete(v) {
  try {
    await showConfirmDialog({ title: '删除车辆', message: `确认删除车牌 ${v.plateNo || ''} 的车辆？` })
  } catch {
    return
  }
  const res = await safeApiCall(() => deleteVehicle(v.id))
  if (!res) return
  showToast('车辆已删除')
  await driverStore.loadVehicles({ silentError: true })
}

onMounted(() => {
  void driverStore.loadVehicles({ silentError: true })
})
</script>

<style scoped>
.vehicle-manage { padding: 12px; }
.vm-head { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.vm-head h2 { font-size: 16px; margin: 0; flex: 1; }
.vm-count { font-size: 12px; color: var(--driver-muted); }
.vm-add { border: none; background: var(--driver-primary); color: var(--driver-on-primary); border-radius: 999px; padding: 6px 14px; font-size: 13px; }
.vm-add:disabled { background: var(--driver-soft); }
.vm-hint { font-size: 12px; color: #f59e0b; background: rgba(245,158,11,.14); border-radius: 8px; padding: 8px 10px; margin: 0 0 10px; }
.vm-list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 10px; }
.vm-card { display: flex; align-items: center; background: var(--driver-card); border-radius: 12px; padding: 12px; box-shadow: 0 2px 8px rgba(0,0,0,.04); }
.vm-card-main { flex: 1; min-width: 0; }
.vm-plate { font-size: 15px; font-weight: 600; margin: 0 0 4px; color: var(--driver-ink); }
.vm-sub { font-size: 13px; margin: 0 0 2px; color: var(--driver-ink); }
.vm-meta { font-size: 12px; margin: 0; color: var(--driver-muted); }
.vm-card-actions { display: flex; flex-direction: column; gap: 8px; }
.vm-edit { border: 1px solid var(--driver-primary); background: var(--driver-card); color: var(--driver-primary); border-radius: 8px; padding: 5px 12px; font-size: 13px; }
.vm-del { border: 1px solid #f0c2c2; background: var(--driver-card); color: #e5484d; border-radius: 8px; padding: 5px 12px; font-size: 13px; }
.vm-empty { text-align: center; color: var(--driver-muted); padding: 24px 0; }
.vm-empty p { margin: 0 0 12px; }
.vm-form-actions { display: flex; gap: 10px; padding: 12px 4px 4px; }
.vm-cancel { flex: 1; border: 1px solid var(--driver-line); background: var(--driver-card); border-radius: 8px; padding: 8px; font-size: 14px; }
.vm-submit { flex: 1; border: none; background: var(--driver-primary); color: var(--driver-on-primary); border-radius: 8px; padding: 8px; font-size: 14px; }
.vm-submit:disabled { opacity: .6; }
</style>

<!-- van-dialog 默认 teleport 到 body，弹窗样式需用非 scoped 规则才能命中 -->
<style>
.vehicle-modal {
  background: var(--driver-soft) !important;
  border-radius: 16px;
  max-width: 88vw;
}
.vehicle-modal .van-dialog__header {
  color: var(--driver-primary);
  font-weight: 600;
}
.vehicle-modal .van-cell {
  background: rgba(255, 255, 255, 0.72);
  margin: 6px 0;
  border-radius: 10px;
}
</style>
