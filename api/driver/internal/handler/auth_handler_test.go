package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWriteAuthErrorMapsTransportFailuresToBadGateway(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{name: "unavailable", err: status.Error(codes.Unavailable, "downstream unavailable")},
		{name: "deadline exceeded", err: status.Error(codes.DeadlineExceeded, "deadline exceeded")},
		{name: "unimplemented", err: status.Error(codes.Unimplemented, "unknown method")},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			writeAuthError(recorder, testCase.err)
			if recorder.Code != http.StatusBadGateway {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadGateway)
			}
		})
	}
}
