package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGormDriverRepositoryRefreshDriverScoreMetricsUsesFixedSources(t *testing.T) {
	db, mock, cleanup := newDriverRepoGormMock(t)
	defer cleanup()

	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `ride_order` WHERE driver_id = \\? AND status = \\? AND updated_at >= \\? AND updated_at < \\?").
		WithArgs(uint64(25), 5, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `ride_order` WHERE driver_id = \\? AND status = \\? AND updated_at >= \\? AND updated_at < \\?").
		WithArgs(uint64(25), 6, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT AVG\\(rating\\) FROM `order_review` WHERE driver_id = \\? AND created_at >= \\? AND created_at < \\?").
		WithArgs(uint64(25), start, end).
		WillReturnRows(sqlmock.NewRows([]string{"avg"}).AddRow(4.0))
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `admin_complaint_work_order` WHERE driver_id = \\? AND created_at >= \\? AND created_at < \\?").
		WithArgs(uint64(25), start, end).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery("SELECT \\* FROM `driver_score` WHERE driver_id = \\? ORDER BY `driver_score`.`id` LIMIT \\?").
		WithArgs(uint64(25), 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "driver_id", "score", "level", "month_orders",
			"month_cancel_rate", "month_complaint_count", "updated_at",
		}).AddRow(uint64(7), uint64(25), 100.0, int8(5), 0, 0.0, 0, start))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `driver_score` SET .*`level`=\\?.*`month_cancel_rate`=\\?.*`month_complaint_count`=\\?.*`month_orders`=\\?.*`score`=\\?.*`updated_at`=\\?.* WHERE driver_id = \\?").
		WithArgs(int8(3), 25.0, 2, 3, 80.0, sqlmock.AnyArg(), uint64(25)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	score, err := NewGormDriverRepository(db).RefreshDriverScoreMetrics(context.Background(), 25, start, end)
	if err != nil {
		t.Fatalf("RefreshDriverScoreMetrics() error = %v", err)
	}
	if score.Score != 80 || score.Level != 3 || score.MonthOrders != 3 ||
		score.MonthCancelRate != 25 || score.MonthComplaintCount != 2 {
		t.Fatalf("score = %+v", score)
	}
}
