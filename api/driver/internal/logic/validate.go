package logic

import (
	"errors"
	"regexp"
	"strings"

	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

var (
	ErrDriverClientNotConfigured      = errors.New("driver client not configured")
	ErrOrderClientNotConfigured       = errors.New("order client not configured")
	ErrDispatchClientNotConfigured    = errors.New("dispatch client not configured")
	ErrReviewStorageNotConfigured     = errors.New("passenger review storage not configured")
	ErrTrajectoryStorageNotConfigured = errors.New("trip trajectory storage not configured")
	ErrInvalidParam                   = errors.New("invalid param")
	ErrForbiddenDriverResource        = errors.New("forbidden driver resource")
)

const (
	maxVehicleBrandLen     = 64
	maxVehicleModelLen     = 64
	maxVehicleColorLen     = 32
	maxVehicleInsuranceLen = 64
)

var phoneRegexp = regexp.MustCompile(`^(?:1[3-9]\d{9}|\d{12,15})$`)
var vehiclePlateRegexp = regexp.MustCompile("^[京津沪渝冀云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼][A-Z][A-HJ-NP-Z0-9]{4,5}[A-HJ-NP-Z0-9挂学警港澳]$")

func validPhone(phone string) bool {
	return phoneRegexp.MatchString(phone)
}

// validLocation 校验经纬度是否落在合法范围内。
func validLocation(longitude, latitude float64) bool {
	return longitude >= -180 && longitude <= 180 && latitude >= -90 && latitude <= 90
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

func enumVehicleStatus(status *string) *driversproto.VehicleStatus {
	if status == nil || *status == "" {
		return nil
	}
	value := strings.ToUpper(strings.TrimSpace(*status))
	var mapped driversproto.VehicleStatus
	switch value {
	case "VEHICLE_STATUS_PENDING", "VEHICLE_STATUS_NORMAL", "VEHICLE_STATUS_DISABLED":
		switch value {
		case "VEHICLE_STATUS_PENDING":
			mapped = driversproto.VehicleStatus_VEHICLE_STATUS_PENDING
		case "VEHICLE_STATUS_NORMAL":
			mapped = driversproto.VehicleStatus_VEHICLE_STATUS_NORMAL
		case "VEHICLE_STATUS_DISABLED":
			mapped = driversproto.VehicleStatus_VEHICLE_STATUS_DISABLED
		}
		return &mapped
	default:
		return nil
	}
}

func validVehiclePlate(plateNo string) bool {
	return vehiclePlateRegexp.MatchString(strings.ToUpper(strings.TrimSpace(plateNo)))
}

func validVehicleType(vehicleType int32) bool {
	return vehicleType >= 1 && vehicleType <= 5
}

func validCreateVehicle(plateNo, brand, model, color string, vehicleType int32, insuranceNo string) bool {
	if !validVehiclePlate(plateNo) || !validVehicleType(vehicleType) {
		return false
	}
	return validRequiredLength(brand, maxVehicleBrandLen) &&
		validRequiredLength(model, maxVehicleModelLen) &&
		validOptionalLength(color, maxVehicleColorLen) &&
		validOptionalLength(insuranceNo, maxVehicleInsuranceLen)
}

func validRequiredLength(value string, max int) bool {
	value = strings.TrimSpace(value)
	return value != "" && len([]rune(value)) <= max
}

func validOptionalLength(value string, max int) bool {
	return len([]rune(strings.TrimSpace(value))) <= max
}

func validNearbyDriversQuery(longitude, latitude, radiusMeters float64, limit int32) bool {
	return longitude >= -180 && longitude <= 180 &&
		latitude >= -90 && latitude <= 90 &&
		radiusMeters > 0 &&
		limit > 0
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
