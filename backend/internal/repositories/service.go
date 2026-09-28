package repositories

import (
	"context"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/organizations"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

type Service struct {
	repository    *RepositoryStore
	organizations *organizations.Repository
}

// NewService creates the organization repository domain service.
func NewService(repository *RepositoryStore, organizations *organizations.Repository) *Service {
	return &Service{repository: repository, organizations: organizations}
}

// List returns a filtered page after membership validation.
func (s *Service) List(ctx context.Context, actor users.User, slug, query, provider string, configured *bool, agentID *uuid.UUID, limit, offset int) ([]Item, int64, error) {
	organization, err := s.organization(ctx, actor, slug)
	if err != nil {
		return nil, 0, err
	}
	if provider != "" && !validProvider(provider) {
		return nil, 0, ErrInvalid
	}
	return s.repository.List(ctx, organization.ID, strings.TrimSpace(query), provider, configured, agentID, normalizeLimit(limit), max(offset, 0))
}

// Get returns one organization repository after membership validation.
func (s *Service) Get(ctx context.Context, actor users.User, slug string, id uuid.UUID) (Item, error) {
	organization, err := s.organization(ctx, actor, slug)
	if err != nil {
		return Item{}, err
	}
	return s.repository.Get(ctx, organization.ID, id)
}

// Create validates and creates an organization repository.
func (s *Service) Create(ctx context.Context, actor users.User, slug string, input Input) (Item, error) {
	organization, err := s.organization(ctx, actor, slug)
	if err != nil {
		return Item{}, err
	}
	row, agentID, err := s.validated(ctx, organization.ID, uuid.Nil, input)
	if err != nil {
		return Item{}, err
	}
	if err := s.repository.Create(ctx, &row, agentID); err != nil {
		return Item{}, err
	}
	return s.repository.Get(ctx, organization.ID, row.ID)
}

// Update validates and updates an organization repository.
func (s *Service) Update(ctx context.Context, actor users.User, slug string, id uuid.UUID, input Input) (Item, error) {
	organization, err := s.organization(ctx, actor, slug)
	if err != nil {
		return Item{}, err
	}
	row, agentID, err := s.validated(ctx, organization.ID, id, input)
	if err != nil {
		return Item{}, err
	}
	if err := s.repository.Update(ctx, &row, agentID); err != nil {
		return Item{}, err
	}
	return s.repository.Get(ctx, organization.ID, id)
}

// Delete permanently removes an organization repository.
func (s *Service) Delete(ctx context.Context, actor users.User, slug string, id uuid.UUID) error {
	organization, err := s.organization(ctx, actor, slug)
	if err != nil {
		return err
	}
	return s.repository.Delete(ctx, organization.ID, id)
}

// organization resolves a tenant only through the actor membership.
func (s *Service) organization(ctx context.Context, actor users.User, slug string) (organizations.Organization, error) {
	organization, err := s.organizations.FindForUserBySlug(ctx, actor.ID, slug)
	if err == organizations.ErrForbidden {
		return organizations.Organization{}, ErrForbidden
	}
	return organization, err
}

// validated normalizes repository input and verifies the optional agent.
func (s *Service) validated(ctx context.Context, organizationID, id uuid.UUID, input Input) (Repository, *uuid.UUID, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	input.DefaultBranch = strings.TrimSpace(input.DefaultBranch)
	normalizedURL, err := normalizeCloneURL(input.CloneURL)
	if err != nil || input.Name == "" || len(input.Name) > 120 || !validProvider(input.Provider) || input.DefaultBranch == "" || len(input.DefaultBranch) > 255 {
		return Repository{}, nil, ErrInvalid
	}
	if input.AgentID != nil {
		available, err := s.repository.AgentAvailable(ctx, *input.AgentID)
		if err != nil {
			return Repository{}, nil, err
		}
		if !available {
			return Repository{}, nil, ErrAgent
		}
	}
	return Repository{ID: id, OrganizationID: organizationID, Name: input.Name, Provider: input.Provider, CloneURL: normalizedURL, DefaultBranch: input.DefaultBranch, IncludeSubmodules: input.IncludeSubmodules}, input.AgentID, nil
}

// normalizeCloneURL validates and canonicalizes supported Git URLs.
func normalizeCloneURL(value string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "ssh") || parsed.Host == "" || strings.Trim(parsed.Path, "/") == "" || parsed.Scheme == "https" && parsed.User != nil || parsed.User != nil && (parsed.User.Username() != "git" || parsed.User.String() != "git") {
		return "", ErrInvalid
	}
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimSuffix(strings.TrimSuffix(parsed.Path, "/"), ".git")
	parsed.RawQuery, parsed.Fragment = "", ""
	return parsed.String(), nil
}

// validProvider reports whether a provider is supported by the UI contract.
func validProvider(value string) bool {
	return value == "github" || value == "gitlab" || value == "bitbucket"
}

// normalizeLimit applies the public pagination bounds.
func normalizeLimit(value int) int {
	if value <= 0 {
		return 20
	}
	if value > 100 {
		return 100
	}
	return value
}
