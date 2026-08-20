package logic

import (
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var driverPhoneRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)
var driverIDCardRegexp = regexp.MustCompile(`^\d{17}[\dXx]$`)

// validateDriverIdentity 校验司机账号的手机号、姓名、身份证号和驾驶证号。
func validateDriverIdentity(phone, realName, idCardNo, driverLicenseNo string) error {
	if !driverPhoneRegexp.MatchString(phone) {
		return errors.New("手机号格式不合法")
	}
	if strings.TrimSpace(realName) == "" {
		return errors.New("真实姓名不能为空")
	}
	if !driverIDCardRegexp.MatchString(idCardNo) {
		return errors.New("身份证号格式不合法")
	}
	if strings.TrimSpace(driverLicenseNo) == "" {
		return errors.New("驾驶证号不能为空")
	}
	return nil
}

// validateDriverPasswordHash 校验 RPC 入参必须是可被 bcrypt 识别的密码哈希。
func validateDriverPasswordHash(passwordHash string) error {
	if strings.TrimSpace(passwordHash) == "" {
		return errors.New("密码哈希不能为空")
	}
	if _, err := bcrypt.Cost([]byte(passwordHash)); err != nil {
		return errors.New("密码哈希格式不合法")
	}
	return nil
}

// validateDriverPassword 校验登录明文密码长度，bcrypt 只支持最多 72 字节。
func validateDriverPassword(password string) error {
	passwordLength := len([]byte(password))
	if passwordLength < 8 || passwordLength > 72 {
		return errors.New("密码长度必须为8到72字节")
	}
	return nil
}
