package repository

import (
	"database/sql"
	"time"
)

// timeLayout 是后台接口统一输出时间格式。
const timeLayout = "2006-01-02 15:04:05"

// FormatTime 将数据库时间转换为前端易读字符串。
func FormatTime(t time.Time) string {
	return t.Format(timeLayout)
}

// FormatOptionalTime 将可空时间转换为字符串。
// 空值返回空字符串，避免前端收到无意义的零时间。
func FormatOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(timeLayout)
}

// formatNullTime 将 sql.NullTime 转换为字符串。
func formatNullTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(timeLayout)
}
