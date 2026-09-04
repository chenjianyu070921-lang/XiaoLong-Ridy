package svc

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRPCContextDetachesCancellationAndSetsDeadline(t *testing.T) {
	type traceKey struct{}

	parent := context.WithValue(context.Background(), traceKey{}, "trace-123")
	canceled, cancel := context.WithCancel(parent)
	cancel()

	ctx, done := rpcContext(canceled)
	defer done()

	if ctx.Err() != nil {
		t.Fatalf("rpcContext() should detach cancellation, got err=%v", ctx.Err())
	}
	if got := ctx.Value(traceKey{}); got != "trace-123" {
		t.Fatalf("rpcContext() value = %v, want trace-123", got)
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("rpcContext() should set a deadline")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > defaultRPCTimeout+time.Second {
		t.Fatalf("rpcContext() deadline remaining = %v, want within timeout window", remaining)
	}
}

func TestValidateSigningKeyRejectsDefaultSigningKey(t *testing.T) {
	t.Setenv("DRIVERSVC_SIGNING_KEY", "")

	ctx := &ServiceContext{SigningKey: defaultSigningKey}

	err := ctx.ValidateSigningKey()
	if err == nil || !strings.Contains(err.Error(), "default") {
		t.Fatalf("ValidateSigningKey() error = %v, want default signing key rejection", err)
	}
}

func TestResolveSigningKeyFallsBackToLocalKeyWhenEnvMissing(t *testing.T) {
	t.Setenv("DRIVER_SIGNING_KEY", "")

	if got := resolveSigningKey(); got != localFallbackSigningKey {
		t.Fatalf("resolveSigningKey() = %q, want local fallback key %q", got, localFallbackSigningKey)
	}
}
