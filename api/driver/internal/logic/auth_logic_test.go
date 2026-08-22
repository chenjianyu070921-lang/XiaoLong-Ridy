package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

func TestLoginByPasswordDelegatesToDriverService(t *testing.T) {
	driverClient := &fakeDriverClient{}
	logic := NewAuthLogic(context.Background(), &svc.ServiceContext{
		DriverClient: driverClient,
		CodeCache:    svc.NewCodeCache(time.Minute),
	})

	resp, err := logic.LoginByPassword(&types.LoginByPasswordRequest{
		Phone:    "13800000001",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("LoginByPassword() error = %v", err)
	}
	if resp.Token != "token" || resp.Driver.ID != 25 || resp.Driver.Phone != "138****0001" {
		t.Fatalf("LoginByPassword() response = %+v", resp)
	}
	req := driverClient.loginRequest
	if req == nil || req.GetPhone() != "13800000001" || req.GetPassword() != "password123" {
		t.Fatalf("driversvc Login request = %+v", req)
	}
}

func TestLoginBySMSVerifiesCodeThenDelegatesToDriverService(t *testing.T) {
	driverClient := &fakeDriverClient{}
	codeCache := svc.NewCodeCache(time.Minute)
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
		CodeCache:    svc.NewCodeCache(time.Minute),
	})

	if _, err := logic.LoginBySMS(&types.LoginBySMSRequest{Phone: "13800000001", Code: "000000"}); err != ErrCodeInvalid {
		t.Fatalf("LoginBySMS() error = %v, want %v", err, ErrCodeInvalid)
	}
	if driverClient.loginBySMSRequest != nil {
		t.Fatalf("driversvc LoginBySMS should not be called, got %+v", driverClient.loginBySMSRequest)
	}
}
