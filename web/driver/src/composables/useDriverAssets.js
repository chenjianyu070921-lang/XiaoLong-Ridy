import { computed, reactive, ref } from 'vue'
import { showConfirmDialog, showDialog, showToast } from 'vant'
import {
  createVehicle,
  createWithdraw,
  deleteVehicle,
  getCertification,
  getIncomeSummary,
  getTodayIncome,
  getVehicle,
  getWeekIncome,
  listIncomeBills,
  listWithdraws,
  updateVehicle,
  uploadCertification
} from '@/api/driver'
import { useDriverStore } from '@/stores/driver'
import { compact, dateToUnixSeconds, unixSecondsToDateInput } from '@/utils/driver-format'
import { resolveCreatedVehicle as resolveVehicleRecord } from '@/utils/vehicle'

export function useDriverAssets() {
  const driverStore = useDriverStore()

  const vehicleForm = reactive({
    plateNo: '',
    brand: '',
    model: '',
    color: '',
    vehicleType: 1,
    registrationDate: '',
    insuranceNo: '',
    insuranceExpireAt: ''
  })
  const certificationForm = reactive({ idCardFront: '', idCardBack: '', driverLicense: '', vehicleLicense: '' })
  const certUploading = reactive({ idCardFront: false, idCardBack: false, driverLicense: false, vehicleLicense: false })
  const certSubmitting = ref(false)
  const certItems = [
    { key: 'idCardFront', title: '身份证正面', tip: '人像面，信息清晰' },
    { key: 'idCardBack', title: '身份证反面', tip: '国徽面，有效期清晰' },
    { key: 'driverLicense', title: '驾驶证', tip: '正副页，准驾车型清晰' },
    { key: 'vehicleLicense', title: '行驶证', tip: '正副页，车牌号清晰' }
  ]
  const incomeSummary = ref({})
  const todayIncome = ref({})
  const weekIncome = ref({})
  const incomeBills = ref([])
  const withdrawRecords = ref([])
  const withdrawVisible = ref(false)
  const withdrawLoading = ref(false)
  const withdrawForm = reactive({ amount: '', payeeName: '', payAccount: '' })

  function syncVehicleForm() {
    if (!driverStore.vehicle) return
    Object.assign(vehicleForm, {
      plateNo: driverStore.vehicle.plateNo || '',
      brand: driverStore.vehicle.brand || '',
      model: driverStore.vehicle.model || '',
      color: driverStore.vehicle.color || '',
      vehicleType: Number(driverStore.vehicle.vehicleType || 1),
      registrationDate: unixSecondsToDateInput(driverStore.vehicle.registrationDate),
      insuranceNo: driverStore.vehicle.insuranceNo || '',
      insuranceExpireAt: unixSecondsToDateInput(driverStore.vehicle.insuranceExpireAt)
    })
  }

  function vehiclePayload() {
    return compact({
      plateNo: vehicleForm.plateNo,
      brand: vehicleForm.brand,
      model: vehicleForm.model,
      color: vehicleForm.color,
      vehicleType: Number(vehicleForm.vehicleType || 0),
      registrationDate: dateToUnixSeconds(vehicleForm.registrationDate),
      insuranceNo: vehicleForm.insuranceNo,
      insuranceExpireAt: dateToUnixSeconds(vehicleForm.insuranceExpireAt)
    })
  }

  async function loadVehicle(config = {}) {
    if (!driverStore.vehicleId) {
      try {
        await driverStore.refreshProfile({ silentError: true })
      } catch {
        return null
      }
    }
    if (!driverStore.vehicleId) return null
    const res = await safeApiCall(() => getVehicle(driverStore.vehicleId, config), '车辆查询失败', { silent: config.silentError })
    if (!res) return null
    driverStore.setVehicle(res.vehicle || res)
    syncVehicleForm()
    return res
  }

  async function submitVehicle() {
    const payload = vehiclePayload()
    if (driverStore.vehicleId) {
      const res = await safeApiCall(() => updateVehicle({ ...payload, id: driverStore.vehicleId }), '车辆更新失败')
      if (!res) return null
      showToast('车辆已更新')
      await loadVehicle()
      return res
    }
    const res = await safeApiCall(() => createVehicle(payload), '车辆提交失败')
    if (!res) return null
    const vehicle = await resolveVehicleRecord(payload, res, (id) => getVehicle(id, { silentError: true }))
    driverStore.setVehicle(vehicle)
    syncVehicleForm()
    showToast('车辆提交成功')
    return res
  }

  async function submitVehicleUpdate() {
    if (!driverStore.vehicleId) {
      showToast('请先查询或提交车辆')
      return null
    }
    const res = await safeApiCall(() => updateVehicle({ ...vehiclePayload(), id: driverStore.vehicleId }), '车辆更新失败')
    if (!res) return null
    showToast('车辆已更新')
    await loadVehicle()
    return res
  }

  async function removeVehicle() {
    if (!driverStore.vehicleId) {
      showToast('请先查询或提交车辆')
      return null
    }
    try {
      await showConfirmDialog({ title: '删除车辆', message: '确认删除当前车辆？' })
    } catch {
      return null
    }
    const res = await safeApiCall(() => deleteVehicle(driverStore.vehicleId), '车辆删除失败')
    if (!res) return null
    driverStore.setVehicle(null)
    showToast('车辆已删除')
    return res
  }

  async function loadCertification(config = {}) {
    const res = await safeApiCall(() => getCertification(config), '资质查询失败', { silent: config.silentError })
    if (!res) return null
    driverStore.setCertification(res.found === false ? null : (res.certification || res))
    return res
  }

  async function submitCertification() {
    const vehicleId = driverStore.vehicleId || driverStore.certification?.vehicleId
    if (!vehicleId) {
      showToast('请先绑定车辆')
      return null
    }
    const hasImage = certItems.some((item) => certificationForm[item.key])
    if (!hasImage) {
      showToast('请至少上传一项资质图片')
      return null
    }
    certSubmitting.value = true
    const payload = compact({
      vehicleId: Number(vehicleId),
      idCardFront: certificationForm.idCardFront,
      idCardBack: certificationForm.idCardBack,
      driverLicense: certificationForm.driverLicense,
      vehicleLicense: certificationForm.vehicleLicense
    })
    try {
      const res = await safeApiCall(() => uploadCertification(payload), '资质上传失败')
      if (!res) return null
      showToast('资质已提交，请等待审核')
      await loadCertification()
      return res
    } finally {
      certSubmitting.value = false
    }
  }

  async function readCertFile(event, field) {
    const file = event.target.files?.[0]
    if (!file) return
    if (file.size > 5 * 1024 * 1024) {
      showToast('图片不能超过5MB')
      event.target.value = ''
      return
    }
    certUploading[field] = true
    try {
      certificationForm[field] = await fileToBase64(file)
    } finally {
      certUploading[field] = false
      event.target.value = ''
    }
  }

  function removeCertImage(field) {
    certificationForm[field] = ''
  }

  async function loadIncome(config = {}) {
    const incomeRequests = [
      { label: '收入汇总', task: () => getIncomeSummary(config) },
      { label: '今日收入', task: () => getTodayIncome(config) },
      { label: '本周收入', task: () => getWeekIncome(config) },
      { label: '收入明细', task: () => listIncomeBills({ page: 1, pageSize: 20 }, config) },
      { label: '提现记录', task: () => listWithdraws({ page: 1, pageSize: 20 }, config) }
    ]
    const [summary, today, week, bills, withdraws] = await Promise.allSettled(incomeRequests.map((item) => item.task()))
    incomeSummary.value = summary.status === 'fulfilled' ? summary.value : {}
    todayIncome.value = today.status === 'fulfilled' ? today.value : {}
    weekIncome.value = week.status === 'fulfilled' ? week.value : {}
    incomeBills.value = bills.status === 'fulfilled' && Array.isArray(bills.value.list) ? bills.value.list : []
    withdrawRecords.value = withdraws.status === 'fulfilled' && Array.isArray(withdraws.value.list) ? withdraws.value.list : []

    if (!config.silentError) {
      const failures = [summary, today, week, bills, withdraws]
        .map((result, index) => result.status === 'rejected' ? incomeRequests[index].label + ': ' + apiErrorMessage(result.reason, '请求失败') : '')
        .filter(Boolean)
      if (failures.length) showIncomeLoadFailure(failures)
    }
    return { summary: incomeSummary.value, today: todayIncome.value, week: weekIncome.value, bills: incomeBills.value, withdraws: withdrawRecords.value }
  }

  function openWithdraw() {
    withdrawVisible.value = true
  }

  async function submitWithdraw() {
    const amount = Number(withdrawForm.amount)
    if (!Number.isFinite(amount) || amount <= 0 || !withdrawForm.payeeName.trim() || !withdrawForm.payAccount.trim()) {
      showToast('请填写完整的提现信息')
      return null
    }
    withdrawLoading.value = true
    try {
      const res = await safeApiCall(() => createWithdraw({
        amount,
        payeeName: withdrawForm.payeeName.trim(),
        payAccount: withdrawForm.payAccount.trim()
      }), '提现申请失败')
      if (!res) return null
      withdrawVisible.value = false
      Object.assign(withdrawForm, { amount: '', payeeName: '', payAccount: '' })
      showToast('提现申请已提交')
      await loadIncome({ silentError: true })
      return res
    } finally {
      withdrawLoading.value = false
    }
  }

  const certStatusIcon = computed(() => {
    const status = driverStore.certification?.auditStatus
    if (status === 2) return 'checked'
    if (status === 3) return 'warning-o'
    return 'clock-o'
  })

  function syncFromStore() {
    syncVehicleForm()
  }

  return {
    driverStore,
    vehicleForm,
    certificationForm,
    certUploading,
    certItems,
    certSubmitting,
    certStatusIcon,
    incomeSummary,
    todayIncome,
    weekIncome,
    incomeBills,
    withdrawRecords,
    withdrawVisible,
    withdrawLoading,
    withdrawForm,
    syncVehicleForm,
    syncFromStore,
    loadVehicle,
    submitVehicle,
    submitVehicleUpdate,
    removeVehicle,
    loadCertification,
    submitCertification,
    readCertFile,
    removeCertImage,
    loadIncome,
    openWithdraw,
    submitWithdraw
  }
}

function safeApiCall(task, fallbackMessage = '请求失败', options = {}) {
  return task().catch((error) => {
    if (!options.silent) showToast(apiErrorMessage(error, fallbackMessage))
    return null
  })
}

function apiErrorMessage(error, fallbackMessage = '请求失败') {
  return error?.response?.data?.message || error?.message || fallbackMessage
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('文件读取失败'))
    reader.readAsDataURL(file)
  })
}

function showIncomeLoadFailure(failures) {
  showDialog({
    title: '收入数据加载失败',
    message: failures.join('\n')
  }).catch(() => {})
}
