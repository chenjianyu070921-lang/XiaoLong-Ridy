package svc

import (
	"context"
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
