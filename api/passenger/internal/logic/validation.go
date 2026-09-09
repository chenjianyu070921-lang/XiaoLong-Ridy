package logic

import (
	"regexp"
	"strings"
)

var passengerPhonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

// isValidPassengerPhone 校验国内手机号基础格式，用于 API 层提前拦截明显错误。
func isValidPassengerPhone(phone string) bool {
	return passengerPhonePattern.MatchString(strings.TrimSpace(phone))
}

// isValidLongitudeLatitude 校验经纬度范围，并拒绝业务上无意义的 0 坐标。
func isValidLongitudeLatitude(longitude, latitude float64) bool {
	if longitude == 0 || latitude == 0 {
		return false
	}
	return longitude >= -180 && longitude <= 180 && latitude >= -90 && latitude <= 90
}
