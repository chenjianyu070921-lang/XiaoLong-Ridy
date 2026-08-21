package alipay

import (
	"os"
	"testing"
)

// clearAlipayEnv 清理所有以 ALIPAY_ 开头的环境变量，避免污染测试。
func clearAlipayEnv(t *testing.T) {
	t.Helper()
	for _, env := range []string{
		"ALIPAY_APP_ID",
		"ALIPAY_PRIVATE_KEY",
		"ALIPAY_PUBLIC_KEY",
		"ALIPAY_GATEWAY",
		"ALIPAY_NOTIFY_URL",
		"ALIPAY_RETURN_URL",
	} {
		old := os.Getenv(env)
		if old != "" {
			os.Unsetenv(env)
			t.Cleanup(func() { os.Setenv(env, old) })
		} else {
			t.Cleanup(func() {})
		}
	}
}

func TestWithDefaults_Empty(t *testing.T) {
	clearAlipayEnv(t)
	c := Config{AppId: "test-app"}.WithDefaults()
	if c.Gateway != GatewayProduction {
		t.Errorf("gateway = %q, want %q", c.Gateway, GatewayProduction)
	}
	if c.SignType != SignTypeRSA2 {
		t.Errorf("signType = %q, want %q", c.SignType, SignTypeRSA2)
	}
	if c.Charset != CharsetUTF8 {
		t.Errorf("charset = %q, want %q", c.Charset, CharsetUTF8)
	}
	if c.AppId != "test-app" {
		t.Errorf("appId = %q, want test-app", c.AppId)
	}
}

func TestWithDefaults_Explicit(t *testing.T) {
	clearAlipayEnv(t)
	c := Config{Gateway: "https://custom", SignType: SignTypeRSA, Charset: "gbk"}.WithDefaults()
	if c.Gateway != "https://custom" {
		t.Errorf("gateway should preserve explicit value")
	}
	if c.SignType != SignTypeRSA {
		t.Errorf("signType should preserve explicit value")
	}
	if c.Charset != "gbk" {
		t.Errorf("charset should preserve explicit value")
	}
}

func TestWithDefaults_Sandbox(t *testing.T) {
	clearAlipayEnv(t)
	c := Config{Sandbox: true}.WithDefaults()
	if c.Gateway != GatewaySandbox {
		t.Errorf("sandbox gateway = %q, want %q", c.Gateway, GatewaySandbox)
	}
}

// TestFromEnv_OverridesYAML 验证 M5-8：环境变量覆盖 yaml 字段，避免明文密钥入库。
func TestFromEnv_OverridesYAML(t *testing.T) {
	clearAlipayEnv(t)
	os.Setenv("ALIPAY_APP_ID", "from-env")
	os.Setenv("ALIPAY_PRIVATE_KEY", "pk-env")
	os.Setenv("ALIPAY_PUBLIC_KEY", "pub-env")
	t.Cleanup(func() {
		os.Unsetenv("ALIPAY_APP_ID")
		os.Unsetenv("ALIPAY_PRIVATE_KEY")
		os.Unsetenv("ALIPAY_PUBLIC_KEY")
	})

	c := Config{AppId: "from-yaml", PrivateKey: "from-yaml"}.FromEnv()
	if c.AppId != "from-env" {
		t.Errorf("AppId = %q, want from-env", c.AppId)
	}
	if c.PrivateKey != "pk-env" {
		t.Errorf("PrivateKey = %q, want pk-env", c.PrivateKey)
	}
	if c.AlipayPublicKey != "pub-env" {
		t.Errorf("AlipayPublicKey = %q, want pub-env", c.AlipayPublicKey)
	}
}

// TestHasRealKeys 验证 M5-3 的判定函数。
func TestHasRealKeys(t *testing.T) {
	clearAlipayEnv(t)

	cases := []struct {
		name     string
		cfg      Config
		envAppId string
		want     bool
	}{
		{"全部为空", Config{}, "", false},
		{"缺公钥", Config{AppId: "x", PrivateKey: "x"}, "", false},
		{"全配置", Config{AppId: "x", PrivateKey: "x", AlipayPublicKey: "x"}, "", true},
		{"环境变量补齐", Config{}, "x", false}, // env 缺 PrivateKey
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			clearAlipayEnv(t)
			if c.envAppId != "" {
				os.Setenv("ALIPAY_APP_ID", c.envAppId)
			}
			if got := c.cfg.HasRealKeys(); got != c.want {
				t.Errorf("HasRealKeys() = %v, want %v", got, c.want)
			}
		})
	}
}
