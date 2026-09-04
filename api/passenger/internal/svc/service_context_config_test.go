package svc

import "testing"

// TestApplyRuntimeDefaults_DefaultsToGRPC 验证乘客端未显式配置模式时默认连接真实下游服务。
func TestApplyRuntimeDefaults_DefaultsToGRPC(t *testing.T) {
	cfg := applyRuntimeDefaults(RuntimeConfig{})
	if cfg.ClientMode != clientModeGRPC {
		t.Fatalf("applyRuntimeDefaults().ClientMode = %q, want %q", cfg.ClientMode, clientModeGRPC)
	}
}

// TestApplyRuntimeDefaults_PreservesExplicitLocal 验证本地模式必须由调用方显式指定后才会保留。
func TestApplyRuntimeDefaults_PreservesExplicitLocal(t *testing.T) {
	cfg := applyRuntimeDefaults(RuntimeConfig{ClientMode: clientModeLocal})
	if cfg.ClientMode != clientModeLocal {
		t.Fatalf("applyRuntimeDefaults().ClientMode = %q, want %q", cfg.ClientMode, clientModeLocal)
	}
}
