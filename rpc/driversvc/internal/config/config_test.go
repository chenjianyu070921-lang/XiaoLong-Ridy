package config

import (
	"strings"
	"testing"
)

func TestConfigValidateSigningKeyRejectsEmptySigningKey(t *testing.T) {
	cfg := Config{}

	err := cfg.ValidateSigningKey()
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("ValidateSigningKey() error = %v, want empty signing key rejection", err)
	}
}

func TestConfigValidateSigningKeyRejectsDefaultSigningKey(t *testing.T) {
	cfg := Config{SigningKey: defaultSigningKey}

	err := cfg.ValidateSigningKey()
	if err == nil || !strings.Contains(err.Error(), "default") {
		t.Fatalf("ValidateSigningKey() error = %v, want default signing key rejection", err)
	}
}

func TestConfigApplyRuntimeSigningKeyUsesEnvOverride(t *testing.T) {
	t.Setenv("DRIVERSVC_SIGNING_KEY", "runtime-driver-signing-key")
	cfg := Config{SigningKey: defaultSigningKey}

	cfg.ApplyRuntimeSigningKey()

	if cfg.SigningKey != "runtime-driver-signing-key" {
		t.Fatalf("ApplyRuntimeSigningKey() signing key = %q, want env override", cfg.SigningKey)
	}
	if err := cfg.ValidateSigningKey(); err != nil {
		t.Fatalf("ValidateSigningKey() after env override error = %v", err)
	}
}
