<script setup>
// 管理后台统一业务工作区：根据当前路由选择真实接口、字段列和可用操作。
// 页面不保存业务副本，所有查询和写操作完成后重新从后台读取，避免显示过期状态。
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, Search, Plus, Refresh, Tickets, CircleCheck, User, Avatar, Money } from '@element-plus/icons-vue'
import { ordersApi, marketingApi, statisticsApi, riskApi, logsApi, refundRetryApi, notificationOutboxApi } from '../../api/modules'
import { text, pageData } from '../../utils/format'
import { orderStatusText, userStatusText, driverStatusText, vehicleStatusText, auditStatusText, carTypeText, roleText } from '../../utils/enums'
import BusinessFormDialog from '../../components/BusinessFormDialog.vue'

const route = useRoute(); const router = useRouter(); const loading = ref(false); const rows = ref([]); const total = ref(0); const page = ref(1); const pageSize = ref(20); const detail = ref(null); const dialog = ref(false); const dialogType = ref(''); const dialogRecord = ref({}); const submitting = ref(false); const downloading = ref(false); const selectionRows = ref([])
const filterState = reactive({})
const title = computed(() => route.meta.title || '工作台');
const kind = computed(() => route.path.split('/')[1] || 'dashboard');
// routeKey 用完整静态路径区分普通订单与异常订单，避免被首段 orders 误判。
const routeKey = computed(() => route.path === '/orders/abnormal' ? 'orders/abnormal' : kind.value)
// filters 按路由保存筛选值，避免不同列表复用 status 等同名字段而发生条件串用。
const filters = computed(() => {
  if (!filterState[routeKey.value]) filterState[routeKey.value] = {}
  return filterState[routeKey.value]
})
const isAbnormalOrders = computed(() => routeKey.value === 'orders/abnormal')
const isStatistics = computed(() => kind.value === 'statistics')
const isDetail = computed(() => !!route.params.id || !!route.params.taskNo)
const columns = computed(() => {
  const map = {
    orders: [['order_no','订单号'],['user_id','用户'],['driver_id','司机'],['from_address','出发地'],['to_address','目的地'],['estimated_price','预估金额'],['status','状态'],['created_at','下单时间']],
    'orders/abnormal': [['order_no','订单号'],['abnormal_type','异常类型'],['abnormal_reason','异常原因'],['payment_status','支付状态'],['dispatch_status','派单状态'],['status','订单状态'],['created_at','下单时间']],
    'refund-retry-tasks': [['order_id','订单 ID'],['order_no','订单号'],['refund_no','退款单号'],['refund_cents','退款金额(分)'],['attempt','重试次数'],['next_retry_at','下次重试']],
    coupons: [['name','券名称'],['type','类型'],['face_value','面额'],['threshold_amount','门槛'],['total_count','库存'],['received_count','已发放'],['status','状态']],
    'coupon-issue-tasks': [['task_no','任务号'],['coupon_id','优惠券'],['target_type','目标'],['total_count','总量'],['success_count','成功'],['fail_count','失败'],['status','状态']],
    'price-rules': [['name','规则名称'],['city_code','城市'],['car_type','车型'],['base_price','起步价'],['status','状态'],['effective_at','生效时间']],
    'promotion-activities': [['name','活动名称'],['type','类型'],['start_at','开始时间'],['end_at','结束时间'],['status','状态']],
    'export-tasks': [['task_no','任务号'],['export_type','类型'],['status','状态'],['failure_reason','失败原因'],['created_at','创建时间']],
    blacklist: [['id','记录号'],['target_type','目标类型'],['target_id','目标 ID'],['reason','原因'],['status','状态'],['created_at','加入时间']],
    'risk-hits': [['id','记录号'],['target_type','目标类型'],['target_id','目标 ID'],['scene','场景'],['risk_level','风险等级'],['hit_reason','命中原因'],['handle_status','处置状态'],['blacklist_id','黑名单 ID'],['work_order_id','工单 ID'],['handled_by','处置人'],['handled_at','处置时间'],['created_at','时间']],
    'notification-outbox': [['event_no','事件号'],['module','模块'],['action','动作'],['target_type','目标类型'],['target_id','目标 ID'],['status','补偿状态'],['retry_count','重试次数'],['failure_reason','失败原因'],['updated_at','更新时间']],
    'operation-logs': [['id','日志号'],['admin_id','管理员'],['module','模块'],['action','动作'],['target_type','目标'],['detail','详情'],['created_at','时间']],
  }
  // 异常订单的 kind 仍为 orders，必须使用完整 routeKey 才能命中异常订单列定义。
  return map[routeKey.value] || []
})
const api = computed(() => ({ orders: ordersApi.list, 'orders/abnormal': ordersApi.abnormal, 'refund-retry-tasks': refundRetryApi.list, coupons: marketingApi.coupons, 'coupon-issue-tasks': marketingApi.issueTasks, 'price-rules': marketingApi.priceRules, 'promotion-activities': marketingApi.activities, 'export-tasks': statisticsApi.exports, blacklist: riskApi.blacklist, 'risk-hits': riskApi.hits, 'notification-outbox': notificationOutboxApi.list, 'operation-logs': logsApi.list }[routeKey.value] || (() => Promise.resolve({ list: [], total: 0 }))))
// filterSchemas 将每个列表已开放的查询参数映射为字段化控件，避免模板持续增加平行条件分支。
const filterSchemas = {
  orders: [{ key: 'keyword', type: 'text', placeholder: '订单号' }, { key: 'status', type: 'select', placeholder: '订单状态', options: [[1, '待接单'], [2, '已接单'], [3, '行程中'], [4, '待支付'], [5, '已完成'], [6, '已取消']] }, { key: 'user_id', type: 'number', placeholder: '用户 ID' }, { key: 'driver_id', type: 'number', placeholder: '司机 ID' }, { key: 'date_range', type: 'date-range', placeholder: '下单时间' }],
  'orders/abnormal': [{ key: 'keyword', type: 'text', placeholder: '订单号' }, { key: 'abnormal_type', type: 'select', placeholder: '异常类型', options: [['cancel', '取消异常'], ['payment', '支付异常'], ['dispatch', '派单异常']] }, { key: 'user_id', type: 'number', placeholder: '用户 ID' }, { key: 'driver_id', type: 'number', placeholder: '司机 ID' }, { key: 'date_range', type: 'date-range', placeholder: '下单时间' }],
  'refund-retry-tasks': [],
  coupons: [{ key: 'keyword', type: 'text', placeholder: '优惠券名称' }, { key: 'type', type: 'select', placeholder: '优惠券类型', options: [[1, '满减券'], [2, '折扣券'], [3, '立减券']] }, { key: 'status', type: 'select', placeholder: '优惠券状态', options: [[1, '草稿'], [2, '启用'], [3, '停用']] }, { key: 'date_range', type: 'date-range', placeholder: '创建时间' }],
  'coupon-issue-tasks': [{ key: 'coupon_id', type: 'number', placeholder: '优惠券 ID' }, { key: 'status', type: 'select', placeholder: '任务状态', options: [['pending', '待执行'], ['processing', '执行中'], ['success', '成功'], ['failed', '失败']] }, { key: 'date_range', type: 'date-range', placeholder: '创建时间' }],
  'price-rules': [{ key: 'keyword', type: 'text', placeholder: '规则名称' }, { key: 'city_code', type: 'text', placeholder: '城市编码' }, { key: 'car_type', type: 'number', placeholder: '车型' }, { key: 'status', type: 'select', placeholder: '规则状态', options: [[1, '启用'], [2, '停用']] }],
  'operation-logs': [{ key: 'admin_id', type: 'number', placeholder: '管理员 ID' }, { key: 'module', type: 'text', placeholder: '模块名' }, { key: 'action', type: 'text', placeholder: '操作动作' }, { key: 'target_type', type: 'text', placeholder: '目标类型' }, { key: 'target_id', type: 'number', placeholder: '目标 ID' }, { key: 'date_range', type: 'date-range', placeholder: '操作时间' }],
  'risk-hits': [{ key: 'target_type', type: 'select', placeholder: '目标对象', options: [['user', '用户'], ['driver', '司机'], ['device', '设备'], ['phone', '手机号']] }, { key: 'target_id', type: 'number', placeholder: '目标 ID' }, { key: 'scene', type: 'text', placeholder: '场景，如 login' }, { key: 'risk_level', type: 'select', placeholder: '风险等级', options: [[1, '低风险'], [2, '中风险'], [3, '高风险']] }],
  'notification-outbox': [{ key: 'status', type: 'select', placeholder: '补偿状态', options: [['pending', '待重试'], ['running', '处理中'], ['success', '已完成'], ['failed', '失败']] }, { key: 'module', type: 'text', placeholder: '模块' }, { key: 'action', type: 'text', placeholder: '动作' }, { key: 'target_id', type: 'number', placeholder: '目标 ID' }],
  blacklist: [{ key: 'target_type', type: 'select', placeholder: '目标类型', options: [['user', '用户'], ['driver', '司机'], ['device', '设备'], ['phone', '手机号']] }, { key: 'target_id', type: 'number', placeholder: '目标 ID' }, { key: 'status', type: 'select', placeholder: '状态', options: [[1, '生效'], [2, '已解除']] }],
}
const activeFilters = computed(() => filterSchemas[routeKey.value] || [])
// listParams 仅透传当前页面已有值的字段；日期控件转换为接口约定的 start_time 和 end_time。
const listParams = () => activeFilters.value.reduce((params, field) => {
  const value = filters.value[field.key]
  if (field.type === 'date-range') {
    if (Array.isArray(value) && value.length === 2) Object.assign(params, { start_time: value[0], end_time: value[1] })
  } else if (field.type === 'number') {
    // 关联 ID 未填写或不是正数时不传给后端，避免控件默认值改变查询结果。
    const numberValue = Number(value)
    if (Number.isInteger(numberValue) && numberValue > 0) params[field.key] = numberValue
  } else if (value !== '' && value !== undefined && value !== null) params[field.key] = value
  return params
}, { page: page.value, page_size: pageSize.value })
// syncRouteQuery 接收跨模块跳转带来的筛选条件，仅同步当前列表已声明的字段，保证订单、风控、黑名单等页面都能按目标对象定位。
const syncRouteQuery = () => {
  activeFilters.value.forEach((field) => {
    const queryValue = route.query[field.key]
    if (Array.isArray(queryValue) || queryValue === undefined || queryValue === null || queryValue === '') return
    filters.value[field.key] = field.type === 'number' ? Number(queryValue) : queryValue
  })
}
// statusMaps 按路由隔离各列表的 status/type 等枚举文案，避免订单语义串用到优惠券、计价规则等模块。
const statusMaps = {
  orders: { status: { 1: '待接单', 2: '已接单', 3: '行程中', 4: '待支付', 5: '已完成', 6: '已取消' }, type: { 1: '满减', 2: '折扣', 3: '立减' } },
  'orders/abnormal': { status: { 1: '待接单', 2: '已接单', 3: '行程中', 4: '待支付', 5: '已完成', 6: '已取消' }, payment_status: { 1: '待支付', 2: '已支付', 3: '支付异常' }, dispatch_status: { 1: '派单中', 2: '已接受', 3: '派单异常', 4: '改派失败', 5: '已取消派单' }, abnormal_type: { cancel: '取消异常', payment: '支付异常', dispatch: '派单异常' } },
  coupons: { status: { 1: '草稿', 2: '启用', 3: '停用' }, type: { 1: '满减券', 2: '折扣券', 3: '立减券' } },
  'coupon-issue-tasks': { status: { pending: '待执行', processing: '执行中', success: '成功', failed: '失败' } },
  'price-rules': { status: { 1: '启用', 2: '停用' } },
  'promotion-activities': { status: { 1: '未开始', 2: '进行中', 3: '已结束' }, type: { 1: '拉新', 2: '折扣', 3: '立减' } },
  'export-tasks': { status: { pending: '待生成', running: '生成中', success: '已完成', failed: '失败', canceled: '已取消' } },
  blacklist: { status: { 1: '生效', 2: '已解除' } },
  'risk-hits': { handle_status: { pending: '待处理', review_pass: '复核通过', blacklisted: '已拉黑', work_order: '已转工单' }, risk_level: { 1: '低', 2: '中', 3: '高' } },
  'notification-outbox': { status: { pending: '待重试', running: '处理中', success: '已完成', failed: '失败' } },
}
// label 取当前路由对应的枚举映射；键不在映射中时回退原始值，保证未知状态显示为可读而非数字。
// enumResolvers 统一复用全局枚举定义，避免不同列表对同一个状态码产生串号；模块专属枚举再回退到本页映射。
const enumResolvers = {
  orders: { status: orderStatusText, car_type: carTypeText },
  'orders/abnormal': { status: orderStatusText },
  users: { status: userStatusText, role: roleText },
  drivers: { status: driverStatusText, vehicle_status: vehicleStatusText, audit_status: auditStatusText },
}
const label = (key, value) => {
  const resolver = enumResolvers[routeKey.value]?.[key]
  if (resolver) return resolver(value)
  const m = statusMaps[routeKey.value]?.[key]
  return m?.[value] || text(value)
}
const load = async () => { if (kind.value === 'dashboard' || isStatistics.value || isDetail.value) return; loading.value = true; try { const data = await api.value(listParams()); const p = pageData(data); rows.value = p.list; total.value = p.total } finally { loading.value = false } }
// reset 只清空当前列表声明的筛选字段，避免跨路由残留条件并保留其他页面的独立状态。
const reset = () => { activeFilters.value.forEach((field) => { filters.value[field.key] = field.type === 'date-range' ? [] : undefined }); page.value = 1; load() }
const loadDetail = async () => { const id = route.params.id; if (kind.value === 'orders') detail.value = await ordersApi.detail(id); else if (kind.value === 'price-rules') detail.value = await marketingApi.priceRule(id); else if (kind.value === 'export-tasks') detail.value = await statisticsApi.exportDetail(route.params.taskNo); }
// detailableKinds 指有独立详情路由或详情接口的模块；其余模块"查看"以行内弹窗展示字段，避免被通配路由重定向回工作台。
const detailableKinds = ['orders', 'price-rules', 'export-tasks', 'refund-retry-tasks']
const rowDetail = ref(null)
const openDetail = (row) => {
  if (isAbnormalOrders.value) { rowDetail.value = row; return }
  if (detailableKinds.includes(kind.value)) {
    // 退款补偿任务列表没有独立 id，使用关联订单 ID 打开订单详情。
    const id = kind.value === 'refund-retry-tasks' ? row.order_id : (row.id || row.task_no)
    if (!id) { rowDetail.value = row; return }
    const path = kind.value === 'export-tasks' ? `/export-tasks/${id}` : kind.value === 'refund-retry-tasks' ? `/orders/${id}` : `/${kind.value}/${id}`
    router.push(path)
    return
  }
  rowDetail.value = row
}
// 风控命中与通知补偿列表没有独立详情路由，统一使用行详情抽屉展示。
const canOpenRowDetail = computed(() => ['risk-hits', 'notification-outbox'].includes(kind.value))
const rowDetailEntries = computed(() => Object.entries(rowDetail.value || {}).filter(([key, value]) => value !== null && value !== undefined && value !== '').map(([key, value]) => ({ key, value: typeof value === 'object' ? JSON.stringify(value) : label(key, value) })))
// openRelated 将风控命中的派生关联对象带回对应运营页面，形成从命中到处置结果的追踪链路。
const openRelated = (type, row) => {
  const id = type === 'work-order' ? row?.work_order_id : row?.target_id
  if (!id || id <= 0) return
  router.push(type === 'work-order'
    ? `/work-orders/${id}`
    : `/blacklist?target_type=${encodeURIComponent(row.target_type || '')}&target_id=${id}`)
}
// action 仅记录操作类型和当前行，字段预填、校验与请求体转换全部下沉至业务表单组件。
const action = (name, row = {}) => { dialogType.value = name; dialogRecord.value = row; dialog.value = true }
// retryRefundTask 触发退款补偿任务立即进入 job 扫描窗口，成功后刷新列表。
const retryRefundTask = async (row) => { try { await ElMessageBox.confirm(`确认立即重试退款任务 ${row.refund_no}？`, '退款补偿', { type: 'warning' }); await refundRetryApi.retry(row.refund_no); ElMessage.success('已触发重试'); await load() } catch (error) { if (error !== 'cancel' && error !== 'close') throw error } }
const selectedIDs = computed(() => selectionRows.value.map((row) => row.id).filter(Boolean))
// batchAction 将当前表格勾选项转换为显式 ID 列表，避免批量接口依赖前端筛选条件重复计算。
const batchAction = (name) => { if (!selectedIDs.value.length) return ElMessage.warning('请先勾选需要处理的记录'); action(name, { ids: selectedIDs.value }) }
const createTypes = { coupons: 'createCoupon', 'price-rules': 'createPriceRule', 'promotion-activities': 'createActivity', blacklist: 'createBlacklist', 'export-tasks': 'createExport' }
const openCreate = () => action(createTypes[kind.value])
const dialogTitle = computed(() => ({ createCoupon: '新建优惠券模板', editCoupon: '编辑优惠券模板', issueCoupon: '发放优惠券', createPriceRule: '新建计价规则', editPriceRule: '编辑计价规则', createActivity: '新建活动', editActivity: '编辑活动', riskHitAction: '风控命中处置', createBlacklist: '加入黑名单', createExport: '创建导出任务', cancel: '取消订单', release: '解除黑名单', disableCoupon: '停用优惠券', enableRule: '启用计价规则', disableRule: '停用计价规则', publish: '发布活动', rollback: '回滚活动' }[dialogType.value] || '确认操作'))
// submitAction 只按类型调度现有接口；组件已将字段化值转换为契约要求的 payload。
const submitAction = async ({ payload, record }) => { submitting.value = true; try { const id = record?.id; const handlers = { createCoupon: () => marketingApi.createCoupon(payload), editCoupon: () => marketingApi.updateCoupon(id, payload), issueCoupon: () => marketingApi.issueCoupon(id, payload), disableCoupon: () => marketingApi.disableCoupon(id), createPriceRule: () => marketingApi.createPriceRule(payload), editPriceRule: () => marketingApi.updatePriceRule(id, payload), enableRule: () => marketingApi.enablePriceRule(id), disableRule: () => marketingApi.disablePriceRule(id), createActivity: () => marketingApi.createActivity(payload), editActivity: () => marketingApi.updateActivity(id, payload), publish: () => marketingApi.publishActivity(id, payload), rollback: () => marketingApi.rollbackActivity(id, { publish_scope: '', target_config: '{}' }), riskHitAction: () => riskApi.handleHits(payload), createBlacklist: () => riskApi.add(payload), release: () => riskApi.release(id, payload), createExport: () => statisticsApi.createExport(payload), cancel: () => ordersApi.cancel(id, payload) }; await handlers[dialogType.value]?.(); dialog.value = false; selectionRows.value = []; ElMessage.success('操作成功'); if (isDetail.value) await loadDetail(); else await load() } finally { submitting.value = false } }
// isExportReady 兼容后端数值与文本状态，只有异步导出完成后才显示下载入口。
const isExportReady = computed(() => String(detail.value?.status) === 'success')
// downloadExport 通过受保护的下载接口获取 CSV Blob，浏览器端只使用任务号生成文件名。
const downloadExport = async () => { downloading.value = true; try { const response = await statisticsApi.downloadExport(route.params.taskNo); const objectURL = URL.createObjectURL(response.data); const link = document.createElement('a'); link.href = objectURL; link.download = `${route.params.taskNo}.csv`; document.body.appendChild(link); link.click(); link.remove(); URL.revokeObjectURL(objectURL) } finally { downloading.value = false } }
// detailEntries 显式列出详情字段并过滤空值，防止把服务端对象直接以 JSON 暴露给运营人员。
const detailEntries = computed(() => Object.entries(detail.value || {}).filter(([key, value]) => value !== null && value !== undefined && value !== '').map(([key, value]) => ({ key, value: typeof value === 'object' ? JSON.stringify(value) : label(key, value) })))
const dashboard = ref({ overview: {}, orders: {}, coupons: {}, drivers: {}, revenue: {}, users: {} });
const statisticsRange = ref([])
const statisticsError = ref('')
// statisticsMeta 展示后端统一返回的生成时间和数据截至时间，避免运营误以为数据来自离线快照。
const statisticsMeta = computed(() => ({
  generatedAt: dashboard.value.overview?.generated_at || dashboard.value.orders?.generated_at || '',
  dataAsOf: dashboard.value.overview?.data_as_of || dashboard.value.orders?.data_as_of || '',
}))
// chartGroups 将后端聚合指标转换为比例条，避免虚构接口未提供的时间序列数据。
const chartGroups = computed(() => {
  const number = (value) => Math.max(0, Number.parseFloat(value) || 0)
  const percent = (value, total) => total > 0 ? Math.min(100, value / total * 100) : 0
  const overview = dashboard.value.overview || {}
  const orders = dashboard.value.orders || {}
  const drivers = dashboard.value.drivers || {}
  const revenue = dashboard.value.revenue || {}
  const coupons = dashboard.value.coupons || {}
  return [
    { title: '订单结构', items: [
      { label: '已完成', value: number(overview.completed_order_count), percent: percent(number(overview.completed_order_count), number(overview.order_count)), color: '#3fb984' },
      { label: '异常订单', value: number(overview.abnormal_order_count), percent: percent(number(overview.abnormal_order_count), number(overview.order_count)), color: '#e6a23c' },
      { label: '已取消', value: number(orders.canceled_order_count), percent: percent(number(orders.canceled_order_count), number(orders.order_count)), color: '#f56c6c' },
    ] },
    { title: '司机指标', items: [
      { label: '司机总量', value: number(overview.driver_count), percent: 100, color: '#409eff' },
      { label: '新增司机', value: number(drivers.new_driver_count), percent: percent(number(drivers.new_driver_count), number(overview.driver_count)), color: '#67c23a' },
      { label: '待审核', value: number(drivers.pending_audit_count), percent: percent(number(drivers.pending_audit_count), number(overview.driver_count)), color: '#e6a23c' },
    ] },
    { title: '收入指标', items: [
      { label: '实付金额', value: revenue.paid_amount || 0, percent: 100, color: '#409eff' },
      { label: '退款金额', value: revenue.refund_amount || 0, percent: percent(number(revenue.refund_amount), number(revenue.paid_amount)), color: '#f56c6c' },
      { label: '平台抽佣', value: revenue.platform_commission || 0, percent: percent(number(revenue.platform_commission), number(revenue.paid_amount)), color: '#9b59b6' },
    ] },
    { title: '优惠券指标', items: [
      { label: '已发放', value: number(coupons.issued_coupon_count), percent: percent(number(coupons.issued_coupon_count), number(coupons.coupon_count)), color: '#409eff' },
      { label: '已使用', value: number(coupons.used_coupon_count), percent: percent(number(coupons.used_coupon_count), number(coupons.issued_coupon_count)), color: '#3fb984' },
      { label: '已过期', value: number(coupons.expired_coupon_count), percent: percent(number(coupons.expired_coupon_count), number(coupons.issued_coupon_count)), color: '#909399' },
    ] },
  ]
})
// statisticsParams 统一生成统计接口时间范围；未选择时默认近 30 天。
const statisticsParams = () => { if (statisticsRange.value?.length === 2) return { start_time: statisticsRange.value[0], end_time: statisticsRange.value[1] }; const end = new Date(); const start = new Date(end); start.setDate(end.getDate() - 30); const fmt = (d) => d.toISOString().slice(0, 19).replace('T', ' '); return { start_time: fmt(start), end_time: fmt(end) } }
// loadDashboard 并发请求后台已开放的真实统计接口，工作台和数据统计页共用同一份数据。
// loadDashboard 并发读取所有统计域；失败时保留局部错误状态，页面提供显式重试入口。
const loadDashboard = async () => {
  loading.value = true
  statisticsError.value = ''
  const params = statisticsParams()
  // 使用 allSettled 分块处理统计接口，单个指标失败时保留其他已成功加载的数据。
  const requests = [
    ['overview', statisticsApi.overview(params)],
    ['orders', statisticsApi.orders(params)],
    ['drivers', statisticsApi.drivers(params)],
    ['revenue', statisticsApi.revenue(params)],
    ['coupons', statisticsApi.coupons(params)],
    ['users', statisticsApi.users(params)],
  ]
  const results = await Promise.allSettled(requests.map(([, request]) => request))
  const failed = []
  results.forEach((result, index) => {
    const [key] = requests[index]
    if (result.status === 'fulfilled') {
      dashboard.value[key] = result.value || {}
      return
    }
    failed.push(key)
  })
  if (failed.length) {
    statisticsError.value = `部分统计数据加载失败（${failed.join('、')}），可点击重试`
  }
  loading.value = false
}
onMounted(async () => { syncRouteQuery(); if (kind.value === 'dashboard' || isStatistics.value) return loadDashboard(); if (isDetail.value) { await loadDetail(); return }; await load() }); watch(() => route.fullPath, async () => { detail.value = null; selectionRows.value = []; page.value = 1; syncRouteQuery(); if (kind.value === 'dashboard' || isStatistics.value) return loadDashboard(); if (isDetail.value) { await loadDetail(); return }; await load() });
</script>

<template>
  <section class="console" v-loading="loading">
    <template v-if="kind === 'dashboard' || isStatistics">
      <div class="hero"><div><span class="eyebrow">运营数据中心</span><h1>{{ isStatistics ? '数据统计' : '工作台' }}</h1><p>{{ isStatistics ? '按时间范围查看用户、订单与营销核心指标' : '实时掌握用户、订单与营销运营状态' }}</p></div><div class="statistics-actions"><el-date-picker v-if="isStatistics" v-model="statisticsRange" type="datetimerange" value-format="YYYY-MM-DD HH:mm:ss" start-placeholder="开始时间" end-placeholder="结束时间" /><el-button :icon="Refresh" @click="loadDashboard">刷新数据</el-button></div></div>
      <el-alert v-if="statisticsError" class="statistics-error" :title="statisticsError" type="error" show-icon :closable="false"><template #default><el-button link type="danger" @click="loadDashboard">重新加载统计</el-button></template></el-alert>
      <div v-if="statisticsMeta.generatedAt || statisticsMeta.dataAsOf" class="statistics-meta"><span>数据生成时间：{{ statisticsMeta.generatedAt || '-' }}</span><span>数据截至时间：{{ statisticsMeta.dataAsOf || '-' }}</span></div>
      <div class="metric-grid"><div v-for="item in [{k:'order_count',t:'订单总量',i:Tickets},{k:'completed_order_count',t:'完成订单',i:CircleCheck},{k:'user_count',t:'用户总量',i:User},{k:'driver_count',t:'司机总量',i:Avatar},{k:'gmv',t:'GMV',i:Money}]" :key="item.k" class="metric"><span class="metric-icon"><el-icon><component :is="item.i" /></el-icon></span><div class="metric-body"><span>{{ item.t }}</span><strong>{{ text(dashboard.overview[item.k]) }}</strong><small>来自实时统计接口</small></div></div></div>
      <div class="panel-grid"><div class="panel"><div class="panel-title">用户运营</div><div class="stat-line"><span>新增用户</span><b>{{ text(dashboard.users.new_user_count) }}</b></div><div class="stat-line"><span>下单用户</span><b>{{ text(dashboard.users.order_user_count) }}</b></div><div class="stat-line"><span>复购率</span><b>{{ text(dashboard.users.reorder_rate) }}</b></div><div class="stat-line"><span>投诉率</span><b>{{ text(dashboard.users.complaint_rate) }}</b></div></div><div class="panel"><div class="panel-title">订单运营</div><div class="stat-line"><span>完成率</span><b>{{ text(dashboard.orders.completion_rate) }}</b></div><div class="stat-line"><span>取消率</span><b>{{ text(dashboard.orders.cancel_rate) }}</b></div><div class="stat-line"><span>异常订单</span><b>{{ text(dashboard.overview.abnormal_order_count) }}</b></div></div><div class="panel"><div class="panel-title">司机经营</div><div class="stat-line"><span>新增司机</span><b>{{ text(dashboard.drivers.new_driver_count) }}</b></div><div class="stat-line"><span>待审核</span><b>{{ text(dashboard.drivers.pending_audit_count) }}</b></div><div class="stat-line"><span>司机收入</span><b>¥{{ text(dashboard.drivers.driver_income) }}</b></div></div><div class="panel"><div class="panel-title">财务收入</div><div class="stat-line"><span>实付金额</span><b>¥{{ text(dashboard.revenue.paid_amount) }}</b></div><div class="stat-line"><span>退款金额</span><b>¥{{ text(dashboard.revenue.refund_amount) }}</b></div><div class="stat-line"><span>平台抽佣</span><b>¥{{ text(dashboard.revenue.platform_commission) }}</b></div></div><div class="panel"><div class="panel-title">优惠券运营</div><div class="stat-line"><span>优惠券模板</span><b>{{ text(dashboard.coupons.coupon_count) }}</b></div><div class="stat-line"><span>已启用</span><b>{{ text(dashboard.coupons.enabled_coupon_count) }}</b></div><div class="stat-line"><span>使用率</span><b>{{ text(dashboard.coupons.use_rate) }}</b></div></div></div>
      <div class="chart-grid"><div v-for="group in chartGroups" :key="group.title" class="panel chart-panel"><div class="panel-title">{{ group.title }}<small>接口聚合指标</small></div><div v-for="item in group.items" :key="item.label" class="chart-row"><div class="chart-label"><span>{{ item.label }}</span><b>{{ text(item.value) }}</b></div><div class="chart-track"><i :style="{ width: `${item.percent}%`, backgroundColor: item.color }" /></div></div></div></div>
    </template>
    <template v-else-if="isDetail"><div class="page-head"><div><span class="eyebrow">详情视图</span><h1>{{ title }}</h1></div><div class="head-actions"><el-button v-if="kind === 'export-tasks' && isExportReady" type="primary" :icon="Download" :loading="downloading" @click="downloadExport">下载 CSV</el-button><el-button @click="router.back()">返回列表</el-button></div></div><div class="panel detail-grid"><div v-for="item in detailEntries" :key="item.key" class="detail-item"><span>{{ item.key }}</span><strong>{{ item.value }}</strong></div></div></template>
    <template v-else>
      <div class="page-head"><div><span class="eyebrow">管理后台 / {{ title }}</span><h1>{{ title }}</h1></div><div class="head-actions"><el-button v-if="kind==='risk-hits'" type="primary" :disabled="!selectedIDs.length" @click="batchAction('riskHitAction')">批量处置</el-button><el-button :icon="Refresh" @click="load">刷新</el-button><el-button v-if="['coupons','price-rules','promotion-activities','blacklist','export-tasks'].includes(kind)" type="primary" :icon="Plus" @click="openCreate">新建</el-button></div></div>
      <div class="panel filters"><template v-for="field in activeFilters" :key="field.key"><el-input v-if="field.type === 'text'" v-model="filters[field.key]" :placeholder="field.placeholder" clearable :prefix-icon="field.key === 'keyword' ? Search : undefined" @keyup.enter="page=1;load" /><el-input v-else-if="field.type === 'number'" v-model="filters[field.key]" type="number" :placeholder="field.placeholder" clearable @keyup.enter="page=1;load" /><el-select v-else-if="field.type === 'select'" v-model="filters[field.key]" :placeholder="field.placeholder" clearable><el-option v-for="option in field.options" :key="option[0]" :label="option[1]" :value="option[0]" /></el-select><el-date-picker v-else-if="field.type === 'date-range'" v-model="filters[field.key]" type="datetimerange" value-format="YYYY-MM-DD HH:mm:ss" :start-placeholder="`${field.placeholder}开始`" :end-placeholder="`${field.placeholder}结束`" /></template><el-button type="primary" @click="page=1;load">查询</el-button><el-button @click="reset">重置</el-button></div>
      <el-alert v-if="kind === 'notification-outbox'" class="outbox-note" title="补偿任务由 job 自动抢占和重试，本页面只展示状态与失败原因。" type="info" show-icon :closable="false" />
      <div class="panel table-panel"><el-table :data="rows" stripe empty-text="暂无数据" @selection-change="selectionRows=$event"><el-table-column v-if="kind==='risk-hits'" type="selection" width="46" /><el-table-column v-for="c in columns" :key="c[0]" :prop="c[0]" :label="c[1]" min-width="140"><template #default="scope"><template v-if="kind==='risk-hits' && c[0]==='blacklist_id' && scope.row[c[0]]"><el-button link type="primary" @click="openRelated('blacklist', scope.row)">查看黑名单</el-button></template><template v-else-if="kind==='risk-hits' && c[0]==='work_order_id' && scope.row[c[0]]"><el-button link type="primary" @click="openRelated('work-order', scope.row)">查看工单</el-button></template><span v-else :class="{mono:c[0].includes('id') || c[0].includes('no')}" >{{ label(c[0], scope.row[c[0]]) }}</span></template></el-table-column><el-table-column label="操作" fixed="right" width="320"><template #default="scope"><el-button v-if="!['risk-hits','notification-outbox'].includes(kind)" link type="primary" @click="openDetail(scope.row)">查看</el-button><el-button v-if="kind==='orders' && !isAbnormalOrders && ![5,6].includes(scope.row.status)" link type="danger" @click="action('cancel',scope.row)">取消订单</el-button><el-button v-if="kind==='risk-hits' && scope.row.handle_status==='work_order'" link type="primary" @click="openRelated('work-order', scope.row)">跟进工单</el-button><el-button v-if="kind==='risk-hits' && scope.row.handle_status==='blacklisted'" link type="primary" @click="openRelated('blacklist', scope.row)">查看黑名单</el-button><el-button v-if="kind==='risk-hits' && scope.row.handle_status==='pending'" link type="warning" @click="action('riskHitAction',{ ids:[scope.row.id], risk_action:'create_work_order', reason:scope.row.hit_reason })">处置</el-button><el-button v-if="kind==='coupons'" link type="primary" @click="action('editCoupon',scope.row)">编辑</el-button><el-button v-if="kind==='coupons' && scope.row.status!=3" link type="success" @click="action('issueCoupon',scope.row)">发券</el-button><el-button v-if="kind==='coupons' && scope.row.status!=3" link type="warning" @click="action('disableCoupon',scope.row)">停用</el-button><el-button v-if="kind==='blacklist' && scope.row.status==1" link type="danger" @click="action('release',scope.row)">解除</el-button><el-button v-if="kind==='price-rules'" link type="primary" @click="action('editPriceRule',scope.row)">编辑</el-button><el-button v-if="kind==='price-rules'" link type="success" @click="action(scope.row.status==1?'disableRule':'enableRule',scope.row)">{{scope.row.status==1?'停用':'启用'}}</el-button><el-button v-if="kind==='promotion-activities'" link type="primary" @click="action('editActivity',scope.row)">编辑</el-button><el-button v-if="kind==='promotion-activities'" link type="warning" @click="action(scope.row.status==2?'rollback':'publish',scope.row)">{{scope.row.status==2?'回滚':'发布'}}</el-button></template></el-table-column></el-table><div class="table-footer"><span>共 {{ total }} 条记录</span><el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="prev, pager, next, sizes" @current-change="load" @size-change="load" /></div></div>
    </template>
    <el-button v-if="kind === 'refund-retry-tasks' && rows.length" type="warning" @click="retryRefundTask(rows[0])">立即重试首条任务</el-button>
    <el-drawer :model-value="!!rowDetail" :title="title + ' 详情'" size="520px" destroy-on-close class="row-detail-drawer" @update:model-value="(v) => { if (!v) rowDetail = null }">
      <div class="panel detail-grid" style="border:none;box-shadow:none">
        <div v-for="item in rowDetailEntries" :key="item.key" class="detail-item"><span>{{ item.key }}</span><strong>{{ item.value }}</strong></div>
      </div>
    </el-drawer>
    <BusinessFormDialog v-model="dialog" :type="dialogType" :title="dialogTitle" :record="dialogRecord" :loading="submitting" @submit="submitAction" />
  </section>
</template>

<style scoped>
.console{color:var(--text-color,#2e2c4e)}.hero,.page-head{display:flex;align-items:flex-end;justify-content:space-between;margin-bottom:22px}.eyebrow{color:var(--brand,#6c5ce7);font-size:12px;letter-spacing:.1em;font-weight:600}.hero h1,.page-head h1{margin:7px 0 4px;font-size:26px;color:var(--text-color,#2e2c4e)}.hero p{margin:0;color:var(--muted-color,#8b88a3)}.statistics-actions{display:flex;align-items:center;gap:10px}.statistics-error{margin-bottom:16px}
.statistics-meta{display:flex;gap:22px;align-items:center;margin:-8px 0 16px;color:var(--muted-color,#8b88a3);font-size:12px}
.metric-grid{display:grid;grid-template-columns:repeat(5,1fr);gap:16px}.metric,.panel{background:var(--panel-bg,#fff);border:1px solid var(--border-color,#e5e4f0);border-radius:14px;box-shadow:var(--card-shadow,none)}.outbox-note{margin-bottom:16px}
.metric{display:flex;align-items:center;gap:16px;padding:20px}.metric-icon{width:52px;height:52px;flex:0 0 52px;border-radius:50%;background:linear-gradient(135deg,var(--brand,#6c5ce7),#9a8ff2);color:#fff;display:flex;align-items:center;justify-content:center;font-size:22px;box-shadow:0 6px 14px rgba(108,92,231,.28)}
.metric-body{min-width:0}.metric-body span,.metric-body small{display:block;color:var(--muted-color,#8b88a3)}.metric-body strong{display:block;color:var(--text-color,#2e2c4e);font-size:26px;margin:6px 0 2px}.metric-body small{font-size:12px}
.panel-grid{display:grid;grid-template-columns:1fr 1fr;gap:16px;margin-top:16px}.panel{padding:20px}.panel-title{font-size:16px;font-weight:700;color:var(--text-color,#2e2c4e);margin-bottom:14px;padding-left:10px;border-left:3px solid var(--brand,#6c5ce7)}
.chart-grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:16px;margin-top:16px}.chart-panel{min-height:180px}.chart-panel .panel-title{display:flex;align-items:center;justify-content:space-between}.chart-panel .panel-title small{color:var(--muted-color,#8b88a3);font-size:11px;font-weight:400}.chart-row{margin:16px 0}.chart-label{display:flex;justify-content:space-between;gap:8px;margin-bottom:7px;color:var(--muted-color,#8b88a3);font-size:12px}.chart-label b{color:var(--text-color,#2e2c4e);font-weight:600}.chart-track{height:8px;overflow:hidden;background:#eef0f5;border-radius:4px}.chart-track i{display:block;height:100%;border-radius:4px;transition:width .3s ease}
.stat-line{display:flex;justify-content:space-between;border-top:1px solid var(--border-color,#e5e4f0);padding:14px 0;color:var(--muted-color,#8b88a3)}.stat-line b{color:var(--brand,#6c5ce7)}
.filters{display:flex;gap:10px;align-items:center;margin-bottom:16px;flex-wrap:wrap}.filters .el-input{width:240px}.filters .el-select,.history-toolbar .el-select{width:160px}
.table-panel{padding:0;overflow:hidden}.table-footer{display:flex;justify-content:space-between;align-items:center;padding:14px 18px;color:var(--muted-color,#8b88a3);border-top:1px solid var(--border-color,#e5e4f0)}
.mono{font-family:ui-monospace;color:var(--brand,#6c5ce7)}
.detail-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:0;padding:0;overflow:hidden}.detail-item{padding:16px 20px;border-bottom:1px solid var(--border-color,#e5e4f0)}.detail-item span{display:block;margin-bottom:7px;color:var(--muted-color,#8b88a3);font-size:12px}.detail-item strong{display:block;overflow-wrap:anywhere;color:var(--text-color,#2e2c4e);font-weight:500}
.head-actions,.history-toolbar{display:flex;gap:10px}.history-toolbar{justify-content:flex-end;margin-bottom:8px}
@media(max-width:1000px){.metric-grid,.detail-grid{grid-template-columns:1fr}.panel-grid,.chart-grid{grid-template-columns:1fr}.filters,.statistics-actions{flex-wrap:wrap}}
</style>
