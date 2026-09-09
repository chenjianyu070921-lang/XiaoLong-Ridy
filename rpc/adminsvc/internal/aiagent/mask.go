package aiagent

import "strings"

// maskPhone 脱敏手机号，仅保留前 3 位与后 4 位。
// 用于演示快照或模板报告中可能出现的演示数据，以及防御性处理。
func maskPhone(s string) string {
	s = strings.TrimSpace(s)
	if len(s) < 7 {
		return "***"
	}
	return s[:3] + "****" + s[len(s)-4:]
}

// maskIDCard 脱敏身份证号，仅保留前 3 位与后 4 位。
func maskIDCard(s string) string {
	s = strings.TrimSpace(s)
	if len(s) < 8 {
		return "***"
	}
	return s[:3] + "***********" + s[len(s)-4:]
}
