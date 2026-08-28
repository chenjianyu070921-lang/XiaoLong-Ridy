package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGormDriverRepositoryListSearchesVehicleAndIdentityFields(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(DISTINCT(`driver`.`id`)) FROM `driver` LEFT JOIN driver_vehicle v ON v.driver_id = driver.id WHERE driver.phone LIKE ? OR driver.real_name LIKE ? OR driver.id_card_no LIKE ? OR driver.driver_license_no LIKE ? OR v.plate_no LIKE ?")).
		WithArgs("%粤B12345%", "%粤B12345%", "%粤B12345%", "%粤B12345%", "%粤B12345%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT driver.* FROM `driver` LEFT JOIN driver_vehicle v ON v.driver_id = driver.id WHERE driver.phone LIKE ? OR driver.real_name LIKE ? OR driver.id_card_no LIKE ? OR driver.driver_license_no LIKE ? OR v.plate_no LIKE ? GROUP BY `driver`.`id` ORDER BY driver.id DESC LIMIT ?")).
		WithArgs("%粤B12345%", "%粤B12345%", "%粤B12345%", "%粤B12345%", "%粤B12345%", 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "phone", "password_hash", "real_name", "id_card_no", "driver_license_no", "avatar_url", "status", "online_status", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(25), "13800138000", "", "driver", "440300199001011234", "DL-25", "", int8(2), int8(1), time.Unix(100, 0), time.Unix(200, 0), nil))

	drivers, total, err := NewGormDriverRepository(db).List(context.Background(), DriverListFilter{
		Page:     1,
		PageSize: 20,
		Keyword:  "粤B12345",
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 1 || len(drivers) != 1 || drivers[0].Id != 25 {
		t.Fatalf("List() total=%d drivers=%+v", total, drivers)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
