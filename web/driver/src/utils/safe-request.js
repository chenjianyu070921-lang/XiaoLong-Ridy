/**
 * 司机端报错策略（全端唯一约定，改动前请先读 api/request.js 的响应拦截器）：
 *
 * 1. 提示只由「拦截器」负责。任何请求失败（业务错误码 / HTTP 错误 / 401）
 *    都会在拦截器弹 Toast，调用方无需自己提示。
 * 2. 业务层需要定制文案时，调用 api 时传 `{ silentError: true }` 关掉拦截器提示，
 *    再自行 catch + showToast。二者互斥，物理上不可能重复提示。
 * 3. 因此本文件的 safeApiCall 只负责把失败「吞」成 null，绝不弹提示。
 *
 * 反面教材：此前 safeApiCall 自带一套 options.silent 与 fallbackMessage，
 * 与拦截器的 silentError 是两套互不相通的开关，非静默调用会连弹两次 Toast；
 * 且不接受 config 的 api 函数（如 heartbeatDriver）连「想静默」都做不到。
 */

export function safeApiCall(task) {
  return Promise.resolve()
    .then(task)
    .catch(() => null)
}

export function apiErrorMessage(error, fallbackMessage = '请求失败') {
  return error?.response?.data?.message || error?.message || fallbackMessage
}
