package adminservicelogic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	priceclient "XiaoLong-Ridy/rpc/pricesvc/price"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc"
)

// fakePriceClient 仅覆盖计价规则创建调用，用于验证 adminsvc 的审计补偿语义。
type fakePriceClient struct {
	priceclient.Price
	createResponse *priceclient.CreatePriceRuleResponse
}

// CreatePriceRule 返回下游 pricesvc 分配的真实规则 ID。
func (f *fakePriceClient) CreatePriceRule(ctx context.Context, in *priceclient.PriceRuleRequest, opts ...grpc.CallOption) (*priceclient.CreatePriceRuleResponse, error) {
	return f.createResponse, nil
}

// TestCreatePriceRule_AuditFailureWritesOutbox 验证跨服务规则已生效后，审计失败会落 outbox 且接口仍返回成功。
func TestCreatePriceRule_AuditFailureWritesOutbox(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	svcCtx.PricesSvc = &fakePriceClient{createResponse: &priceclient.CreatePriceRuleResponse{Id: 7001, Message: "ok"}}

	mock.ExpectExec(`INSERT INTO admin_operation_log`).WillReturnError(errors.New("audit unavailable"))
	mock.ExpectExec(`INSERT INTO admin_audit_outbox`).
		WithArgs(sqlmock.AnyArg(), "price", "create", "price_rule", int64(7001), int64(9001), "创建计价规则：标准快车", "127.0.0.1", "audit unavailable").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := NewCreatePriceRuleLogic(context.Background(), svcCtx).CreatePriceRule(&adminsvc.PriceRuleRequest{
		Name: "标准快车", AdminId: 9001, Ip: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("CreatePriceRule() error = %v", err)
	}
	if resp.GetId() != 7001 {
		t.Fatalf("CreatePriceRule() id = %d, want 7001", resp.GetId())
	}
}
