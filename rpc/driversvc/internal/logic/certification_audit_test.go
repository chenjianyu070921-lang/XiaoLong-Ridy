package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// newDriverGormMock 创建 driversvc 审核逻辑测试使用的 gorm/sqlmock 数据库。
// 测试不连接真实 MySQL，只断言 SQL 行为和事务边界，避免污染业务数据。
func newDriverGormMock(t *testing.T) (*svc.ServiceContext, sqlmock.Sqlmock, func()) {
	t.Helper()
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
	cleanup := func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
		_ = sqlDB.Close()
	}
	return &svc.ServiceContext{DB: db}, mock, cleanup
}

// TestApproveCertification_UpdatesCertificationDriverAndVehicle 验证审核通过后 driversvc 在本地事务中联动认证、司机和车辆状态。
func TestApproveCertification_UpdatesCertificationDriverAndVehicle(t *testing.T) {
	svcCtx, mock, cleanup := newDriverGormMock(t)
	defer cleanup()

	mock.ExpectQuery("SELECT driver_id, vehicle_id, audit_status AS status FROM `driver_certification` WHERE id = \\? LIMIT \\?").
		WithArgs(int64(3001), 1).
		WillReturnRows(sqlmock.NewRows([]string{"driver_id", "vehicle_id", "status"}).AddRow(uint64(2001), uint64(4001), int8(1)))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `driver_certification` SET .*`audit_remark`=\\?.*`audit_status`=\\?.*`audited_at`=\\?.*`audited_by`=\\?.*`updated_at`=\\?.* WHERE id = \\?").
		WithArgs("资料齐全", int8(2), sqlmock.AnyArg(), int64(9001), sqlmock.AnyArg(), int64(3001)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `driver` SET .*`status`=\\?.*`updated_at`=\\?.* WHERE id = \\?").
		WithArgs(2, sqlmock.AnyArg(), uint64(2001)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `driver_vehicle` SET .*`status`=\\?.*`updated_at`=\\?.* WHERE id = \\?").
		WithArgs(2, sqlmock.AnyArg(), uint64(4001)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	resp, err := NewApproveCertificationLogic(context.Background(), svcCtx).ApproveCertification(&proto.AuditCertificationRequest{
		CertificationId: 3001,
		OperatorId:      9001,
		Remark:          "资料齐全",
		Ip:              "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("ApproveCertification() error = %v", err)
	}
	if resp == nil || resp.GetMessage() != "ok" {
		t.Fatalf("ApproveCertification() response = %#v, want ok", resp)
	}
	_ = time.Now()
}

// TestRejectCertification_UpdatesCertificationOnly 验证审核驳回只更新认证审核状态，不激活司机和车辆。
func TestRejectCertification_UpdatesCertificationOnly(t *testing.T) {
	svcCtx, mock, cleanup := newDriverGormMock(t)
	defer cleanup()

	mock.ExpectQuery("SELECT driver_id, vehicle_id, audit_status AS status FROM `driver_certification` WHERE id = \\? LIMIT \\?").
		WithArgs(int64(3001), 1).
		WillReturnRows(sqlmock.NewRows([]string{"driver_id", "vehicle_id", "status"}).AddRow(uint64(2001), uint64(4001), int8(1)))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `driver_certification` SET .*`audit_remark`=\\?.*`audit_status`=\\?.*`audited_at`=\\?.*`audited_by`=\\?.*`updated_at`=\\?.* WHERE id = \\?").
		WithArgs("证件照片模糊", int8(3), sqlmock.AnyArg(), int64(9001), sqlmock.AnyArg(), int64(3001)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	resp, err := NewRejectCertificationLogic(context.Background(), svcCtx).RejectCertification(&proto.AuditCertificationRequest{
		CertificationId: 3001,
		OperatorId:      9001,
		Remark:          "证件照片模糊",
		Ip:              "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("RejectCertification() error = %v", err)
	}
	if resp == nil || resp.GetMessage() != "ok" {
		t.Fatalf("RejectCertification() response = %#v, want ok", resp)
	}
}

// TestRejectCertification_RequiresRemark 验证驳回审核必须填写原因，避免下游写入不可追责的空驳回记录。
func TestRejectCertification_RequiresRemark(t *testing.T) {
	logic := NewRejectCertificationLogic(context.Background(), &svc.ServiceContext{})
	resp, err := logic.RejectCertification(&proto.AuditCertificationRequest{
		CertificationId: 3001,
		OperatorId:      9001,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("RejectCertification() error code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
	if resp != nil {
		t.Fatalf("RejectCertification() response = %#v, want nil", resp)
	}
}
