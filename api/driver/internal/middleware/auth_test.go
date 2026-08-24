package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/common/jwtx"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

func TestRequireAuthRejectsFrozenDriverToken(t *testing.T) {
	token := signDriverToken(t, int(driversproto.DriverStatus_DRIVER_STATUS_FROZEN))
	called := false
	handler := RequireAuth(&svc.ServiceContext{SigningKey: "middleware-test-key"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusForbidden, recorder.Body.String())
	}
	if called {
		t.Fatal("next handler should not be called")
	}
}

func TestRequireAuthAllowsNormalDriverToken(t *testing.T) {
	token := signDriverToken(t, int(driversproto.DriverStatus_DRIVER_STATUS_NORMAL))
	called := false
	handler := RequireAuth(&svc.ServiceContext{SigningKey: "middleware-test-key"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if ClaimsFromContext(r.Context()) == nil {
			t.Fatal("claims missing from request context")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusNoContent, recorder.Body.String())
	}
	if !called {
		t.Fatal("next handler was not called")
	}
}

func TestRequireAuthAllowsPendingDriverToCompleteProfile(t *testing.T) {
	token := signDriverToken(t, int(driversproto.DriverStatus_DRIVER_STATUS_PENDING))
	called := false
	handler := RequireAuth(&svc.ServiceContext{SigningKey: "middleware-test-key"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/driver/v1/vehicles", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusNoContent, recorder.Body.String())
	}
	if !called {
		t.Fatal("next handler was not called")
	}
}

func TestRequireAuthRejectsPendingDriverForOnlineBusiness(t *testing.T) {
	token := signDriverToken(t, int(driversproto.DriverStatus_DRIVER_STATUS_PENDING))
	called := false
	handler := RequireAuth(&svc.ServiceContext{SigningKey: "middleware-test-key"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/driver/v1/drivers/online", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusForbidden, recorder.Body.String())
	}
	if called {
		t.Fatal("next handler should not be called")
	}
}

func signDriverToken(t *testing.T, status int) string {
	t.Helper()
	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     25,
		AccountType:   "driver",
		AccountStatus: status,
		TTL:           time.Minute,
	}, "middleware-test-key")
	if err != nil {
		t.Fatal(err)
	}
	return token
}
