package logic

import (
	"testing"

	"XiaoLong-Ridy/common/cryptox"

	"golang.org/x/crypto/bcrypt"
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

func TestPrepareDriverPasswordHash(t *testing.T) {
	plaintext := "Driver@123"
	hash, err := prepareDriverPasswordHash(plaintext)
	if err != nil {
		t.Fatalf("prepareDriverPasswordHash() error = %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)); err != nil {
		t.Fatalf("prepared hash does not match plaintext: %v", err)
	}

	existingHash, err := cryptox.BcryptHash(plaintext)
	if err != nil {
		t.Fatalf("BcryptHash() error = %v", err)
	}
	hash, err = prepareDriverPasswordHash(existingHash)
	if err != nil {
		t.Fatalf("prepareDriverPasswordHash() error = %v", err)
	}
	if hash != existingHash {
		t.Fatalf("prepared hash = %q, want existing bcrypt hash", hash)
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
