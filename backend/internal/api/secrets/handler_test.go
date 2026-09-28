package secrets

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	"github.com/thomas-illiet/KubeCoder/backend/internal/organizations"
	domain "github.com/thomas-illiet/KubeCoder/backend/internal/secrets"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

type testVerifier struct{}

// Verify returns a deterministic authenticated identity.
func (testVerifier) Verify(context.Context, string) (auth.Identity, error) {
	return auth.Identity{Issuer: "https://issuer.example", Subject: "subject"}, nil
}

type testUsers struct{ actor users.User }

// Current returns the configured actor.
func (f testUsers) Current(context.Context, auth.Identity) (users.User, error) { return f.actor, nil }

type testOrganizations struct {
	organization organizations.Organization
	err          error
}

// Get returns the configured administrator organization.
func (f testOrganizations) Get(context.Context, users.User, uuid.UUID) (organizations.Organization, error) {
	return f.organization, f.err
}

// GetForUser returns the configured member organization.
func (f testOrganizations) GetForUser(context.Context, users.User, string) (organizations.Organization, error) {
	return f.organization, f.err
}

type testSecretService struct {
	items      []domain.View
	err        error
	replaceOID *uuid.UUID
}

// ListPlatform returns the configured sanitized page.
func (f *testSecretService) ListPlatform(context.Context, users.User, string, string, string, string, int, int) ([]domain.View, int64, error) {
	return f.items, int64(len(f.items)), f.err
}

// ListOrganizationAdmin returns the configured sanitized page.
func (f *testSecretService) ListOrganizationAdmin(context.Context, users.User, uuid.UUID, string, string, string, string, string, int, int) ([]domain.View, int64, error) {
	return f.items, int64(len(f.items)), f.err
}

// ListOrganization returns the configured sanitized page.
func (f *testSecretService) ListOrganization(context.Context, uuid.UUID, string, string, string, string, string, int, int) ([]domain.View, int64, error) {
	return f.items, int64(len(f.items)), f.err
}

// CreatePlatform returns the configured result.
func (f *testSecretService) CreatePlatform(context.Context, users.User, domain.Input) (domain.View, error) {
	return domain.View{}, f.err
}

// CreateOrganization returns the configured result.
func (f *testSecretService) CreateOrganization(context.Context, users.User, uuid.UUID, domain.Input) (domain.View, error) {
	return domain.View{}, f.err
}

// Replace records the ownership boundary supplied by the handler.
func (f *testSecretService) Replace(_ context.Context, _ users.User, _ uuid.UUID, oid *uuid.UUID, _ domain.ReplaceInput) (domain.View, error) {
	f.replaceOID = oid
	return domain.View{}, f.err
}

// AddBinding returns the configured result.
func (f *testSecretService) AddBinding(context.Context, users.User, uuid.UUID, *uuid.UUID, domain.BindingInput) (domain.BindingView, error) {
	return domain.BindingView{}, f.err
}

// RemoveBinding returns the configured result.
func (f *testSecretService) RemoveBinding(context.Context, users.User, uuid.UUID, uuid.UUID, *uuid.UUID) error {
	return f.err
}

// Targets returns the configured result.
func (f *testSecretService) Targets(context.Context, users.User, *uuid.UUID) ([]domain.Target, error) {
	return nil, f.err
}

// secretTestHandler builds a fully routed handler with deterministic dependencies.
func secretTestHandler(organization organizations.Organization, service *testSecretService) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewHandler(logger, testVerifier{}, testUsers{actor: users.User{ID: uuid.New()}}, testOrganizations{organization: organization}, service))
	return httpx.Middleware(logger, mux)
}

// TestOrganizationListNeverSerializesProtectedStorage checks the member projection.
func TestOrganizationListNeverSerializesProtectedStorage(t *testing.T) {
	t.Parallel()
	organization := organizations.Organization{ID: uuid.New(), Slug: "northstar"}
	service := &testSecretService{items: []domain.View{{ID: uuid.New(), Scope: domain.ScopeOrganization, OrganizationID: &organization.ID, VariableName: "TOKEN", Bindings: []domain.BindingView{}, CreatedAt: time.Now(), UpdatedAt: time.Now(), ValueReplacedAt: time.Now()}}}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/northstar/secrets", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	secretTestHandler(organization, service).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	for _, forbidden := range []string{"encrypted_value", "ciphertext", "nonce", "fingerprint"} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("response exposes %q: %s", forbidden, response.Body.String())
		}
	}
}

// TestOrganizationMutationCarriesMembershipBoundary checks tenant routing.
func TestOrganizationMutationCarriesMembershipBoundary(t *testing.T) {
	t.Parallel()
	organization := organizations.Organization{ID: uuid.New(), Slug: "northstar"}
	service := &testSecretService{}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/organizations/northstar/secrets/"+uuid.NewString()+"/replace", strings.NewReader(`{"value":"replacement"}`))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	secretTestHandler(organization, service).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if service.replaceOID == nil || *service.replaceOID != organization.ID {
		t.Fatalf("organization boundary = %v, want %s", service.replaceOID, organization.ID)
	}
}

// TestInternalErrorDoesNotExposeSensitiveDetails checks Problem JSON redaction.
func TestInternalErrorDoesNotExposeSensitiveDetails(t *testing.T) {
	t.Parallel()
	service := &testSecretService{err: errors.New("ciphertext secret-nonce fingerprint")}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/secrets", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	secretTestHandler(organizations.Organization{}, service).ServeHTTP(response, request)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	for _, forbidden := range []string{"ciphertext", "secret-nonce", "fingerprint"} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("problem response exposes %q: %s", forbidden, response.Body.String())
		}
	}
}
