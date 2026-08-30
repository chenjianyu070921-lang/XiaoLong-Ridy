<script setup>
// 管理后台业务字段化表单弹窗：只处理展示、校验和请求体转换，不直接发起网络请求。
// 父页面根据 type 调用现有 API，确保前端改造不改变后端接口边界。
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  modelValue: Boolean,
  type: { type: String, required: true },
  title: { type: String, default: '' },
  record: { type: Object, default: () => ({}) },
  loading: Boolean,
})
const emit = defineEmits(['update:modelValue', 'submit'])

const formRef = ref()
const form = reactive({})
const isCoupon = computed(() => ['createCoupon', 'editCoupon'].includes(props.type))
const isPriceRule = computed(() => ['createPriceRule', 'editPriceRule'].includes(props.type))
const isActivity = computed(() => ['createActivity', 'editActivity'].includes(props.type))
const isReasonAction = computed(() => ['freeze', 'unfreeze', 'freezeDriver', 'cancel', 'release'].includes(props.type))

// createDefaults 按接口字段初始化表单；金额始终保留字符串，避免前端浮点计算误差。
const createDefaults = () => ({
  name: '', type: 1, face_value: '', discount: '1.00', threshold_amount: '0.00', total_count: 1, per_user_limit: 1,
  valid_range: [], status: 1, target_type: 'explicit_users', user_ids: '', city_code: '', car_type: 1,
  base_price: '', base_distance_km: '', per_km_price: '', per_minute_price: '', night_start_time: '22:00:00', night_end_time: '06:00:00',
  night_surcharge: '', dynamic_max_factor: '', effective_range: [], config: '{}', start_range: [],
  publish_scope: 'full', gray_type: 'user_ids', gray_user_ids: '', gray_city_codes: '', gray_ratio: 10,
  work_order_type: 1, source_type: 'order', source_id: '', order_id: '', user_id: '', driver_id: '', title: '', content: '', priority: 2,
  action: 'follow', assignee_id: '', arbitration_result: '', version: 0, evidence_type: 'text', evidence_url: '',
  ids: [], risk_action: 'review_pass', work_order_title: '',
  reason: '', remark: '', target_id: '', export_type: 'orders', export_range: [], export_user_id: '', export_driver_id: '', export_order_id: '', export_admin_id: '', export_city_code: '',
})
const copyRecord = () => {
  Object.keys(form).forEach((key) => delete form[key])
  const defaults = createDefaults()
  Object.assign(form, defaults, props.record || {})
  if (isCoupon.value) form.valid_range = [props.record?.valid_start_at, props.record?.valid_end_at].filter(Boolean)
  if (isPriceRule.value) form.effective_range = [props.record?.effective_at, props.record?.expire_at].filter(Boolean)
  if (isActivity.value) form.start_range = [props.record?.start_at, props.record?.end_at].filter(Boolean)
  if (props.type === 'workOrderAction') form.version = props.record?.version || 0
}
watch(() => [props.modelValue, props.type, props.record], ([visible]) => { if (visible) copyRecord() }, { deep: true })

const required = (message) => ({ required: true, message, trigger: 'blur' })
const positiveInteger = (_rule, value, callback) => (/^[1-9]\d*$/.test(String(value)) ? callback() : callback(new Error('请输入正整数')))
const amount = (_rule, value, callback) => (/^\d+(\.\d{1,2})?$/.test(String(value)) ? callback() : callback(new Error('请输入最多两位小数的非负金额')))
// validRange 校验双端日期，防止将不完整或倒置的时间范围提交给服务端。
const validRange = (_rule, value, callback) => (Array.isArray(value) && value.length === 2 && value[0] < value[1] ? callback() : callback(new Error('请选择完整且有效的时间范围')))
const rules = computed(() => ({
  name: [required('请输入名称')], face_value: [required('请输入面额'), { validator: amount, trigger: 'blur' }], discount: [required('请输入折扣'), { validator: amount, trigger: 'blur' }],
  threshold_amount: [required('请输入使用门槛'), { validator: amount, trigger: 'blur' }], total_count: [{ validator: positiveInteger, trigger: 'blur' }], per_user_limit: [{ validator: positiveInteger, trigger: 'blur' }],
  city_code: [required('请输入城市编码')], base_price: [required('请输入起步价'), { validator: amount, trigger: 'blur' }], base_distance_km: [required('请输入起步里程'), { validator: amount, trigger: 'blur' }],
  per_km_price: [required('请输入每公里价格'), { validator: amount, trigger: 'blur' }], per_minute_price: [required('请输入每分钟价格'), { validator: amount, trigger: 'blur' }],
  night_surcharge: [required('请输入夜间加价'), { validator: amount, trigger: 'blur' }], dynamic_max_factor: [required('请输入动态上限'), { validator: amount, trigger: 'blur' }],
  valid_range: [{ validator: validRange, trigger: 'change' }], effective_range: [{ validator: validRange, trigger: 'change' }], start_range: [{ validator: validRange, trigger: 'change' }], user_ids: [required('请输入至少一个用户 ID')], title: [required('请输入工单标题')], content: [required('请输入处理内容')], reason: [required('请输入原因')], target_id: [{ validator: positiveInteger, trigger: 'blur' }],
}))

// toPayload 是唯一的参数转换出口，严格输出后端已声明字段，禁止透传列表中的只读数据。
const toPayload = () => {
  if (isCoupon.value) return { name: form.name, type: Number(form.type), face_value: form.face_value, discount: form.discount, threshold_amount: form.threshold_amount, total_count: Number(form.total_count), per_user_limit: Number(form.per_user_limit), valid_start_at: form.valid_range?.[0] || '', valid_end_at: form.valid_range?.[1] || '', status: Number(form.status) }
  if (props.type === 'issueCoupon') {
    const ids = [...new Set(form.user_ids.split(/[\s,，]+/).filter(Boolean).map(Number))]
    if (!ids.length || ids.some((id) => !Number.isSafeInteger(id) || id <= 0)) throw new Error('用户 ID 必须是以逗号或换行分隔的正整数')
    return { target_type: 'explicit_users', target_config: JSON.stringify({ user_ids: ids }) }
  }
  if (isPriceRule.value) return { name: form.name, city_code: form.city_code, car_type: Number(form.car_type), base_price: form.base_price, base_distance_km: form.base_distance_km, per_km_price: form.per_km_price, per_minute_price: form.per_minute_price, night_start_time: form.night_start_time, night_end_time: form.night_end_time, night_surcharge: form.night_surcharge, dynamic_max_factor: form.dynamic_max_factor, status: Number(form.status), effective_at: form.effective_range?.[0] || '', expire_at: form.effective_range?.[1] || '' }
  if (isActivity.value) { JSON.parse(form.config); return { name: form.name, type: Number(form.type), config: form.config, start_at: form.start_range?.[0] || '', end_at: form.start_range?.[1] || '', status: Number(form.status) } }
  // 活动发布只输出后端约定的发布范围和 JSON 字符串，不透传表单临时字段。
  if (props.type === 'publish') {
    if (form.publish_scope === 'full') return { publish_scope: 'full', target_config: '{}' }
    if (form.publish_scope !== 'gray') throw new Error('请选择合法的活动发布范围')
    if (form.gray_type === 'user_ids') {
      const ids = [...new Set(String(form.gray_user_ids || '').split(/[\s,，]+/).filter(Boolean).map(Number))]
      if (!ids.length || ids.some((id) => !Number.isSafeInteger(id) || id <= 0)) throw new Error('灰度用户 ID 必须是以逗号或换行分隔的正整数')
      return { publish_scope: 'gray', target_config: JSON.stringify({ user_ids: ids }) }
    }
    if (form.gray_type === 'city_codes') {
      const cityCodes = [...new Set(String(form.gray_city_codes || '').split(/[\s,，]+/).map((code) => code.trim()).filter(Boolean))]
      if (!cityCodes.length) throw new Error('请至少填写一个灰度城市编码')
      return { publish_scope: 'gray', target_config: JSON.stringify({ city_codes: cityCodes }) }
    }
    if (form.gray_type === 'ratio') {
      const ratio = Number(form.gray_ratio)
      if (!Number.isInteger(ratio) || ratio < 1 || ratio > 100) throw new Error('灰度比例必须是 1 到 100 的整数')
      return { publish_scope: 'gray', target_config: JSON.stringify({ ratio }) }
    }
    throw new Error('请选择合法的灰度方式')
  }
  if (props.type === 'createWorkOrder') return { work_order_type: Number(form.work_order_type), source_type: form.source_type, source_id: Number(form.source_id || 0), order_id: Number(form.order_id || 0), user_id: Number(form.user_id || 0), driver_id: Number(form.driver_id || 0), title: form.title, content: form.content, priority: Number(form.priority) }
  if (props.type === 'workOrderAction') return { action: form.action, assignee_id: Number(form.assignee_id || 0), content: form.content, arbitration_result: form.arbitration_result, version: Number(form.version) }
  if (props.type === 'batchWorkOrderAction') return { ids: form.ids || [], action: form.action, assignee_id: Number(form.assignee_id || 0), content: form.content, arbitration_result: form.arbitration_result }
  if (props.type === 'workOrderEvidence') return { evidence_type: form.evidence_type, evidence_url: form.evidence_url, content: form.content }
  if (props.type === 'riskHitAction') return { ids: form.ids || [], action: form.risk_action, reason: form.reason, work_order_title: form.work_order_title, priority: Number(form.priority || 0) }
  if (props.type === 'createBlacklist') return { target_type: form.target_type, target_id: Number(form.target_id), reason: form.reason }
  if (props.type === 'createExport') { const filters = {}; [['start_time', form.export_range?.[0]], ['end_time', form.export_range?.[1]], ['user_id', form.export_user_id], ['driver_id', form.export_driver_id], ['order_id', form.export_order_id], ['admin_id', form.export_admin_id], ['city_code', form.export_city_code]].forEach(([key, value]) => { if (value !== '' && value !== undefined) filters[key] = key.endsWith('_id') ? Number(value) : value }); return { export_type: form.export_type, filters: JSON.stringify(filters) } }
  if (props.type === 'approve' || props.type === 'reject') return { remark: form.remark }
  if (props.type === 'freeze' || props.type === 'unfreeze' || props.type === 'freezeDriver') return { reason: form.reason, remark: form.remark }
  if (props.type === 'cancel') return { reason: form.reason, request_id: crypto.randomUUID ? crypto.randomUUID() : `cancel-${Date.now()}-${Math.random().toString(16).slice(2)}` }
  if (props.type === 'release') return { reason: form.reason }
  return {}
}
const submit = async () => {
  try { await formRef.value?.validate(); emit('submit', { payload: toPayload(), record: props.record }) } catch (error) { if (error?.message) ElMessage.error(error.message) }
}
const close = () => emit('update:modelValue', false)
</script>

<template>
  <el-dialog :model-value="modelValue" :title="title" width="680px" destroy-on-close @update:model-value="close">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="108px" class="business-form">
      <template v-if="isCoupon"><el-form-item label="优惠券名称" prop="name"><el-input v-model="form.name" maxlength="50" /></el-form-item><el-form-item label="优惠券类型"><el-select v-model="form.type"><el-option label="满减券" :value="1"/><el-option label="折扣券" :value="2"/><el-option label="立减券" :value="3"/></el-select></el-form-item><el-row :gutter="16"><el-col :span="12"><el-form-item label="面额" prop="face_value"><el-input v-model="form.face_value" /></el-form-item></el-col><el-col :span="12"><el-form-item label="折扣" prop="discount"><el-input v-model="form.discount" /></el-form-item></el-col></el-row><el-row :gutter="16"><el-col :span="12"><el-form-item label="使用门槛" prop="threshold_amount"><el-input v-model="form.threshold_amount" /></el-form-item></el-col><el-col :span="12"><el-form-item label="库存" prop="total_count"><el-input v-model="form.total_count" /></el-form-item></el-col></el-row><el-form-item label="单人限领" prop="per_user_limit"><el-input v-model="form.per_user_limit" /></el-form-item><el-form-item label="有效期" prop="valid_range"><el-date-picker v-model="form.valid_range" type="datetimerange" value-format="YYYY-MM-DD HH:mm:ss" /></el-form-item><el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio :value="1">草稿</el-radio><el-radio :value="2">启用</el-radio></el-radio-group></el-form-item></template>
      <template v-else-if="props.type === 'issueCoupon'"><el-form-item label="发放对象"><el-input value="显式用户" disabled /></el-form-item><el-form-item label="用户 ID" prop="user_ids"><el-input v-model="form.user_ids" type="textarea" :rows="4" placeholder="多个 ID 用逗号或换行分隔" /></el-form-item></template>
      <template v-else-if="isPriceRule"><el-form-item label="规则名称" prop="name"><el-input v-model="form.name" /></el-form-item><el-row :gutter="16"><el-col :span="12"><el-form-item label="城市编码" prop="city_code"><el-input v-model="form.city_code" /></el-form-item></el-col><el-col :span="12"><el-form-item label="车型"><el-input-number v-model="form.car_type" :min="1" /></el-form-item></el-col></el-row><el-row :gutter="16"><el-col :span="12"><el-form-item label="起步价" prop="base_price"><el-input v-model="form.base_price" /></el-form-item></el-col><el-col :span="12"><el-form-item label="起步里程" prop="base_distance_km"><el-input v-model="form.base_distance_km" /></el-form-item></el-col></el-row><el-row :gutter="16"><el-col :span="12"><el-form-item label="每公里价" prop="per_km_price"><el-input v-model="form.per_km_price" /></el-form-item></el-col><el-col :span="12"><el-form-item label="每分钟价" prop="per_minute_price"><el-input v-model="form.per_minute_price" /></el-form-item></el-col></el-row><el-row :gutter="16"><el-col :span="12"><el-form-item label="夜间开始"><el-time-picker v-model="form.night_start_time" value-format="HH:mm:ss" /></el-form-item></el-col><el-col :span="12"><el-form-item label="夜间结束"><el-time-picker v-model="form.night_end_time" value-format="HH:mm:ss" /></el-form-item></el-col></el-row><el-row :gutter="16"><el-col :span="12"><el-form-item label="夜间加价" prop="night_surcharge"><el-input v-model="form.night_surcharge" /></el-form-item></el-col><el-col :span="12"><el-form-item label="动态上限" prop="dynamic_max_factor"><el-input v-model="form.dynamic_max_factor" /></el-form-item></el-col></el-row><el-form-item label="生效期" prop="effective_range"><el-date-picker v-model="form.effective_range" type="datetimerange" value-format="YYYY-MM-DD HH:mm:ss" /></el-form-item><el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio :value="1">草稿</el-radio><el-radio :value="2">启用</el-radio></el-radio-group></el-form-item></template>
      <template v-else-if="isActivity"><el-form-item label="活动名称" prop="name"><el-input v-model="form.name" /></el-form-item><el-form-item label="活动类型"><el-select v-model="form.type"><el-option label="拉新" :value="1"/><el-option label="折扣" :value="2"/><el-option label="立减" :value="3"/></el-select></el-form-item><el-form-item label="活动时段" prop="start_range"><el-date-picker v-model="form.start_range" type="datetimerange" value-format="YYYY-MM-DD HH:mm:ss" /></el-form-item><el-form-item label="活动配置"><el-input v-model="form.config" type="textarea" :rows="5" placeholder="活动配置 JSON" /></el-form-item><el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio :value="1">草稿</el-radio><el-radio :value="2">启用</el-radio></el-radio-group></el-form-item></template>
      <template v-else-if="props.type === 'publish'"><el-form-item label="发布范围"><el-radio-group v-model="form.publish_scope"><el-radio value="full">全量发布</el-radio><el-radio value="gray">灰度发布</el-radio></el-radio-group></el-form-item><template v-if="form.publish_scope === 'gray'"><el-form-item label="灰度方式"><el-select v-model="form.gray_type"><el-option label="指定用户" value="user_ids"/><el-option label="指定城市" value="city_codes"/><el-option label="按比例" value="ratio"/></el-select></el-form-item><el-form-item v-if="form.gray_type === 'user_ids'" label="用户 ID"><el-input v-model="form.gray_user_ids" type="textarea" :rows="4" placeholder="多个用户 ID 用逗号或换行分隔" /></el-form-item><el-form-item v-if="form.gray_type === 'city_codes'" label="城市编码"><el-input v-model="form.gray_city_codes" type="textarea" :rows="3" placeholder="多个城市编码用逗号或换行分隔" /></el-form-item><el-form-item v-if="form.gray_type === 'ratio'" label="灰度比例"><el-input-number v-model="form.gray_ratio" :min="1" :max="100" :step="1" controls-position="right" /><span class="field-suffix">%</span></el-form-item></template></template>
      <template v-else-if="props.type === 'createWorkOrder'"><el-row :gutter="16"><el-col :span="12"><el-form-item label="工单类型"><el-select v-model="form.work_order_type"><el-option label="投诉" :value="1"/><el-option label="申诉" :value="2"/></el-select></el-form-item></el-col><el-col :span="12"><el-form-item label="优先级"><el-select v-model="form.priority"><el-option label="低" :value="1"/><el-option label="中" :value="2"/><el-option label="高" :value="3"/></el-select></el-form-item></el-col></el-row><el-row :gutter="16"><el-col :span="12"><el-form-item label="来源类型"><el-select v-model="form.source_type"><el-option label="订单" value="order"/><el-option label="用户" value="user"/><el-option label="司机" value="driver"/></el-select></el-form-item></el-col><el-col :span="12"><el-form-item label="来源 ID"><el-input v-model="form.source_id" /></el-form-item></el-col></el-row><el-row :gutter="16"><el-col :span="8"><el-form-item label="订单 ID"><el-input v-model="form.order_id" /></el-form-item></el-col><el-col :span="8"><el-form-item label="用户 ID"><el-input v-model="form.user_id" /></el-form-item></el-col><el-col :span="8"><el-form-item label="司机 ID"><el-input v-model="form.driver_id" /></el-form-item></el-col></el-row><el-form-item label="工单标题" prop="title"><el-input v-model="form.title" /></el-form-item><el-form-item label="工单内容" prop="content"><el-input v-model="form.content" type="textarea" :rows="4" /></el-form-item></template>
      <template v-else-if="props.type === 'workOrderAction'"><el-form-item label="流转动作"><el-select v-model="form.action"><el-option label="分配" value="assign"/><el-option label="跟进" value="follow"/><el-option label="仲裁" value="arbitrate"/><el-option label="结案" value="close"/><el-option label="重新打开" value="reopen"/></el-select></el-form-item><el-form-item label="处理人 ID"><el-input v-model="form.assignee_id" /></el-form-item><el-form-item label="处理内容" prop="content"><el-input v-model="form.content" type="textarea" /></el-form-item><el-form-item label="仲裁结论"><el-input v-model="form.arbitration_result" type="textarea" /></el-form-item></template>
      <template v-else-if="props.type === 'batchWorkOrderAction'"><el-form-item label="工单数量"><el-input :model-value="`${form.ids?.length || 0} 个`" disabled /></el-form-item><el-form-item label="批量动作"><el-select v-model="form.action"><el-option label="批量分配" value="assign"/><el-option label="批量跟进" value="follow"/><el-option label="批量仲裁" value="arbitrate"/><el-option label="批量结案" value="close"/><el-option label="批量重开" value="reopen"/></el-select></el-form-item><el-form-item label="处理人 ID"><el-input v-model="form.assignee_id" /></el-form-item><el-form-item label="处理内容" prop="content"><el-input v-model="form.content" type="textarea" /></el-form-item><el-form-item label="仲裁结论"><el-input v-model="form.arbitration_result" type="textarea" /></el-form-item></template>
      <template v-else-if="props.type === 'workOrderEvidence'"><el-form-item label="证据类型"><el-select v-model="form.evidence_type"><el-option v-for="item in ['track','audio','chat','payment','image','text']" :key="item" :label="item" :value="item"/></el-select></el-form-item><el-form-item label="证据地址"><el-input v-model="form.evidence_url" /></el-form-item><el-form-item label="证据说明" prop="content"><el-input v-model="form.content" type="textarea" /></el-form-item></template>
      <template v-else-if="props.type === 'riskHitAction'"><el-form-item label="命中数量"><el-input :model-value="`${form.ids?.length || 0} 条`" disabled /></el-form-item><el-form-item label="处置动作"><el-select v-model="form.risk_action"><el-option label="复核通过" value="review_pass"/><el-option label="加入黑名单" value="add_blacklist"/><el-option label="转工单" value="create_work_order"/></el-select></el-form-item><el-form-item v-if="form.risk_action === 'create_work_order'" label="工单标题"><el-input v-model="form.work_order_title" /></el-form-item><el-form-item v-if="form.risk_action === 'create_work_order'" label="优先级"><el-select v-model="form.priority"><el-option label="低" :value="1"/><el-option label="中" :value="2"/><el-option label="高" :value="3"/><el-option label="紧急" :value="4"/></el-select></el-form-item><el-form-item label="处置说明" prop="reason"><el-input v-model="form.reason" type="textarea" /></el-form-item></template>
      <template v-else-if="props.type === 'createBlacklist'"><el-form-item label="目标类型"><el-select v-model="form.target_type"><el-option label="用户" value="user"/><el-option label="司机" value="driver"/><el-option label="设备" value="device"/></el-select></el-form-item><el-form-item label="目标 ID" prop="target_id"><el-input v-model="form.target_id" /></el-form-item><el-form-item label="拉黑原因" prop="reason"><el-input v-model="form.reason" type="textarea" /></el-form-item></template>
      <template v-else-if="props.type === 'createExport'"><el-form-item label="导出类型"><el-select v-model="form.export_type"><el-option label="用户" value="users"/><el-option label="司机" value="drivers"/><el-option label="订单" value="orders"/><el-option label="操作日志" value="operation_logs"/><el-option label="统计数据" value="statistics"/></el-select></el-form-item><el-form-item label="时间范围"><el-date-picker v-model="form.export_range" type="datetimerange" value-format="YYYY-MM-DD HH:mm:ss" /></el-form-item><el-row :gutter="16"><el-col :span="12"><el-form-item label="用户 ID"><el-input v-model="form.export_user_id" /></el-form-item></el-col><el-col :span="12"><el-form-item label="司机 ID"><el-input v-model="form.export_driver_id" /></el-form-item></el-col></el-row><el-row :gutter="16"><el-col :span="12"><el-form-item label="订单 ID"><el-input v-model="form.export_order_id" /></el-form-item></el-col><el-col :span="12"><el-form-item label="管理员 ID"><el-input v-model="form.export_admin_id" /></el-form-item></el-col></el-row><el-form-item label="城市编码"><el-input v-model="form.export_city_code" /></el-form-item></template>
      <template v-else-if="isReasonAction"><el-form-item label="原因" prop="reason"><el-input v-model="form.reason" type="textarea" /></el-form-item><el-form-item v-if="props.type !== 'release'" label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item></template>
      <template v-else-if="['approve','reject'].includes(props.type)"><el-form-item label="审核备注"><el-input v-model="form.remark" type="textarea" /></el-form-item></template>
      <template v-else><p class="operation-confirm">确认执行此操作？</p></template>
    </el-form>
    <template #footer><el-button :disabled="loading" @click="close">取消</el-button><el-button type="primary" :loading="loading" @click="submit">确认提交</el-button></template>
  </el-dialog>
</template>

<style scoped>
.business-form :deep(.el-select),.business-form :deep(.el-date-editor){width:100%}.business-form :deep(.el-input-number){width:180px}.field-suffix{margin-left:8px;color:var(--muted-color,#8b88a3)}.operation-confirm{margin:12px 0;color:var(--text-color,#2e2c4e)}
</style>
