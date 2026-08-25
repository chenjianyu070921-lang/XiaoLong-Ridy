package adminservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestExecuteAdminWriteTx_CommitsBusinessAndAudit 验证管理员业务变更与审计日志在同一事务中提交。
func TestExecuteAdminWriteTx_CommitsBusinessAndAudit(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE admin_user SET status=\?, updated_at=NOW\(\) WHERE id=\? AND deleted_at IS NULL`).
		WithArgs(int32(2), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "admin", "disable", "admin_user", int64(1001), "停用管理员", "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	logic := &AdminManagementLogic{ctx: context.Background(), svcCtx: svcCtx}
	err := logic.executeAdminWriteTx(func(tx *sql.Tx) error {
		result, err := tx.ExecContext(context.Background(),
			"UPDATE admin_user SET status=?, updated_at=NOW() WHERE id=? AND deleted_at IS NULL",
			int32(2), int64(1001))
		if err != nil {
			return err
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			t.Fatalf("RowsAffected() = %d, want 1", affected)
		}
		return createOperationLogTx(context.Background(), tx, 9001, "admin", "disable", "admin_user", 1001, "停用管理员", "127.0.0.1")
	})
	if err != nil {
		t.Fatalf("executeAdminWriteTx() error = %v", err)
	}
}

// TestExecuteAdminWriteTx_RollsBackBusinessWhenAuditFails 验证审计失败时管理员业务变更会回滚。
func TestExecuteAdminWriteTx_RollsBackBusinessWhenAuditFails(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE admin_user SET status=\?, updated_at=NOW\(\) WHERE id=\? AND deleted_at IS NULL`).
		WithArgs(int32(2), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "admin", "disable", "admin_user", int64(1001), "停用管理员", "127.0.0.1").
		WillReturnError(errors.New("operation log write failed"))
	mock.ExpectRollback()

	logic := &AdminManagementLogic{ctx: context.Background(), svcCtx: svcCtx}
	err := logic.executeAdminWriteTx(func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(context.Background(),
			"UPDATE admin_user SET status=?, updated_at=NOW() WHERE id=? AND deleted_at IS NULL",
			int32(2), int64(1001)); err != nil {
			return err
		}
		return createOperationLogTx(context.Background(), tx, 9001, "admin", "disable", "admin_user", 1001, "停用管理员", "127.0.0.1")
	})
	if err == nil {
		t.Fatal("executeAdminWriteTx() error = nil, want audit error")
	}
}
