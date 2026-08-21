package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/rpc/usersvc/internal/config"
	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// TestLoginBySMS_RecordsUserBlacklistHit 验证登录用户命中黑名单时会写入真实登录场景记录。
func TestLoginBySMS_RecordsUserBlacklistHit(t *testing.T) {
	users := repository.NewMemoryUserRepository()
	if err := users.Create(context.Background(), &model.User{Phone: "13800138000", Nickname: "测试用户", Status: model.UserStatusNormal}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	risk := repository.NewMemoryRiskBlacklistRepository()
	risk.SetActive("user", 1, repository.BlacklistEntry{ID: 88, Reason: "历史风险命中"})
	sms := logicTestSMSVerifier{valid: true}
	ctx := svc.NewServiceContext(config.Config{}, users, repository.NewMemoryAddressRepository(), repository.NewMemoryCouponRepository(), risk, sms, sms, logicTestTokenManager{})

	_, err := NewLoginBySMSLogic(context.Background(), ctx).LoginBySMS(&userproto.LoginBySMSRequest{Phone: "13800138000", Code: "123456"})
	if err != nil {
		t.Fatalf("LoginBySMS() error = %v", err)
	}
	if len(risk.Hits) != 1 || risk.Hits[0].Scene != "login" || risk.Hits[0].BlacklistID != 88 || risk.Hits[0].TargetID != 1 {
		t.Fatalf("risk hit records = %+v, want one login record", risk.Hits)
	}
}

// logicTestSMSVerifier 为登录逻辑测试提供最小短信校验实现。
type logicTestSMSVerifier struct{ valid bool }

func (logicTestSMSVerifier) Send(context.Context, string) (int64, error) { return 60, nil }
func (v logicTestSMSVerifier) Verify(context.Context, string, string) (bool, error) {
	return v.valid, nil
}

// logicTestTokenManager 为登录逻辑测试提供最小令牌实现。
type logicTestTokenManager struct{}

func (logicTestTokenManager) Issue(uint64, string, int) (string, string, error) {
	return "token", "refresh", nil
}
func (logicTestTokenManager) Refresh(string) (string, string, error) { return "token", "refresh", nil }
func (logicTestTokenManager) Revoke(string) error                    { return nil }
