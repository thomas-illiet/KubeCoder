package organizations

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
	domain "github.com/thomas-illiet/KubeCoder/backend/internal/organizations"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

type fakeVerifier struct{}

// Verify returns a deterministic authenticated identity.
func (fakeVerifier) Verify(context.Context, string) (auth.Identity, error) {
	return auth.Identity{Issuer: "https://issuer.example", Subject: "subject"}, nil
}

type fakeUsers struct{ current users.User }

// Current returns the configured local user.
func (f fakeUsers) Current(context.Context, auth.Identity) (users.User, error) { return f.current, nil }

type fakeOrganizations struct {
	preferred domain.Organization
	err       error
}

// ListForUser returns the configured organization page.
func (f fakeOrganizations) ListForUser(context.Context, users.User, string, int, int) ([]domain.Organization, int64, error) {
	return []domain.Organization{f.preferred}, 1, f.err
}

// GetForUser returns the configured organization.
func (f fakeOrganizations) GetForUser(context.Context, users.User, string) (domain.Organization, error) {
	return f.preferred, f.err
}

// SetPreferred returns the configured preferred organization.
func (f fakeOrganizations) SetPreferred(context.Context, users.User, string) (domain.Organization, error) {
	return f.preferred, f.err
}

// testHandler builds an organization handler with deterministic dependencies.
func testHandler(service Service) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewHandler(logger, fakeVerifier{}, fakeUsers{current: users.User{ID: uuid.New()}}, service))
	return httpx.Middleware(logger, mux)
}

// TestSetPreferredOrganization verifies the bodyless preference endpoint.
func TestSetPreferredOrganization(t *testing.T) {
	t.Parallel()
	expected := domain.Organization{ID: uuid.New(), Name: "Northstar Labs", Slug: "northstar-labs"}
	request := httptest.NewRequest(http.MethodPut, "/api/v1/organizations/northstar-labs/preferred", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	testHandler(fakeOrganizations{preferred: expected}).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

// TestOrganizationMembershipIsRequired verifies that inaccessible slugs return forbidden.
func TestOrganizationMembershipIsRequired(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/private", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	testHandler(fakeOrganizations{err: domain.ErrForbidden}).ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
