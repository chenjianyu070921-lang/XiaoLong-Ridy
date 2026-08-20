// Package logic 实现 driver API 的业务逻辑层，负责入参校验与调用下游 driversvc。
package logic

import (
	"errors" // 用于定义业务错误变量
	"regexp"  // 用于手机号与身份证的正则匹配
)

// 业务错误定义，供 handler 层映射为 HTTP 错误码。
var (
	// ErrDriverClientNotConfigured 表示 driversvc 客户端未配置（连接失败）。
	ErrDriverClientNotConfigured = errors.New("driver client not configured")
	// ErrOrderClientNotConfigured 表示 ordersvc 客户端未配置（连接失败）。
	ErrOrderClientNotConfigured = errors.New("order client not configured")
	// ErrInvalidParam 表示通用参数非法错误（当前未直接使用，保留扩展）。
	ErrInvalidParam = errors.New("invalid param")
)

// phoneRegexp 是中国大陆手机号的正则：1 开头，第二位 3-9，后接 9 位数字。
var phoneRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

// validPhone 校验字符串是否符合中国大陆手机号格式。
func validPhone(phone string) bool {
	// 使用预编译正则进行匹配。
	return phoneRegexp.MatchString(phone)
}

// validIDCard 校验 18 位身份证号基础格式（前 17 位为数字，末位为数字或 X）。
func validIDCard(no string) bool {
	// 使用正则匹配 17 位数字 + 末位数字或 X/x。
	matched, _ := regexp.MatchString(`^\d{17}[\dXx]$`, no)
	return matched
}

// clampPage 将分页参数收敛到合法范围：page 至少为 1，pageSize 限制在 1~100。
func clampPage(page, pageSize int32) (int32, int32) {
	// 页码小于 1 时归正为 1。
	if page < 1 {
		page = 1
	}
	// 每页条数小于 1 时使用默认 20。
	if pageSize < 1 {
		pageSize = 20
	}
	// 每页条数超过上限 100 时收敛为 100，防止一次拉取过多。
	if pageSize > 100 {
		pageSize = 100
	}
	// 返回修正后的页码与每页条数。
	return page, pageSize
}
