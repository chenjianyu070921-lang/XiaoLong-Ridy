package driver

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newMockReviewStore(t *testing.T) (*GormReviewStore, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() err=%v", err)
	}
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() err=%v", err)
	}
	store, err := NewGormReviewStore(gormDB)
	if err != nil {
		t.Fatalf("NewGormReviewStore() err=%v", err)
	}
	return store, mock
}

// Save：写入乘客评价并携带全部字段。
func TestGormReviewStoreSave(t *testing.T) {
	store, mock := newMockReviewStore(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `passenger_review`").
		WillReturnResult(sqlmock.NewResult(7, 1))
	mock.ExpectCommit()

	err := store.Save(context.Background(), &PassengerReview{
		DriverPhone:    "13800000001",
		OrderID:        "60",
		PassengerPhone: "13900000002",
		Rating:         2,
		Comment:        "绕路还态度差",
		Tags:           "司机绕路,司机态度差",
		CreatedAt:      time.Now(),
	})
	if err != nil {
		t.Fatalf("Save() err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL 期望未满足: %v", err)
	}
}

// Save：参数校验（司机手机号必填、星级 0-5）。
func TestGormReviewStoreSaveValidation(t *testing.T) {
	store, _ := newMockReviewStore(t)
	if err := store.Save(context.Background(), &PassengerReview{Rating: 3}); err == nil {
		t.Fatal("缺少司机手机号应报错")
	}
	if err := store.Save(context.Background(), &PassengerReview{DriverPhone: "138", Rating: 6}); err == nil {
		t.Fatal("星级 6 应报错")
	}
	if err := store.Save(context.Background(), nil); err == nil {
		t.Fatal("nil 评价应报错")
	}
}

// ListByDriverPhone：按司机手机号查询、created_at 倒序、limit 生效。
func TestGormReviewStoreListByDriverPhone(t *testing.T) {
	store, mock := newMockReviewStore(t)
	rows := sqlmock.NewRows([]string{"id", "driver_phone", "order_id", "passenger_phone", "rating", "comment", "tags", "created_at"}).
		AddRow(2, "13800000001", "60", "13900000002", 1, "态度很差", "司机态度差", time.Now()).
		AddRow(1, "13800000001", "59", "13900000003", 5, "服务很好", "服务态度好", time.Now().Add(-time.Hour))
	mock.ExpectQuery("SELECT \\* FROM `passenger_review` WHERE driver_phone = \\? ORDER BY created_at DESC LIMIT \\?").
		WithArgs("13800000001", 20).
		WillReturnRows(rows)

	list, err := store.ListByDriverPhone(context.Background(), "13800000001", 20)
	if err != nil {
		t.Fatalf("ListByDriverPhone() err=%v", err)
	}
	if len(list) != 2 || list[0].Comment != "态度很差" {
		t.Fatalf("list=%+v", list)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL 期望未满足: %v", err)
	}
}

func TestGormReviewStoreListValidation(t *testing.T) {
	store, _ := newMockReviewStore(t)
	if _, err := store.ListByDriverPhone(context.Background(), " ", 10); err == nil {
		t.Fatal("空手机号应报错")
	}
}

type stubReviewStore struct {
	reviews []PassengerReview
	err     error
}

func (s *stubReviewStore) Save(context.Context, *PassengerReview) error { return nil }

func (s *stubReviewStore) ListByDriverPhone(context.Context, string, int) ([]PassengerReview, error) {
	return s.reviews, s.err
}

// 端到端：Agent 按手机号取乘客评价 → 拼接 → 规则引擎打分。
func TestRateByDriverPhoneScoresFromStoredReviews(t *testing.T) {
	a := NewAgent(nil, nil) // 模型不可用走规则引擎，保证结果确定性
	store := &stubReviewStore{reviews: []PassengerReview{
		{Comment: "车内干净，服务态度好"},
		{Comment: "接驾迟到了"},
	}}
	got, err := a.RateByDriverPhone(context.Background(), store, "13800000001", 20)
	if err != nil {
		t.Fatalf("RateByDriverPhone() err=%v", err)
	}
	if got.Star != 3 { // 迟到(3星档) + 正向标签并存 → 3星
		t.Fatalf("star=%d reason=%s", got.Star, got.Reason)
	}
	found := map[string]bool{}
	for _, tag := range got.Tags {
		found[tag] = true
	}
	if !found[TagLatePickup] || !found[TagCleanCar] || !found[TagGoodAttitude] {
		t.Fatalf("tags=%v", got.Tags)
	}
}

// 评价拼接限制：单条 80 字截断、总长 600 字截断。
func TestJoinPassengerCommentsTruncation(t *testing.T) {
	long := strings.Repeat("评", 100)
	joined := joinPassengerComments([]PassengerReview{{Comment: long}, {Comment: long}, {Comment: long}})
	if n := len([]rune(joined)); n > 600 {
		t.Fatalf("拼接总长应 ≤600, got %d", n)
	}
}

// 仓储查询失败错误透传。
func TestRateByDriverPhoneStoreError(t *testing.T) {
	a := NewAgent(nil, nil)
	store := &stubReviewStore{err: errors.New("db down")}
	if _, err := a.RateByDriverPhone(context.Background(), store, "13800000001", 20); err == nil {
		t.Fatal("仓储错误应透传")
	}
}

// 入参校验。
func TestRateByDriverPhoneValidation(t *testing.T) {
	a := NewAgent(nil, nil)
	if _, err := a.RateByDriverPhone(context.Background(), nil, "138", 10); err == nil {
		t.Fatal("缺少 store 应报错")
	}
	if _, err := a.RateByDriverPhone(context.Background(), &stubReviewStore{}, "  ", 10); err == nil {
		t.Fatal("空手机号应报错")
	}
}
