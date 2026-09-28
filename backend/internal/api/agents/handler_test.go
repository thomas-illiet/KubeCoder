package agents

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	domain "github.com/thomas-illiet/KubeCoder/backend/internal/agents"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	"github.com/thomas-illiet/KubeCoder/backend/internal/organizations"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

type fakeVerifier struct{}

// Verify returns a deterministic identity.
func (fakeVerifier) Verify(context.Context, string) (auth.Identity, error) {
	return auth.Identity{Issuer: "issuer", Subject: "subject"}, nil
}

type fakeUsers struct{ actor users.User }

// Current returns the configured user.
func (f fakeUsers) Current(context.Context, auth.Identity) (users.User, error) { return f.actor, nil }

type fakeOrganizations struct{ err error }

// GetForUser validates the configured membership result.
func (f fakeOrganizations) GetForUser(context.Context, users.User, string) (organizations.Organization, error) {
	return organizations.Organization{}, f.err
}

type fakeAgents struct{ catalog []domain.PublicAgent }

// List returns no administrative agents.
func (f fakeAgents) List(context.Context, users.User, string, *bool, int, int) ([]domain.Agent, int64, error) {
	return nil, 0, nil
}

// Catalog returns the configured public projection.
func (f fakeAgents) Catalog(context.Context, string, int, int) ([]domain.PublicAgent, int64, error) {
	return f.catalog, int64(len(f.catalog)), nil
}

// Get returns an empty agent.
func (f fakeAgents) Get(context.Context, users.User, uuid.UUID) (domain.Agent, error) {
	return domain.Agent{}, nil
}

// Create returns an empty agent.
func (f fakeAgents) Create(context.Context, users.User, domain.Input) (domain.Agent, error) {
	return domain.Agent{}, nil
}

// Update returns an empty agent.
func (f fakeAgents) Update(context.Context, users.User, uuid.UUID, domain.Input) (domain.Agent, error) {
	return domain.Agent{}, nil
}

// Delete succeeds for the fake service.
func (f fakeAgents) Delete(context.Context, users.User, uuid.UUID) error { return nil }

// TestCatalogRequiresMembership verifies tenant authorization before projection.
func TestCatalogRequiresMembership(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewHandler(logger, fakeVerifier{}, fakeUsers{}, fakeOrganizations{err: organizations.ErrForbidden}, fakeAgents{}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/private/agents", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	httpx.Middleware(logger, mux).ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
