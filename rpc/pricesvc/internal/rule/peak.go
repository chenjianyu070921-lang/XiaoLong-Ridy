package rule

import "time"

// IsPeakTime 判断给定时刻是否处于高峰时段。
// 早高峰 07:00-09:00，晚高峰 17:00-19:00。
func IsPeakTime(t time.Time) bool {
	hour := t.Hour()
	return (hour >= 7 && hour < 9) || (hour >= 17 && hour < 19)
}
