package logic

import (
	"testing"

	"XiaoLong-Ridy/common/cryptox"
)

func TestValidateDriverPasswordHash(t *testing.T) {
	passwordHash, err := cryptox.BcryptHash("Driver@123")
	if err != nil {
		t.Fatalf("BcryptHash() error = %v", err)
	}
	if err := validateDriverPasswordHash(passwordHash); err != nil {
		t.Fatalf("validateDriverPasswordHash() error = %v", err)
	}
	if err := validateDriverPasswordHash("e10adc3949ba59abbe56e057f20f883e"); err == nil {
		t.Fatal("validateDriverPasswordHash() accepted an MD5 digest")
	}
}

func TestValidateDriverIdentity(t *testing.T) {
	if err := validateDriverIdentity("13800000000", "张三", "110101199001011234", "DL10000001"); err != nil {
		t.Fatalf("validateDriverIdentity() error = %v", err)
	}
	if err := validateDriverIdentity("12800000000", "张三", "110101199001011234", "DL10000001"); err == nil {
		t.Fatal("validateDriverIdentity() accepted an invalid phone")
	}
}
