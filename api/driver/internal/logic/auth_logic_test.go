package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLoginByPasswordDelegatesToDriverService(t *testing.T) {
	driverClient := &fakeDriverClient{
		loginResponse: &driversproto.LoginResponse{
			Token:    "token",
			ExpireIn: 7200,
			Driver: &driversproto.Driver{
				Id:        25,
				Phone:     "13800000001",
				Status:    driversproto.DriverStatus_DRIVER_STATUS_NORMAL,
				VehicleId: 77,
			},
		},
	}
	logic := NewAuthLogic(context.Background(), &svc.ServiceContext{
		DriverClient: driverClient,
		CodeCache:    svc.NewLocalCodeCache(time.Minute),
	})

	resp, err := logic.LoginByPassword(&types.LoginByPasswordRequest{
		Phone:    "13800000001",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("LoginByPassword() error = %v", err)
	}
	if resp.Token != "token" || resp.Driver.ID != 25 || resp.Driver.Phone != "138****0001" || resp.Driver.VehicleID != 77 {
		t.Fatalf("LoginByPassword() response = %+v", resp)
	}
	req := driverClient.loginRequest
	if req == nil || req.GetPhone() != "13800000001" || req.GetPassword() != "password123" {
		t.Fatalf("driversvc Login request = %+v", req)
	}
}

func TestLoginBySMSVerifiesCodeThenDelegatesToDriverService(t *testing.T) {
	driverClient := &fakeDriverClient{}
	codeCache := svc.NewLocalCodeCache(time.Minute)
	codeCache.Set("13800000001", "123456")
	logic := NewAuthLogic(context.Background(), &svc.ServiceContext{
		DriverClient: driverClient,
		CodeCache:    codeCache,
	})

	resp, err := logic.LoginBySMS(&types.LoginBySMSRequest{
		Phone: "13800000001",
		Code:  "123456",
	})
	if err != nil {
		t.Fatalf("LoginBySMS() error = %v", err)
	}
	if resp.Token != "sms-token" || resp.Driver.ID != 25 {
		t.Fatalf("LoginBySMS() response = %+v", resp)
	}
	req := driverClient.loginBySMSRequest
	if req == nil || req.GetPhone() != "13800000001" {
		t.Fatalf("driversvc LoginBySMS request = %+v", req)
	}
}

func TestLoginBySMSRejectsInvalidCodeBeforeDriverService(t *testing.T) {
	driverClient := &fakeDriverClient{}
	logic := NewAuthLogic(context.Background(), &svc.ServiceContext{
		DriverClient: driverClient,
		CodeCache:    svc.NewLocalCodeCache(time.Minute),
	})

	if _, err := logic.LoginBySMS(&types.LoginBySMSRequest{Phone: "13800000001", Code: "000000"}); err != ErrCodeInvalid {
		t.Fatalf("LoginBySMS() error = %v, want %v", err, ErrCodeInvalid)
	}
	if driverClient.loginBySMSRequest != nil {
		t.Fatalf("driversvc LoginBySMS should not be called, got %+v", driverClient.loginBySMSRequest)
	}
}

func TestSendSMSCodeDoesNotReturnPlaintextCode(t *testing.T) {
	driverClient := &fakeDriverClient{}
	codeCache := svc.NewLocalCodeCache(time.Minute)
	logic := NewAuthLogic(context.Background(), &svc.ServiceContext{
		DriverClient: driverClient,
		CodeCache:    codeCache,
	})

	resp, err := logic.SendSMSCode(&types.SendSMSCodeRequest{Phone: "13800000001"})
	if err != nil {
		t.Fatalf("SendSMSCode() error = %v", err)
	}
	if !resp.Success || resp.ExpireIn != 60 {
		t.Fatalf("SendSMSCode() response = %+v", resp)
	}
}

func TestSendSMSCodeRejectsNilRequest(t *testing.T) {
	logic := NewAuthLogic(context.Background(), &svc.ServiceContext{
		CodeCache: svc.NewLocalCodeCache(time.Minute),
	})

	if _, err := logic.SendSMSCode(nil); err != ErrInvalidParam {
		t.Fatalf("SendSMSCode(nil) error = %v, want %v", err, ErrInvalidParam)
	}
}

func TestSendSMSCodeRequiresCodeCache(t *testing.T) {
	logic := NewAuthLogic(context.Background(), &svc.ServiceContext{})

	if _, err := logic.SendSMSCode(&types.SendSMSCodeRequest{Phone: "13800000001"}); err != ErrCodeSendFailed {
		t.Fatalf("SendSMSCode() error = %v, want %v", err, ErrCodeSendFailed)
	}
}

func TestLoginBySMSRequiresCodeCache(t *testing.T) {
	logic := NewAuthLogic(context.Background(), &svc.ServiceContext{})

	if _, err := logic.LoginBySMS(&types.LoginBySMSRequest{Phone: "13800000001", Code: "123456"}); err != ErrCodeInvalid {
		t.Fatalf("LoginBySMS() error = %v, want %v", err, ErrCodeInvalid)
	}
}

func TestNormalizeLoginErrorTreatsPendingDriverAsForbidden(t *testing.T) {
	err := normalizeLoginError(status.Error(codes.Unknown, "账号未审核通过"))
	if err != ErrDriverFrozen {
		t.Fatalf("normalizeLoginError() = %v, want %v", err, ErrDriverFrozen)
	}
}

func TestNormalizeLoginErrorPreservesDownstreamTransportStatus(t *testing.T) {
	cases := []struct {
		name string
		code codes.Code
	}{
		{name: "unavailable", code: codes.Unavailable},
		{name: "deadline exceeded", code: codes.DeadlineExceeded},
		{name: "unimplemented", code: codes.Unimplemented},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			downstream := status.Error(testCase.code, "downstream failed")
			if got := normalizeLoginError(downstream); got != downstream {
				t.Fatalf("normalizeLoginError() = %v, want %v", got, downstream)
			}
		})
	}
}
