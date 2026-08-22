package logic

import (
	"errors"
	"regexp"
	"strings"

	"XiaoLong-Ridy/api/driver/internal/types"
)

var (
	ErrDriverClientNotConfigured = errors.New("driver client not configured")
	ErrOrderClientNotConfigured  = errors.New("order client not configured")
	ErrInvalidParam              = errors.New("invalid param")
	ErrForbiddenDriverResource   = errors.New("forbidden driver resource")
)

var phoneRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

func validPhone(phone string) bool {
	return phoneRegexp.MatchString(phone)
}

func validIDCard(no string) bool {
	matched, _ := regexp.MatchString(`^\d{17}[\dXx]$`, no)
	return matched
}

func validPassword(password string) bool {
	n := len(password)
	return n >= 8 && n <= 72
}

func validDriverStatus(status string) bool {
	switch status {
	case "DRIVER_STATUS_PENDING", "DRIVER_STATUS_NORMAL", "DRIVER_STATUS_FROZEN", "DRIVER_STATUS_CANCELLED":
		return true
	default:
		return false
	}
}

func normalizeCreateDriverRequest(req *types.CreateDriverRequest) {
	if req == nil {
		return
	}
	req.Phone = strings.TrimSpace(req.Phone)
	req.RealName = strings.TrimSpace(req.RealName)
	req.IdCardNo = strings.ToUpper(strings.TrimSpace(req.IdCardNo))
	req.DriverLicenseNo = strings.TrimSpace(req.DriverLicenseNo)
	req.AvatarURL = strings.TrimSpace(req.AvatarURL)
}

func normalizeRegisterDriverRequest(req *types.RegisterDriverRequest) {
	if req == nil {
		return
	}
	req.Phone = strings.TrimSpace(req.Phone)
	req.RealName = strings.TrimSpace(req.RealName)
	req.IdCardNo = strings.ToUpper(strings.TrimSpace(req.IdCardNo))
	req.DriverLicenseNo = strings.TrimSpace(req.DriverLicenseNo)
	req.AvatarURL = strings.TrimSpace(req.AvatarURL)
}

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
