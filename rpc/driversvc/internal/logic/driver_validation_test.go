package logic

import (
	"strings"
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
	if err := validateDriverIdentity("13800000000", "张三", "110101199001011237", "DL10000001"); err != nil {
		t.Fatalf("validateDriverIdentity() error = %v", err)
	}
	if err := validateDriverIdentity("852656554556", "张三", "110101199001011237", "DL10000001"); err != nil {
		t.Fatalf("validateDriverIdentity() accepted a driver-phone format used by the driver app: %v", err)
	}
	if err := validateDriverIdentity("12800000000", "张三", "110101199001011237", "DL10000001"); err == nil {
		t.Fatal("validateDriverIdentity() accepted an invalid phone")
	}
}

func TestValidateCreateVehicle(t *testing.T) {
	if err := validateCreateVehicle("粤B12345", "BYD", "Han", "black", 1, "INS-1"); err != nil {
		t.Fatalf("validateCreateVehicle() error = %v", err)
	}

	cases := []struct {
		name        string
		plateNo     string
		brand       string
		model       string
		color       string
		vehicleType int32
		insuranceNo string
	}{
		{name: "bad plate", plateNo: "ABC123", brand: "BYD", model: "Han", vehicleType: 1},
		{name: "bad type", plateNo: "粤B12345", brand: "BYD", model: "Han", vehicleType: 0},
		{name: "missing brand", plateNo: "粤B12345", model: "Han", vehicleType: 1},
		{name: "missing model", plateNo: "粤B12345", brand: "BYD", vehicleType: 1},
		{name: "brand too long", plateNo: "粤B12345", brand: strings.Repeat("B", maxVehicleBrandLen+1), model: "Han", vehicleType: 1},
		{name: "model too long", plateNo: "粤B12345", brand: "BYD", model: strings.Repeat("M", maxVehicleModelLen+1), vehicleType: 1},
		{name: "color too long", plateNo: "粤B12345", brand: "BYD", model: "Han", color: strings.Repeat("C", maxVehicleColorLen+1), vehicleType: 1},
		{name: "insurance too long", plateNo: "粤B12345", brand: "BYD", model: "Han", vehicleType: 1, insuranceNo: strings.Repeat("I", maxVehicleInsuranceLen+1)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateCreateVehicle(tc.plateNo, tc.brand, tc.model, tc.color, tc.vehicleType, tc.insuranceNo); err == nil {
				t.Fatal("validateCreateVehicle() accepted invalid input")
			}
		})
	}
}
