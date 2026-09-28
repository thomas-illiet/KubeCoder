package repositories

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	domain "github.com/thomas-illiet/KubeCoder/backend/internal/repositories"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

type fakeVerifier struct{}

// Verify returns a deterministic identity.
func (fakeVerifier) Verify(context.Context, string) (auth.Identity, error) {
	return auth.Identity{Issuer: "issuer", Subject: "subject"}, nil
}

type fakeUsers struct{}

// Current returns a deterministic user.
func (fakeUsers) Current(context.Context, auth.Identity) (users.User, error) {
	return users.User{ID: uuid.New()}, nil
}

type fakeRepositories struct{ err error }

// List returns the configured repository error.
func (f fakeRepositories) List(context.Context, users.User, string, string, string, *bool, *uuid.UUID, int, int) ([]domain.Item, int64, error) {
	return nil, 0, f.err
}

// Get returns the configured repository error.
func (f fakeRepositories) Get(context.Context, users.User, string, uuid.UUID) (domain.Item, error) {
	return domain.Item{}, f.err
}

// Create returns the configured repository error.
func (f fakeRepositories) Create(context.Context, users.User, string, domain.Input) (domain.Item, error) {
	return domain.Item{}, f.err
}

// Update returns the configured repository error.
func (f fakeRepositories) Update(context.Context, users.User, string, uuid.UUID, domain.Input) (domain.Item, error) {
	return domain.Item{}, f.err
}

// Delete returns the configured repository error.
func (f fakeRepositories) Delete(context.Context, users.User, string, uuid.UUID) error { return f.err }

// TestCrossOrganizationRepositoryIsForbidden verifies membership errors remain forbidden.
func TestCrossOrganizationRepositoryIsForbidden(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewHandler(logger, fakeVerifier{}, fakeUsers{}, fakeRepositories{err: domain.ErrForbidden}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/private/repositories", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	httpx.Middleware(logger, mux).ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
