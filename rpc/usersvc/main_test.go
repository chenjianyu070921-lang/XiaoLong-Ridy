package main

import (
	"context"
	"testing"

	"XiaoLong-Ridy/rpc/usersvc/client"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

func TestLocalUserServiceCanLoginBySMS(t *testing.T) {
	var code string
	userClient := client.NewLocalClient("test-signing-key", func(_ string, sentCode string) {
		code = sentCode
	})

	if _, err := userClient.SendSMSCode(context.Background(), &userproto.SendSMSCodeRequest{
		Phone: "13800138000",
	}); err != nil {
		t.Fatalf("SendSMSCode() error = %v", err)
	}
	resp, err := userClient.LoginBySMS(context.Background(), &userproto.LoginBySMSRequest{
		Phone: "13800138000",
		Code:  code,
	})
	if err != nil {
		t.Fatalf("LoginBySMS() error = %v", err)
	}
	if !resp.GetIsNewUser() || resp.GetToken() == "" || resp.GetUser().GetUserId() == 0 {
		t.Fatalf("LoginBySMS() response = %+v", resp)
	}
}
