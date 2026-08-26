package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"
)

func TestCreateWithdrawCreatesPendingRecord(t *testing.T) {
	repo := &fakeWithdrawRepository{}
	logic := NewCreateWithdrawLogic(context.Background(), &svc.ServiceContext{DriverWithdrawRepository: repo})

	resp, err := logic.CreateWithdraw(&proto.CreateWithdrawRequest{
		DriverId:   25,
		Amount:     128.5,
		PayeeName:  " 张三 ",
		PayAccount: " acct-1 ",
	})
	if err != nil {
		t.Fatalf("CreateWithdraw() error = %v", err)
	}
	if resp.GetId() != 88 || resp.GetStatus() != int32(withdrawStatusPending) || resp.GetWithdrawNo() == "" {
		t.Fatalf("CreateWithdraw() response = %+v", resp)
	}
	if repo.created == nil || repo.created.DriverId != 25 || repo.created.Amount != 128.5 ||
		repo.created.PayeeName != "张三" || repo.created.PayAccount != "acct-1" ||
		repo.created.Status != withdrawStatusPending || repo.created.AppliedAt == nil {
		t.Fatalf("created withdraw = %+v", repo.created)
	}
}

func TestCreateWithdrawRejectsInvalidInputBeforeCreate(t *testing.T) {
	repo := &fakeWithdrawRepository{}
	logic := NewCreateWithdrawLogic(context.Background(), &svc.ServiceContext{DriverWithdrawRepository: repo})

	if _, err := logic.CreateWithdraw(&proto.CreateWithdrawRequest{DriverId: 25, Amount: 0, PayeeName: "张三", PayAccount: "acct"}); err == nil {
		t.Fatal("CreateWithdraw() accepted invalid amount")
	}
	if repo.created != nil {
		t.Fatalf("invalid withdraw should not be created: %+v", repo.created)
	}
}

func TestListWithdrawsReturnsRepositoryRecords(t *testing.T) {
	appliedAt := time.Unix(100, 0)
	paidAt := time.Unix(200, 0)
	repo := &fakeWithdrawRepository{
		records: []*model.DriverWithdraw{
			{
				Id:         88,
				DriverId:   25,
				WithdrawNo: "WD1001",
				Amount:     128.5,
				PayeeName:  "张三",
				PayAccount: "acct-1",
				Status:     withdrawStatusPending,
				Remark:     "pending",
				AppliedAt:  &appliedAt,
				PaidAt:     &paidAt,
				CreatedAt:  time.Unix(90, 0),
			},
		},
		total: 1,
	}
	logic := NewListWithdrawsLogic(context.Background(), &svc.ServiceContext{DriverWithdrawRepository: repo})

	resp, err := logic.ListWithdraws(&proto.ListWithdrawsRequest{DriverId: 25, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListWithdraws() error = %v", err)
	}
	if repo.driverID != 25 || repo.page != 1 || repo.pageSize != 20 {
		t.Fatalf("repository args = driverID:%d page:%d pageSize:%d", repo.driverID, repo.page, repo.pageSize)
	}
	if resp.GetTotal() != 1 || len(resp.GetRecords()) != 1 {
		t.Fatalf("ListWithdraws() response = %+v", resp)
	}
	record := resp.GetRecords()[0]
	if record.GetId() != 88 || record.GetWithdrawNo() != "WD1001" || record.GetAppliedAt() != 100 ||
		record.GetPaidAt() != 200 || record.GetCreatedAt() != 90 {
		t.Fatalf("withdraw record = %+v", record)
	}
}

type fakeWithdrawRepository struct {
	created  *model.DriverWithdraw
	records  []*model.DriverWithdraw
	total    int64
	driverID uint64
	page     int32
	pageSize int32
}

func (f *fakeWithdrawRepository) Create(_ context.Context, withdraw *model.DriverWithdraw) error {
	withdraw.Id = 88
	withdraw.CreatedAt = time.Unix(90, 0)
	f.created = withdraw
	return nil
}

func (f *fakeWithdrawRepository) ListByDriver(_ context.Context, driverID uint64, page, pageSize int32) ([]*model.DriverWithdraw, int64, error) {
	f.driverID = driverID
	f.page = page
	f.pageSize = pageSize
	return f.records, f.total, nil
}
