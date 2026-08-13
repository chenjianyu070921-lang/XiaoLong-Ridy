package logic

import (
	"errors"
	"regexp"
)

// 业务错误定义，供 handler 层映射为 HTTP 错误码。
var (
	ErrDriverClientNotConfigured = errors.New("driver client not configured")
	ErrInvalidParam              = errors.New("invalid param")
)

var phoneRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

// validPhone 校验中国大陆手机号格式。
func validPhone(phone string) bool {
	return phoneRegexp.MatchString(phone)
}

// validIDCard 校验 18 位身份证号基础格式（末位可为 X）。
func validIDCard(no string) bool {
	matched, _ := regexp.MatchString(`^\d{17}[\dXx]$`, no)
	return matched
}

// clampPage 将分页参数收敛到合法范围：page>=1，pageSize 1~100。
func clampPage(page, pageSize int32) (int32, int32) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
