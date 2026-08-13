package logic

import "time"

// formatDate 将时间指针格式化为 YYYY-MM-DD 字符串，空指针返回空串。
func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// unixOrZero 将时间指针转为 Unix 秒，空指针返回 0。
func unixOrZero(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Unix()
}
