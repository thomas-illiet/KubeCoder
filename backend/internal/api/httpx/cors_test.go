package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCORSPreflightForConfiguredFrontend verifies the configured browser origin.
func TestCORSPreflightForConfiguredFrontend(t *testing.T) {
	t.Parallel()
	handler := CORS([]string{"http://localhost:5173"}, http.NotFoundHandler())
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/users/me", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("allow origin = %q", got)
	}
}
