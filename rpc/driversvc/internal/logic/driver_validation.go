package logic

import (
	"errors"
	"math"
	"regexp"
	"strings"

	"XiaoLong-Ridy/common/cryptox"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/crypto/bcrypt"
)

const (
	maxVehicleBrandLen     = 64
	maxVehicleModelLen     = 64
	maxVehicleColorLen     = 32
	maxVehicleInsuranceLen = 64
)

var driverPhoneRegexp = regexp.MustCompile(`^(?:1[3-9]\d{9}|\d{12,15})$`)
var driverIDCardRegexp = regexp.MustCompile(`^\d{17}[\dXx]$`)
// idCardWeights 是 GB 11643 身份证校验位权重表，须与前端 useDriverAssets.js 的 isValidIDCard 保持一致。
var idCardWeights = []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
var vehiclePlateRegexp = regexp.MustCompile("^[京津沪渝冀云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼使领][A-Z][A-HJ-NP-Z0-9]{4,5}[A-HJ-NP-Z0-9使领挂学警港澳]?$")

func validateDriverIdentity(phone, realName, idCardNo, driverLicenseNo string) error {
	if !driverPhoneRegexp.MatchString(phone) {
		return status.Error(codes.InvalidArgument, "手机号格式不合法")
	}
	if strings.TrimSpace(realName) == "" {
		return status.Error(codes.InvalidArgument, "真实姓名不能为空")
	}
	if !isValidIDCardNo(idCardNo) {
		return status.Error(codes.InvalidArgument, "身份证号格式不合法")
	}
	if strings.TrimSpace(driverLicenseNo) == "" {
		return status.Error(codes.InvalidArgument, "驾驶证号不能为空")
	}
	return nil
}

// isValidIDCardNo 校验 18 位身份证号：正则格式 + GB 11643 校验位。
// 算法须与前端 useDriverAssets.js 的 isValidIDCard 保持一致。
func isValidIDCardNo(idCardNo string) bool {
	if !driverIDCardRegexp.MatchString(idCardNo) {
		return false
	}
	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(idCardNo[i]-'0') * idCardWeights[i]
	}
	const checkCodes = "10X98765432"
	expected := checkCodes[sum%11]
	last := idCardNo[17]
	if expected == 'X' {
		return last == 'X' || last == 'x'
	}
	return expected == last
}

func validateDriverPasswordHash(passwordHash string) error {
	if strings.TrimSpace(passwordHash) == "" {
		return errors.New("driver password hash is required")
	}
	if _, err := bcrypt.Cost([]byte(passwordHash)); err != nil {
		return errors.New("invalid driver password hash")
	}
	return nil
}

func prepareDriverPasswordHash(password string) (string, error) {
	if err := validateDriverPasswordHash(password); err == nil {
		return password, nil
	}
	if err := validateDriverPassword(password); err != nil {
		return "", err
	}
	return cryptox.BcryptHash(password)
}

func validateDriverPassword(password string) error {
	passwordLength := len([]byte(password))
	if passwordLength < 8 || passwordLength > 72 {
		return errors.New("driver password length must be 8-72 bytes")
	}
	return nil
}

func validateVehiclePlate(plateNo string) error {
	if !vehiclePlateRegexp.MatchString(strings.ToUpper(strings.TrimSpace(plateNo))) {
		return errors.New("invalid vehicle plate no")
	}
	return nil
}

func validateVehicleType(vehicleType int32) error {
	if vehicleType < 1 || vehicleType > 5 {
		return errors.New("invalid vehicle type")
	}
	return nil
}

func validateCreateVehicle(plateNo, brand, model, color string, vehicleType int32, insuranceNo string) error {
	if strings.TrimSpace(plateNo) == "" {
		return errors.New("vehicle plate no is required")
	}
	if err := validateVehiclePlate(plateNo); err != nil {
		return err
	}
	if err := validateVehicleType(vehicleType); err != nil {
		return err
	}
	if err := validateRequiredLength("vehicle brand", brand, maxVehicleBrandLen); err != nil {
		return err
	}
	if err := validateRequiredLength("vehicle model", model, maxVehicleModelLen); err != nil {
		return err
	}
	if err := validateOptionalLength("vehicle color", color, maxVehicleColorLen); err != nil {
		return err
	}
	if err := validateOptionalLength("vehicle insurance no", insuranceNo, maxVehicleInsuranceLen); err != nil {
		return err
	}
	return nil
}

func validateRequiredLength(name, value string, max int) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New(name + " is required")
	}
	if len([]rune(value)) > max {
		return errors.New(name + " is too long")
	}
	return nil
}

func validateOptionalLength(name, value string, max int) error {
	if len([]rune(strings.TrimSpace(value))) > max {
		return errors.New(name + " is too long")
	}
	return nil
}

func validateWithdrawAmount(amount float64) error {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return errors.New("withdraw amount must be a finite number")
	}
	if amount <= 0 {
		return errors.New("withdraw amount must be greater than 0")
	}
	if amount > 100000 {
		return errors.New("withdraw amount exceeds limit")
	}
	cents := amount * 100
	if math.Abs(cents-math.Round(cents)) > 1e-6 {
		return errors.New("withdraw amount must not exceed 2 decimal places")
	}
	return nil
}

func normalizeWithdrawAmount(amount float64) float64 {
	return math.Round(amount*100) / 100
}
