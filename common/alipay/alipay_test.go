package alipay

import "testing"

func TestWithDefaults_Empty(t *testing.T) {
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
	c := Config{Sandbox: true}.WithDefaults()
	if c.Gateway != GatewaySandbox {
		t.Errorf("sandbox gateway = %q, want %q", c.Gateway, GatewaySandbox)
	}
}
