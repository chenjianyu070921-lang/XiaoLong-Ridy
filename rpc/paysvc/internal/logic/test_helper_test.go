package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/common/mq"
	order "XiaoLong-Ridy/rpc/ordersvc/proto"
	"XiaoLong-Ridy/rpc/paysvc/internal/channel"
	"XiaoLong-Ridy/rpc/paysvc/internal/orderclient"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// newMockDB 创建基于 sqlmock 的 GORM DB。
func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return db, mock
}

// newTestSvcCtx 构造测试用 ServiceContext。
func newTestSvcCtx(db *gorm.DB, oc orderclient.OrderClient, verifier channel.SignVerifier) *svc.ServiceContext {
	if oc == nil {
		oc = &mockOrderClient{}
	}
	if verifier == nil {
		verifier = &channel.MockVerifier{}
	}
	return &svc.ServiceContext{
		DB:          db,
		Producer:    &mq.NoopProducer{},
		OrderClient: oc,
		Verifier:    verifier,
	}
}

// mockOrderClient 测试用订单客户端。
type mockOrderClient struct {
	driverId   int64
	err        error
	confirmed  []*order.ConfirmPaidRequest
	confirmErr error
}

func (m *mockOrderClient) GetDriverId(ctx context.Context, orderId int64) (int64, error) {
	return m.driverId, m.err
}

func (m *mockOrderClient) ConfirmPaid(ctx context.Context, in *order.ConfirmPaidRequest) (*order.ConfirmPaidResponse, error) {
	m.confirmed = append(m.confirmed, in)
	if m.confirmErr != nil {
		return nil, m.confirmErr
	}
	return &order.ConfirmPaidResponse{OrderId: in.OrderId, Status: order.OrderStatus_ORDER_STATUS_COMPLETED}, nil
}
