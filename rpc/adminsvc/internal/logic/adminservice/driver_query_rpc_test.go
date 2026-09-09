package adminservicelogic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	driverproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListDrivers 记录管理后台转发给 driversvc 的司机列表请求，并返回可控响应。
func (f *fakeDriversClient) ListDrivers(ctx context.Context, in *driverproto.ListDriversRequest, opts ...grpc.CallOption) (*driverproto.ListDriversResponse, error) {
	f.listReq = in
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResp, nil
}

// GetDriver 记录管理后台转发给 driversvc 的司机详情请求，并返回可控响应。
func (f *fakeDriversClient) GetDriver(ctx context.Context, in *driverproto.GetDriverRequest, opts ...grpc.CallOption) (*driverproto.GetDriverResponse, error) {
	f.getReq = in
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getResp, nil
}

// FreezeDriver 记录管理后台转发给 driversvc 的冻结请求，并返回可控响应。
func (f *fakeDriversClient) FreezeDriver(ctx context.Context, in *driverproto.FreezeDriverRequest, opts ...grpc.CallOption) (*driverproto.CommonResponse, error) {
	f.freezeReq = in
	if f.freezeErr != nil {
		return nil, f.freezeErr
	}
	return &driverproto.CommonResponse{Message: "ok"}, nil
}

func TestListDrivers_CallsDriverService(t *testing.T) {
	svcCtx, _, cleanup := newAdminSQLMock(t)
	defer cleanup()
	driverClient := &fakeDriversClient{
		listResp: &driverproto.ListDriversResponse{
			Drivers: []*driverproto.Driver{{
				Id:              25,
				Phone:           "13800000000",
				RealName:        "张三",
				IdCardNo:        "110101199001011234",
				DriverLicenseNo: "DL10000001",
				AvatarUrl:       "https://example.com/avatar.png",
				Status:          driverproto.DriverStatus_DRIVER_STATUS_NORMAL,
				OnlineStatus:    1,
				VehicleId:       8001,
				PlateNo:         "京A12345",
				VehicleStatus:   2,
				CertificationId: 9001,
				AuditStatus:     2,
				AuditRemark:     "审核通过",
				CreatedAt:       1787143766,
				UpdatedAt:       1787143866,
			}},
			Total: 1,
		},
	}
	svcCtx.DriverSvc = driverClient

	resp, err := NewListDriversLogic(context.Background(), svcCtx).ListDrivers(&adminsvc.DriverListRequest{
		Page:     2,
		PageSize: 10,
		Keyword:  "张",
		Status:   2,
	})
	if err != nil {
		t.Fatalf("ListDrivers() error = %v", err)
	}
	if driverClient.listReq == nil {
		t.Fatal("ListDrivers() did not call driver service")
	}
	if driverClient.listReq.GetPage() != 2 || driverClient.listReq.GetPageSize() != 10 || driverClient.listReq.GetKeyword() != "张" || driverClient.listReq.GetStatus() != driverproto.DriverStatus_DRIVER_STATUS_NORMAL {
		t.Fatalf("driver service list request = %#v, want page/pageSize/keyword/status", driverClient.listReq)
	}
	if resp.GetTotal() != 1 || len(resp.GetList()) != 1 {
		t.Fatalf("ListDrivers() response = %#v, want one item", resp)
	}
	got := resp.GetList()[0]
	if got.GetId() != 25 || got.GetPhone() != "138****0000" || got.GetRealName() != "张三" || got.GetStatus() != 2 || got.GetOnlineStatus() != 1 {
		t.Fatalf("ListDrivers() mapped driver = %#v, want driver service fields", got)
	}
	if got.GetIdCardNo() != "110101********1234" || got.GetDriverLicenseNo() != "DL10000001" {
		t.Fatalf("ListDrivers() sensitive fields = %#v, want masked id card and safe short license value", got)
	}
	if got.GetVehicleId() != 8001 || got.GetPlateNo() != "京A12345" || got.GetCertificationId() != 9001 || got.GetAuditStatus() != 2 {
		t.Fatalf("ListDrivers() aggregate fields = %#v, want driversvc vehicle/certification fields", got)
	}
}

func TestGetDriver_CallsDriverService(t *testing.T) {
	svcCtx, _, cleanup := newAdminSQLMock(t)
	defer cleanup()
	driverClient := &fakeDriversClient{
		getResp: &driverproto.GetDriverResponse{
			Driver: &driverproto.Driver{
				Id:              26,
				Phone:           "13900000000",
				RealName:        "李四",
				IdCardNo:        "110101199001011235",
				DriverLicenseNo: "DL10000002",
				Status:          driverproto.DriverStatus_DRIVER_STATUS_PENDING,
				OnlineStatus:    0,
				VehicleId:       8002,
				PlateNo:         "沪B12345",
				VehicleStatus:   1,
				CertificationId: 9002,
				AuditStatus:     1,
			},
		},
	}
	svcCtx.DriverSvc = driverClient

	resp, err := NewGetDriverLogic(context.Background(), svcCtx).GetDriver(&adminsvc.DriverDetailRequest{Id: 26})
	if err != nil {
		t.Fatalf("GetDriver() error = %v", err)
	}
	if driverClient.getReq == nil || driverClient.getReq.GetId() != 26 {
		t.Fatalf("driver service get request = %#v, want id 26", driverClient.getReq)
	}
	if resp.GetId() != 26 || resp.GetPhone() != "139****0000" || resp.GetRealName() != "李四" || resp.GetStatus() != 1 {
		t.Fatalf("GetDriver() mapped driver = %#v, want driver service fields", resp)
	}
	if resp.GetIdCardNo() != "110101********1235" || resp.GetDriverLicenseNo() != "DL10000002" {
		t.Fatalf("GetDriver() sensitive fields = %#v, want masked id card and safe short license value", resp)
	}
	if resp.GetVehicleId() != 8002 || resp.GetPlateNo() != "沪B12345" || resp.GetCertificationId() != 9002 || resp.GetAuditStatus() != 1 {
		t.Fatalf("GetDriver() aggregate fields = %#v, want driversvc vehicle/certification fields", resp)
	}
}

func TestGetDriver_RevealsSensitiveOnlyWhenRequestedByOpsRole(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	driverClient := &fakeDriversClient{
		getResp: &driverproto.GetDriverResponse{
			Driver: &driverproto.Driver{
				Id:              26,
				Phone:           "13900000000",
				RealName:        "李四",
				IdCardNo:        "110101199001011235",
				DriverLicenseNo: "DL100000021234",
				Status:          driverproto.DriverStatus_DRIVER_STATUS_PENDING,
			},
		},
	}
	svcCtx.DriverSvc = driverClient
	server := miniredis.RunT(t)
	defer server.Close()
	svcCtx.Redis = redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer svcCtx.Redis.Close()

	ctx := contextWithAdminSession(t, svcCtx, 2001, 2)
	mock.ExpectQuery(`SELECT id, username, password_hash, real_name, role, status\s+FROM admin_user\s+WHERE id = \? AND deleted_at IS NULL`).
		WithArgs(int64(2001)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "real_name", "role", "status"}).
			AddRow(2001, "ops", "hash", "运营", 2, 1))
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(2001), "driver", "view_sensitive", "driver", int64(26), "查看司机完整手机号、身份证号和驾驶证号", "").
		WillReturnResult(sqlmock.NewResult(99, 1))

	resp, err := NewGetDriverLogic(ctx, svcCtx).GetDriver(&adminsvc.DriverDetailRequest{Id: 26, Sensitive: true})
	if err != nil {
		t.Fatalf("GetDriver() error = %v", err)
	}
	if resp.GetPhone() != "13900000000" || resp.GetIdCardNo() != "110101199001011235" || resp.GetDriverLicenseNo() != "DL100000021234" {
		t.Fatalf("GetDriver() sensitive response = %+v", resp)
	}
}

func TestDriverQueries_DownstreamDisabled(t *testing.T) {
	svcCtx, _, cleanup := newAdminSQLMock(t)
	defer cleanup()
	svcCtx.DriverSvc = nil

	if _, err := NewListDriversLogic(context.Background(), svcCtx).ListDrivers(&adminsvc.DriverListRequest{}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("ListDrivers() error code = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := NewGetDriverLogic(context.Background(), svcCtx).GetDriver(&adminsvc.DriverDetailRequest{Id: 1}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("GetDriver() error code = %v, want FailedPrecondition", status.Code(err))
	}
}
