package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newDriverRepoGormMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
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
	return db, mock, cleanup
}

func TestCertificationRepository_UpdateAuditApprove(t *testing.T) {
	db, mock, cleanup := newDriverRepoGormMock(t)
	defer cleanup()

	repo := NewGormCertificationRepository(db)
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

	if err := repo.UpdateAudit(context.Background(), 3001, 9001, "资料齐全", 2); err != nil {
		t.Fatalf("UpdateAudit() error = %v", err)
	}
	_ = time.Now()
}

func TestCertificationRepository_UpdateAuditReject(t *testing.T) {
	db, mock, cleanup := newDriverRepoGormMock(t)
	defer cleanup()

	repo := NewGormCertificationRepository(db)
	mock.ExpectQuery("SELECT driver_id, vehicle_id, audit_status AS status FROM `driver_certification` WHERE id = \\? LIMIT \\?").
		WithArgs(int64(3001), 1).
		WillReturnRows(sqlmock.NewRows([]string{"driver_id", "vehicle_id", "status"}).AddRow(uint64(2001), uint64(4001), int8(1)))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `driver_certification` SET .*`audit_remark`=\\?.*`audit_status`=\\?.*`audited_at`=\\?.*`audited_by`=\\?.*`updated_at`=\\?.* WHERE id = \\?").
		WithArgs("证件照片模糊", int8(3), sqlmock.AnyArg(), int64(9001), sqlmock.AnyArg(), int64(3001)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.UpdateAudit(context.Background(), 3001, 9001, "证件照片模糊", 3); err != nil {
		t.Fatalf("UpdateAudit() error = %v", err)
	}
}

func TestCertificationRepository_UpdateAuditRequiresRemark(t *testing.T) {
	db, _, cleanup := newDriverRepoGormMock(t)
	defer cleanup()

	repo := NewGormCertificationRepository(db)
	if err := repo.UpdateAudit(context.Background(), 3001, 9001, "", 3); err == nil {
		t.Fatalf("UpdateAudit() error = nil, want invalid argument")
	}
}
