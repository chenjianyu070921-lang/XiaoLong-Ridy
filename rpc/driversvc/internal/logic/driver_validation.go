package logic

import (
	"errors"
	"regexp"
	"strings"

	"XiaoLong-Ridy/common/cryptox"

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
var vehiclePlateRegexp = regexp.MustCompile("^[\u4eac\u6d25\u6caa\u6e1d\u5180\u4e91\u8fbd\u9ed1\u6e58\u7696\u9c81\u65b0\u82cf\u6d59\u8d63\u9102\u6842\u7518\u664b\u8499\u9655\u5409\u95fd\u8d35\u7ca4\u9752\u85cf\u5ddd\u5b81\u743c\u4f7f\u9886][A-Z][A-HJ-NP-Z0-9]{4,5}[A-HJ-NP-Z0-9\u4f7f\u9886\u6302\u5b66\u8b66\u6e2f\u6fb3]$")

func validateDriverIdentity(phone, realName, idCardNo, driverLicenseNo string) error {
	if !driverPhoneRegexp.MatchString(phone) {
		return errors.New("invalid driver phone")
	}
	if strings.TrimSpace(realName) == "" {
		return errors.New("driver real name is required")
	}
	if !driverIDCardRegexp.MatchString(idCardNo) {
		return errors.New("invalid driver id card")
	}
	if strings.TrimSpace(driverLicenseNo) == "" {
		return errors.New("driver license no is required")
	}
	return nil
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
	if amount <= 0 {
		return errors.New("withdraw amount must be greater than 0")
	}
	if amount > 100000 {
		return errors.New("withdraw amount exceeds limit")
	}
	return nil
}
