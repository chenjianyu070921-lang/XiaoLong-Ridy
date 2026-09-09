package logic

import (
	"context"
	"strings"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"

	"golang.org/x/crypto/bcrypt"
)

func TestRegisterDriverRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name string
		req  *types.RegisterDriverRequest
	}{
		{name: "nil request", req: nil},
		{name: "invalid phone", req: &types.RegisterDriverRequest{Phone: "123", Password: "Driver@123", RealName: "张三", IdCardNo: "110101199001011234", DriverLicenseNo: "DL10000001"}},
		{name: "short password", req: &types.RegisterDriverRequest{Phone: "13800000001", Password: "1234567", RealName: "张三", IdCardNo: "110101199001011234", DriverLicenseNo: "DL10000001"}},
		{name: "empty real name", req: &types.RegisterDriverRequest{Phone: "13800000001", Password: "Driver@123", RealName: "   ", IdCardNo: "110101199001011234", DriverLicenseNo: "DL10000001"}},
		{name: "invalid id card", req: &types.RegisterDriverRequest{Phone: "13800000001", Password: "Driver@123", RealName: "张三", IdCardNo: "11010119900101123Y", DriverLicenseNo: "DL10000001"}},
		{name: "empty driver license", req: &types.RegisterDriverRequest{Phone: "13800000001", Password: "Driver@123", RealName: "张三", IdCardNo: "110101199001011234", DriverLicenseNo: "   "}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logic := NewDriverLogic(context.Background(), &svc.ServiceContext{DriverClient: &fakeDriverClient{}})
			if _, err := logic.RegisterDriver(tc.req); err == nil {
				t.Fatal("RegisterDriver() accepted invalid input")
			}
		})
	}
}

func TestRegisterDriverUsesNormalizedFieldsAndHashesPassword(t *testing.T) {
	client := &fakeDriverClient{}
	logic := NewDriverLogic(context.Background(), &svc.ServiceContext{DriverClient: client})

	resp, err := logic.RegisterDriver(&types.RegisterDriverRequest{
		Phone:           " 13800000001 ",
		Password:        "Driver@123",
		RealName:        " 张三 ",
		IdCardNo:        " 11010119900101123x ",
		DriverLicenseNo: " DL10000001 ",
		AvatarURL:       " avatar ",
	})
	if err != nil {
		t.Fatalf("RegisterDriver() error = %v", err)
	}
	if resp == nil || resp.ID != 25 || resp.Status != "DRIVER_STATUS_PENDING" {
		t.Fatalf("RegisterDriver() response = %+v", resp)
	}
	req := client.registerDriverRequest
	if req == nil {
		t.Fatal("RegisterDriver() did not call downstream client")
	}
	if req.GetPhone() != "13800000001" || req.GetRealName() != "张三" || req.GetIdCardNo() != "11010119900101123X" || req.GetDriverLicenseNo() != "DL10000001" || req.GetAvatarUrl() != "avatar" {
		t.Fatalf("RegisterDriver() downstream request = %+v", req)
	}
	if req.GetPasswordHash() == "" || req.GetPasswordHash() == "Driver@123" {
		t.Fatalf("RegisterDriver() password hash not populated: %q", req.GetPasswordHash())
	}
	if err := bcrypt.CompareHashAndPassword([]byte(req.GetPasswordHash()), []byte("Driver@123")); err != nil {
		t.Fatalf("RegisterDriver() password hash mismatch: %v", err)
	}
	if strings.TrimSpace(req.GetPasswordHash()) == "" {
		t.Fatal("RegisterDriver() password hash is blank")
	}
}
