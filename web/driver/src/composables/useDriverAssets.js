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
import { compact, dateToUnixSeconds, unixSecondsToDateInput, yuanToCents } from '@/utils/driver-format'
import { resolveCreatedVehicle as resolveVehicleRecord } from '@/utils/vehicle'
import { apiErrorMessage, safeApiCall } from '@/utils/safe-request'

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
  const certificationForm = reactive({ idCardNo: '', realName: '', driverLicenseNo: '' })
  const certSubmitting = ref(false)
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
    const res = await safeApiCall(() => getVehicle(driverStore.vehicleId, config))
    if (!res) return null
    driverStore.setVehicle(res.vehicle || res)
    syncVehicleForm()
    return res
  }

  async function submitVehicle() {
    const payload = vehiclePayload()
    if (driverStore.vehicleId) {
      const res = await safeApiCall(() => updateVehicle({ ...payload, id: driverStore.vehicleId }))
      if (!res) return null
      showToast('车辆已更新')
      await loadVehicle()
      return res
    }
    const res = await safeApiCall(() => createVehicle(payload))
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
    const res = await safeApiCall(() => updateVehicle({ ...vehiclePayload(), id: driverStore.vehicleId }))
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
    const res = await safeApiCall(() => deleteVehicle(driverStore.vehicleId))
    if (!res) return null
    driverStore.setVehicle(null)
    showToast('车辆已删除')
    return res
  }

  async function loadCertification(config = {}) {
    const res = await safeApiCall(() => getCertification(config))
    if (!res) return null
    driverStore.setCertification(res.found === false ? null : (res.certification || res))
    return res
  }

  function syncCertificationForm() {
    const driver = driverStore.driver || {}
    certificationForm.idCardNo = driver.idCardNo || ''
    certificationForm.realName = driver.realName || ''
    certificationForm.driverLicenseNo = driver.driverLicenseNo || ''
  }

  function validateCertification() {
    const idCard = (certificationForm.idCardNo || '').trim()
    const name = (certificationForm.realName || '').trim()
    const licenseNo = (certificationForm.driverLicenseNo || '').trim()
    if (!/^\d{17}[\dXx]$/.test(idCard)) {
      showToast('请输入正确的 18 位身份证号')
      return false
    }
    if (name.length < 2) {
      showToast('请输入真实姓名')
      return false
    }
    if (!licenseNo) {
      showToast('请输入驾驶证编号')
      return false
    }
    return true
  }

  async function submitCertification() {
    const vehicleId = driverStore.vehicleId || driverStore.certification?.vehicleId
    if (!vehicleId) {
      showToast('请先绑定车辆')
      return null
    }
    if (!validateCertification()) return null
    certSubmitting.value = true
    const payload = {
      vehicleId: Number(vehicleId),
      idCardNo: certificationForm.idCardNo.trim(),
      realName: certificationForm.realName.trim(),
      driverLicenseNo: certificationForm.driverLicenseNo.trim()
    }
    try {
      const res = await safeApiCall(() => uploadCertification(payload))
      if (!res) return null
      showToast('资质已提交，请等待审核')
      await loadCertification()
      return res
    } finally {
      certSubmitting.value = false
    }
  }

  async function loadIncome(config = {}) {
    // 五个请求固定静默：失败统一由下方 showIncomeLoadFailure 汇总成一个弹窗，
    // 否则每个请求都会各自触发拦截器 Toast，连弹 5 次后再叠一个弹窗。
    const silent = { silentError: true }
    const incomeRequests = [
      { label: '收入汇总', task: () => getIncomeSummary(silent) },
      { label: '今日收入', task: () => getTodayIncome(silent) },
      { label: '本周收入', task: () => getWeekIncome(silent) },
      { label: '收入明细', task: () => listIncomeBills({ page: 1, pageSize: 20 }, silent) },
      { label: '提现记录', task: () => listWithdraws({ page: 1, pageSize: 20 }, silent) }
    ]
    const [summary, today, week, bills, withdraws] = await Promise.allSettled(incomeRequests.map((item) => item.task()))
    incomeSummary.value = summary.status === 'fulfilled' ? summary.value : {}
    todayIncome.value = today.status === 'fulfilled' ? today.value : {}
    weekIncome.value = week.status === 'fulfilled' ? week.value : {}
    incomeBills.value = bills.status === 'fulfilled' && Array.isArray(bills.value.list) ? bills.value.list : []
    withdrawRecords.value = withdraws.status === 'fulfilled' && Array.isArray(withdraws.value.list)
      ? withdraws.value.list.map(normalizeWithdrawRecord)
      : []

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

  // 注意：后端 /withdraws 期望的 amount 单位为「元」（driver_withdraw.amount 为 DECIMAL(10,2)），
  // 与收入的「分」不同，此处不可换算。展示侧的单位差异由 normalizeWithdrawRecord 收敛。
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
      }))
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
    certSubmitting,
    syncCertificationForm,
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
    syncCertificationForm,
    loadIncome,
    openWithdraw,
    submitWithdraw
  }
}

/**
 * 提现记录单位归一化：后端 withdraw.amount 为「元」，前端统一按「分」展示。
 * 在此数据入口一次性转成 amountCents，视图层即可与收入字段（incomeCents 等）共用 formatPrice，
 * 无需各自做单位换算。详见 utils/driver-format.js 中的单位契约说明。
 */
function normalizeWithdrawRecord(record) {
  return { ...record, amountCents: yuanToCents(record?.amount) }
}

function showIncomeLoadFailure(failures) {
  showDialog({
    title: '收入数据加载失败',
    message: failures.join('\n')
  }).catch(() => {})
}
