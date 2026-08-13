package logic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
)

type fakeVerifier struct {
	ok  bool
	err error
}

// Verify 返回测试用例预先配置好的结果。
func (v fakeVerifier) Verify(context.Context, string, string) (bool, error) {
	return v.ok, v.err
}

func TestRegisterSuccess(t *testing.T) {
	repo := repository.NewMemoryUserRepository()
	logic := NewRegisterLogic(repo, fakeVerifier{ok: true})

	resp, err := logic.Register(context.Background(), RegisterRequest{
		Phone:   "13800138000",
		SMSCode: "123456",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if resp.UserID == 0 {
		t.Fatal("Register() returned empty user id")
	}
	if resp.Nickname != "138****8000" {
		t.Fatalf("Register() nickname = %q", resp.Nickname)
	}
	if resp.RegisterSource != "phone" {
		t.Fatalf("Register() source = %q", resp.RegisterSource)
	}
}

func TestRegisterRejectsDuplicatePhone(t *testing.T) {
	repo := repository.NewMemoryUserRepository()
	logic := NewRegisterLogic(repo, fakeVerifier{ok: true})

	req := RegisterRequest{Phone: "13800138000", SMSCode: "123456"}
	if _, err := logic.Register(context.Background(), req); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	_, err := logic.Register(context.Background(), req)
	if !errors.Is(err, ErrPhoneAlreadyExists) {
		t.Fatalf("second Register() error = %v, want %v", err, ErrPhoneAlreadyExists)
	}
}

func TestRegisterRejectsInvalidSMSCode(t *testing.T) {
	repo := repository.NewMemoryUserRepository()
	logic := NewRegisterLogic(repo, fakeVerifier{ok: false})

	_, err := logic.Register(context.Background(), RegisterRequest{
		Phone:   "13800138000",
		SMSCode: "000000",
	})
	if !errors.Is(err, ErrInvalidSMSCode) {
		t.Fatalf("Register() error = %v, want %v", err, ErrInvalidSMSCode)
	}
}
