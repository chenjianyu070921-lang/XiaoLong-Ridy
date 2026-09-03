package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListSettlementsReturnsDriverSettlementBills(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newTestSvcCtx(db, nil, nil)
	start := time.Unix(100, 0)
	end := time.Unix(200, 0)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `settlement` WHERE driver_id = \\? AND status = \\? AND settled_at >= \\? AND settled_at < \\?").
		WithArgs(uint64(25), int32(2), start, end).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT \\* FROM `settlement` WHERE driver_id = \\? AND status = \\? AND settled_at >= \\? AND settled_at < \\? ORDER BY settled_at DESC,id DESC LIMIT \\?").
		WithArgs(uint64(25), int32(2), start, end, 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "settlement_no", "order_id", "driver_id", "total_amount",
			"platform_commission_rate", "platform_commission", "driver_income",
			"status", "settled_at", "created_at",
		}).AddRow(uint64(8), "SET1001", uint64(1001), uint64(25), 12.50, 20.0, 2.50, 10.00, int8(2), start.Add(time.Minute), start))

	logic := NewListSettlementsLogic(context.Background(), svcCtx)
	resp, err := logic.ListSettlements(&proto.ListSettlementsRequest{
		DriverId: 25,
		Status:   2,
		StartAt:  start.Unix(),
		EndAt:    end.Unix(),
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("ListSettlements() error = %v", err)
	}
	if resp.GetTotal() != 1 || resp.GetPage() != 1 || resp.GetPageSize() != 20 || len(resp.GetRecords()) != 1 {
		t.Fatalf("ListSettlements() response = %+v", resp)
	}
	bill := resp.GetRecords()[0]
	if bill.GetSettlementId() != 8 || bill.GetSettlementNo() != "SET1001" || bill.GetOrderId() != 1001 ||
		bill.GetDriverId() != 25 || bill.GetDriverIncomeCents() != 1000 || bill.GetStatus() != 2 ||
		bill.GetSettledAt() != start.Add(time.Minute).Unix() {
		t.Fatalf("settlement bill = %+v", bill)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
