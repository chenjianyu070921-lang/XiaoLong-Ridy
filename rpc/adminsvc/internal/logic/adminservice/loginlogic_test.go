package adminservicelogic

import "testing"

// TestLoginFailureExceeded 验证登录失败锁定阈值判断。
// 失败次数达到 maxLoginFailures 后应返回锁定，低于阈值时允许继续尝试。
func TestLoginFailureExceeded(t *testing.T) {
	cases := []struct {
		name     string
		failures int64
		want     bool
	}{
		{name: "无失败记录不锁定", failures: 0, want: false},
		{name: "失败次数低于阈值不锁定", failures: maxLoginFailures - 1, want: false},
		{name: "失败次数达到阈值锁定", failures: maxLoginFailures, want: true},
		{name: "失败次数超过阈值保持锁定", failures: maxLoginFailures + 3, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := loginFailureExceeded(tc.failures); got != tc.want {
				t.Fatalf("loginFailureExceeded(%d) = %v, want %v", tc.failures, got, tc.want)
			}
		})
	}
}

// TestLoginFailKey 验证失败计数 key 的前缀隔离，避免与其他 Redis key 冲突。
func TestLoginFailKey(t *testing.T) {
	got := loginFailKey("admin")
	want := "admin:login:fail:admin"
	if got != want {
		t.Fatalf("loginFailKey(admin) = %q, want %q", got, want)
	}
}
