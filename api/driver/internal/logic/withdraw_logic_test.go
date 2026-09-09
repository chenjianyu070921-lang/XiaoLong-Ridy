package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

func TestCreateWithdrawRejectsMoreThanTwoDecimals(t *testing.T) {
	client := &fakeDriverClient{}
	logic := NewWithdrawLogic(context.Background(), &svc.ServiceContext{DriverClient: client})

	_, err := logic.CreateWithdraw(25, &types.CreateWithdrawRequest{
		Amount:     12.345,
		PayeeName:  "张三",
		PayAccount: "acct-1",
	})
	if err == nil {
		t.Fatal("CreateWithdraw() accepted amount with more than two decimals")
	}
	if client.createWithdrawRequest != nil {
		t.Fatalf("CreateWithdraw() forwarded request on invalid input: %+v", client.createWithdrawRequest)
	}
}

func TestCreateWithdrawNormalizesAmountBeforeForwarding(t *testing.T) {
	client := &fakeDriverClient{
		createWithdrawResponse: &driversproto.CreateWithdrawResponse{
			Id:         88,
			WithdrawNo: "WD1001",
			Status:     1,
			CreatedAt:  123,
		},
	}
	logic := NewWithdrawLogic(context.Background(), &svc.ServiceContext{DriverClient: client})

	amount := 12.340000000000002
	resp, err := logic.CreateWithdraw(25, &types.CreateWithdrawRequest{
		Amount:     amount,
		PayeeName:  " 张三 ",
		PayAccount: " acct-1 ",
	})
	if err != nil {
		t.Fatalf("CreateWithdraw() error = %v", err)
	}
	if resp.ID != 88 || resp.WithdrawNo != "WD1001" || resp.Status != 1 || resp.CreatedAt != 123 {
		t.Fatalf("CreateWithdraw() response = %+v", resp)
	}
	if client.createWithdrawRequest == nil {
		t.Fatal("CreateWithdraw() did not forward request")
	}
	if client.createWithdrawRequest.GetAmount() != 12.34 {
		t.Fatalf("CreateWithdraw() forwarded amount = %.17g, want 12.34", client.createWithdrawRequest.GetAmount())
	}
	if client.createWithdrawRequest.GetAmount() == amount {
		t.Fatalf("CreateWithdraw() forwarded unnormalized amount = %.17g", client.createWithdrawRequest.GetAmount())
	}
	if client.createWithdrawRequest.GetPayeeName() != "张三" || client.createWithdrawRequest.GetPayAccount() != "acct-1" {
		t.Fatalf("CreateWithdraw() forwarded trimmed fields = %+v", client.createWithdrawRequest)
	}
}
