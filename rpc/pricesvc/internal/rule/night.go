package rule

import (
	"time"
)

// IsNightTime 判断给定时刻是否处于夜间时段。
// start/end 为 "HH:MM:SS" 格式，支持跨天时段（如 23:00:00 ~ 05:00:00）。
// 若 start 或 end 为空，表示无夜间费，返回 false。
func IsNightTime(t time.Time, start, end string) bool {
	if start == "" || end == "" {
		return false
	}

	st, err1 := parseClock(start)
	en, err2 := parseClock(end)
	if err1 != nil || err2 != nil {
		return false
	}

	// 当前时刻（当天第几分钟）
	cur := t.Hour()*60 + t.Minute()

	if st < en {
		// 同时段内：start <= cur < end
		return cur >= st && cur < en
	}
	// 跨天：cur >= start 或 cur < end
	return cur >= st || cur < en
}

// parseClock 将 "HH:MM:SS" 解析为当天第几分钟。
func parseClock(s string) (int, error) {
	t, err := time.Parse("15:04:05", s)
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}
