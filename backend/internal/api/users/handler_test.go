package users

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	domain "github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

type fakeVerifier struct{ identity auth.Identity }

// Verify returns the identity configured by the test.
func (v fakeVerifier) Verify(context.Context, string) (auth.Identity, error) { return v.identity, nil }

type fakeService struct{ current domain.User }

// Current returns the user configured by the test.
func (s fakeService) Current(context.Context, auth.Identity) (domain.User, error) {
	return s.current, nil
}

// testHandler builds a user handler with deterministic test doubles.
func testHandler(verifier auth.Verifier, service Service) http.Handler {
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), verifier, service))
	return httpx.Middleware(slog.New(slog.NewTextHandler(io.Discard, nil)), mux)
}

// TestCurrentUser verifies the authenticated user response contract.
func TestCurrentUser(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC().Truncate(time.Second)
	expected := domain.User{ID: uuid.New(), Subject: "subject", Username: "admin", DisplayName: "Demo Admin", Email: "admin@example.test", IsAdmin: true, CreatedAt: now, UpdatedAt: now}
	handler := testHandler(fakeVerifier{identity: auth.Identity{Issuer: "https://issuer.example", Subject: "subject"}}, fakeService{current: expected})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

// TestCurrentUserRequiresBearerToken verifies the unauthenticated problem response.
func TestCurrentUserRequiresBearerToken(t *testing.T) {
	t.Parallel()
	handler := testHandler(fakeVerifier{}, fakeService{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
