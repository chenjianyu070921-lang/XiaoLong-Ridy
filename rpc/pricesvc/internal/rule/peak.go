package rule

import "time"

// PeakFactor 高峰时段动态调价倍率。
// 高峰时段复用动态调价 factor 机制，不新增独立计费项。
const PeakFactor = 1.3

// IsPeakTime 判断给定时刻是否处于高峰时段。
// 早高峰 07:00-09:00，晚高峰 17:00-19:00。
func IsPeakTime(t time.Time) bool {
	hour := t.Hour()
	return (hour >= 7 && hour < 9) || (hour >= 17 && hour < 19)
}
