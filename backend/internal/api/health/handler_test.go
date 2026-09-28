package health

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeDatabase struct{ err error }

// PingContext returns the readiness error configured by the test.
func (d fakeDatabase) PingContext(context.Context) error { return d.err }

// TestReadinessFailure verifies that an unavailable database blocks readiness.
func TestReadinessFailure(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewHandler(fakeDatabase{err: sql.ErrConnDone}, func(context.Context) error { return nil }))
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", response.Code)
	}
}
