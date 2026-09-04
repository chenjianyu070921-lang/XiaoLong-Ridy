package svc

import (
	"testing"

	"XiaoLong-Ridy/common/alipay"
	"XiaoLong-Ridy/rpc/paysvc/internal/channel"

	"github.com/zeromicro/go-zero/core/service"
)

// TestNewAlipayChannel_ProductionMissingKeys 生产模式（mode 为空串）缺密钥必须返回 error（M5-12 硬校验）。
func TestNewAlipayChannel_ProductionMissingKeys(t *testing.T) {
	ch, err := newAlipayChannel(alipay.Config{}, "") // 生产模式 = mode 为空
	if err == nil {
		t.Fatal("want error when production mode and keys missing, got nil")
	}
	if ch != nil {
		t.Fatalf("want nil channel on error, got %v", ch)
	}
}

// TestNewAlipayChannel_DevModeFallback dev/test 模式缺密钥允许降级（返回 nil, nil）。
func TestNewAlipayChannel_DevModeFallback(t *testing.T) {
	ch, err := newAlipayChannel(alipay.Config{}, service.TestMode)
	if err != nil {
		t.Fatalf("want no error in test mode, got %v", err)
	}
	if ch != nil {
		t.Fatalf("want nil channel in test mode fallback, got %v", ch)
	}
}

// TestGetChannel_AlipayNilFallsBackToMock alipayChannel 为 nil 时 GetChannel 返回 MockChannel。
func TestGetChannel_AlipayNilFallsBackToMock(t *testing.T) {
	s := &ServiceContext{alipayChannel: nil}
	ch := s.GetChannel(channel.Alipay)
	if ch == nil {
		t.Fatal("want non-nil channel")
	}
	if ch.Name() != channel.Alipay {
		// MockChannel 的 Name 返回传入的名字，此处验证确实走了 mock 兜底（Name 匹配渠道名）。
		t.Fatalf("want fallback channel, got name=%s", ch.Name())
	}
}

// TestGetChannel_AlipayReal 生产模式 alipayChannel 非 nil 时 GetChannel 返回真实渠道。
// 用 channel.AlipayChannel 类型，但要避免触发真实 SDK 网络调用，
// 这里直接构造一个零值 AlipayChannel（不调用其方法，仅验证返回同一指针）。
func TestGetChannel_AlipayReal(t *testing.T) {
	real := &channel.AlipayChannel{}
	s := &ServiceContext{alipayChannel: real}
	ch := s.GetChannel(channel.Alipay)
	if ch != real {
		t.Fatal("want real alipay channel")
	}
}
