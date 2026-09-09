package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/common/cryptox"
	"XiaoLong-Ridy/rpc/driversvc/internal/config"
	"XiaoLong-Ridy/rpc/driversvc/internal/crypto"
	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateWithdrawCreatesPendingRecord(t *testing.T) {
	repo := &fakeWithdrawRepository{}
	cardRepo := &fakeBankCardRepository{
		card: &model.DriverBankCard{
			Id:               7,
			DriverId:         25,
			BankName:         "中国工商银行",
			CardNo:           mustEncryptCardNo(t, "6222021234567890123"),
			HolderName:       "张三",
			HolderIdCard:     "110101199001011234",
			ReservedPhone:    "13800000001",
			WithdrawPasswordHash: mustHash(t, "123456"),
			Status:           1,
		},
	}
	driverRepo := &fakeDriverRepository{
		driver: &model.Driver{Id: 25, Phone: "13800000001", RealName: "张三", WithdrawPasswordHash: mustHash(t, "123456")},
	}
	logic := NewCreateWithdrawLogic(context.Background(), &svc.ServiceContext{
		DriverWithdrawRepository:   repo,
		DriverBankCardRepository:   cardRepo,
		DriverRepository:           driverRepo,
		Config:                     config.Config{CardKey: "test-card-key"},
	})

	resp, err := logic.CreateWithdraw(&proto.CreateWithdrawRequest{
		DriverId:        25,
		Amount:          128.5,
		BankCardId:      7,
		WithdrawPassword: "123456",
	})
	if err != nil {
		t.Fatalf("CreateWithdraw() error = %v", err)
	}
	if resp.GetId() != 88 || resp.GetStatus() != int32(withdrawStatusPending) || resp.GetWithdrawNo() == "" {
		t.Fatalf("CreateWithdraw() response = %+v", resp)
	}
	if repo.created == nil || repo.created.DriverId != 25 || repo.created.Amount != 128.5 ||
		repo.created.PayeeName != "张三" || repo.created.PayAccount != "6222021234567890123" ||
		repo.created.Status != withdrawStatusPending || repo.created.AppliedAt == nil {
		t.Fatalf("created withdraw = %+v", repo.created)
	}
}

func TestCreateWithdrawRejectsInvalidInputBeforeCreate(t *testing.T) {
	repo := &fakeWithdrawRepository{}
	cardRepo := &fakeBankCardRepository{
		card: &model.DriverBankCard{
			Id:            7,
			DriverId:      25,
			BankName:      "中国工商银行",
			CardNo:        mustEncryptCardNo(t, "6222021234567890123"),
			HolderName:    "张三",
			Status:        1,
		},
	}
	driverRepo := &fakeDriverRepository{
		driver: &model.Driver{Id: 25, Phone: "13800000001", RealName: "张三", WithdrawPasswordHash: mustHash(t, "123456")},
	}
	logic := NewCreateWithdrawLogic(context.Background(), &svc.ServiceContext{
		DriverWithdrawRepository:   repo,
		DriverBankCardRepository:   cardRepo,
		DriverRepository:           driverRepo,
		Config:                     config.Config{CardKey: "test-card-key"},
	})

	if _, err := logic.CreateWithdraw(&proto.CreateWithdrawRequest{DriverId: 25, Amount: 0}); err == nil {
		t.Fatal("CreateWithdraw() accepted invalid amount")
	}
	if _, err := logic.CreateWithdraw(&proto.CreateWithdrawRequest{DriverId: 25, Amount: 10}); err == nil {
		t.Fatal("CreateWithdraw() accepted missing bank card")
	}
	if _, err := logic.CreateWithdraw(&proto.CreateWithdrawRequest{DriverId: 25, Amount: 10, BankCardId: 7}); err == nil {
		t.Fatal("CreateWithdraw() accepted missing withdraw password")
	}
	if repo.created != nil {
		t.Fatalf("invalid withdraw should not be created: %+v", repo.created)
	}
}

func mustEncryptCardNo(t *testing.T, no string) string {
	t.Helper()
	enc, err := crypto.EncryptCardNo(no, "test-card-key")
	if err != nil {
		t.Fatalf("encrypt card no: %v", err)
	}
	return enc
}

func mustHash(t *testing.T, pwd string) string {
	t.Helper()
	h, err := cryptox.BcryptHash(pwd)
	if err != nil {
		t.Fatalf("bcrypt hash: %v", err)
	}
	return h
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

func TestAuditWithdrawApprovesPendingRecord(t *testing.T) {
	repo := &fakeWithdrawRepository{
		records: []*model.DriverWithdraw{{
			Id:       88,
			DriverId: 25,
			Status:   withdrawStatusPending,
		}},
	}
	logic := NewAuditWithdrawLogic(context.Background(), &svc.ServiceContext{DriverWithdrawRepository: repo})

	resp, err := logic.AuditWithdraw(&proto.AuditWithdrawRequest{
		WithdrawId: 88,
		Approve:    true,
		Remark:     "paid by finance",
		OperatorId: 9001,
		Ip:         "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("AuditWithdraw() error = %v", err)
	}
	if resp.GetDriverId() != 25 {
		t.Fatalf("AuditWithdraw() driver id = %d, want 25", resp.GetDriverId())
	}
	if repo.auditID != 88 || repo.auditStatus != withdrawStatusPaid || repo.auditRemark != "paid by finance" || repo.auditPaidAt == nil {
		t.Fatalf("audit call = id:%d status:%d remark:%q paidAt:%v", repo.auditID, repo.auditStatus, repo.auditRemark, repo.auditPaidAt)
	}
}

func TestAuditWithdrawRejectsPendingRecordWithRemark(t *testing.T) {
	repo := &fakeWithdrawRepository{
		records: []*model.DriverWithdraw{{
			Id:       89,
			DriverId: 26,
			Status:   withdrawStatusPending,
		}},
	}
	logic := NewAuditWithdrawLogic(context.Background(), &svc.ServiceContext{DriverWithdrawRepository: repo})

	resp, err := logic.AuditWithdraw(&proto.AuditWithdrawRequest{
		WithdrawId: 89,
		Approve:    false,
		Remark:     "bank account mismatch",
		OperatorId: 9001,
	})
	if err != nil {
		t.Fatalf("AuditWithdraw() error = %v", err)
	}
	if resp.GetDriverId() != 26 {
		t.Fatalf("AuditWithdraw() driver id = %d, want 26", resp.GetDriverId())
	}
	if repo.auditID != 89 || repo.auditStatus != withdrawStatusFailed || repo.auditRemark != "bank account mismatch" || repo.auditPaidAt != nil {
		t.Fatalf("audit call = id:%d status:%d remark:%q paidAt:%v", repo.auditID, repo.auditStatus, repo.auditRemark, repo.auditPaidAt)
	}
}

func TestAuditWithdrawAllowsOnlyPendingRecords(t *testing.T) {
	repo := &fakeWithdrawRepository{
		records: []*model.DriverWithdraw{{
			Id:       90,
			DriverId: 27,
			Status:   int8(withdrawStatusPaid),
		}},
	}
	logic := NewAuditWithdrawLogic(context.Background(), &svc.ServiceContext{DriverWithdrawRepository: repo})

	_, err := logic.AuditWithdraw(&proto.AuditWithdrawRequest{
		WithdrawId: 90,
		Approve:    false,
		Remark:     "late reject",
		OperatorId: 9001,
	})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("AuditWithdraw() error = %v, want FailedPrecondition", err)
	}
	if repo.auditID != 0 {
		t.Fatalf("non-pending withdraw should not be audited, audit id = %d", repo.auditID)
	}
}

func TestAuditWithdrawRejectRequiresRemark(t *testing.T) {
	repo := &fakeWithdrawRepository{}
	logic := NewAuditWithdrawLogic(context.Background(), &svc.ServiceContext{DriverWithdrawRepository: repo})

	_, err := logic.AuditWithdraw(&proto.AuditWithdrawRequest{
		WithdrawId: 88,
		Approve:    false,
		OperatorId: 9001,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("AuditWithdraw() error = %v, want InvalidArgument", err)
	}
	if repo.auditID != 0 {
		t.Fatalf("invalid reject should not be audited, audit id = %d", repo.auditID)
	}
}

type fakeWithdrawRepository struct {
	created     *model.DriverWithdraw
	records     []*model.DriverWithdraw
	total       int64
	driverID    uint64
	page        int32
	pageSize    int32
	auditID     uint64
	auditStatus int32
	auditRemark string
	auditPaidAt *time.Time
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

func (f *fakeWithdrawRepository) AdminList(_ context.Context, filter repository.AdminWithdrawFilter) ([]*model.DriverWithdraw, int64, error) {
	return f.records, f.total, nil
}

func (f *fakeWithdrawRepository) GetByID(_ context.Context, id uint64) (*model.DriverWithdraw, error) {
	for _, record := range f.records {
		if record.Id == id {
			return record, nil
		}
	}
	return nil, nil
}

func (f *fakeWithdrawRepository) Audit(_ context.Context, id uint64, status int32, remark string, paidAt *time.Time) error {
	f.auditID = id
	f.auditStatus = status
	f.auditRemark = remark
	f.auditPaidAt = paidAt
	for _, record := range f.records {
		if record.Id == id {
			record.Status = int8(status)
			record.Remark = remark
			record.PaidAt = paidAt
		}
	}
	return nil
}


// fakeBankCardRepository 嵌入接口，仅覆盖提现逻辑用到的 GetByID。
type fakeBankCardRepository struct {
	repository.DriverBankCardRepository
	card *model.DriverBankCard
}

func (f *fakeBankCardRepository) GetByID(_ context.Context, id uint64) (*model.DriverBankCard, error) {
	if f.card == nil || f.card.Id != id {
		return nil, nil
	}
	return f.card, nil
}

// fakeDriverRepository 嵌入接口，仅覆盖提现逻辑用到的 GetByID。
type fakeDriverRepository struct {
	repository.DriverRepository
	driver *model.Driver
}

func (f *fakeDriverRepository) GetByID(_ context.Context, id uint64) (*model.Driver, error) {
	if f.driver == nil || f.driver.Id != id {
		return nil, nil
	}
	return f.driver, nil
}
