package organizations

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	domain "github.com/thomas-illiet/KubeCoder/backend/internal/organizations"
	"github.com/thomas-illiet/KubeCoder/backend/internal/sshkeys"
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

type fakeSSHKeys struct {
	key models.OrganizationSSHKey
	err error
}

// Get returns the configured SSH key response.
func (f fakeSSHKeys) Get(context.Context, uuid.UUID) (models.OrganizationSSHKey, error) {
	return f.key, f.err
}

// Regenerate returns the configured SSH key rotation response.
func (f fakeSSHKeys) Regenerate(context.Context, uuid.UUID) (models.OrganizationSSHKey, error) {
	return f.key, f.err
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
	RegisterRoutes(mux, NewHandler(logger, fakeVerifier{}, fakeUsers{current: users.User{ID: uuid.New()}}, service, fakeSSHKeys{}))
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

// TestGetSSHKeyReturnsOnlyPublicMetadata verifies protected columns never reach the response.
func TestGetSSHKeyReturnsOnlyPublicMetadata(t *testing.T) {
	t.Parallel()
	organization := domain.Organization{ID: uuid.New(), Slug: "northstar-labs"}
	key := models.OrganizationSSHKey{
		OrganizationID: organization.ID, PublicKey: "ssh-ed25519 public", Fingerprint: "SHA256:test", KeyAlgorithm: "ssh-ed25519",
		EncryptedPrivateKey: []byte("private-ciphertext"), Nonce: []byte("secret-nonce"), EncryptionVersion: "aes-256-gcm-v1",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewHandler(logger, fakeVerifier{}, fakeUsers{current: users.User{ID: uuid.New()}}, fakeOrganizations{preferred: organization}, fakeSSHKeys{key: key}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/northstar-labs/ssh-key", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	httpx.Middleware(logger, mux).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	for _, forbidden := range []string{"private-ciphertext", "secret-nonce", "encryption_version", "organization_id"} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("response leaks %q: %s", forbidden, response.Body.String())
		}
	}
}

// TestGetSSHKeyNotFound verifies legacy organizations receive an actionable empty state.
func TestGetSSHKeyNotFound(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewHandler(logger, fakeVerifier{}, fakeUsers{current: users.User{ID: uuid.New()}}, fakeOrganizations{preferred: domain.Organization{ID: uuid.New()}}, fakeSSHKeys{err: sshkeys.ErrNotFound}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/legacy/ssh-key", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	httpx.Middleware(logger, mux).ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

// TestRegenerateSSHKeyNotFound verifies rotation never creates a missing initial key.
func TestRegenerateSSHKeyNotFound(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewHandler(logger, fakeVerifier{}, fakeUsers{current: users.User{ID: uuid.New()}}, fakeOrganizations{preferred: domain.Organization{ID: uuid.New()}}, fakeSSHKeys{err: sshkeys.ErrNotFound}))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/organizations/current/ssh-key/regenerate", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	httpx.Middleware(logger, mux).ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

// TestSSHKeyRequiresAuthentication verifies anonymous key reads are rejected.
func TestSSHKeyRequiresAuthentication(t *testing.T) {
	t.Parallel()
	response := httptest.NewRecorder()
	testHandler(fakeOrganizations{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/organizations/northstar/ssh-key", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

// TestSSHKeyRequiresOrganizationMembership verifies one tenant cannot read another tenant's key.
func TestSSHKeyRequiresOrganizationMembership(t *testing.T) {
	t.Parallel()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/private/ssh-key", nil)
	request.Header.Set("Authorization", "Bearer token")
	testHandler(fakeOrganizations{err: domain.ErrForbidden}).ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
