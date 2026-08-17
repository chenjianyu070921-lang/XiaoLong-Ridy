package logic

import "regexp"

var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

// IsValidPhone 判断手机号是否符合国内手机号基础格式。
func IsValidPhone(phone string) bool {
	return phonePattern.MatchString(phone)
}

// MaskPhone 对手机号中间四位脱敏，用于接口返回和默认昵称。
func MaskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}
