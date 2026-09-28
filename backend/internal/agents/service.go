package agents

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

type Service struct{ repository *Repository }

// NewService creates the agent domain service.
func NewService(repository *Repository) *Service { return &Service{repository: repository} }

// List returns full agent definitions to platform administrators.
func (s *Service) List(ctx context.Context, actor users.User, query string, active *bool, limit, offset int) ([]Agent, int64, error) {
	if !actor.IsAdmin {
		return nil, 0, ErrForbidden
	}
	return s.repository.List(ctx, strings.TrimSpace(query), active, normalizeLimit(limit), max(offset, 0))
}

// Catalog returns active agents through their non-sensitive projection.
func (s *Service) Catalog(ctx context.Context, query string, limit, offset int) ([]PublicAgent, int64, error) {
	active := true
	items, total, err := s.repository.List(ctx, strings.TrimSpace(query), &active, normalizeLimit(limit), max(offset, 0))
	if err != nil {
		return nil, 0, err
	}
	result := make([]PublicAgent, 0, len(items))
	for _, item := range items {
		result = append(result, PublicAgent{ID: item.ID.String(), Name: item.Name, Description: item.Description, Capabilities: item.Capabilities, Active: item.Active})
	}
	return result, total, nil
}

// Get returns one full agent definition to a platform administrator.
func (s *Service) Get(ctx context.Context, actor users.User, id uuid.UUID) (Agent, error) {
	if !actor.IsAdmin {
		return Agent{}, ErrForbidden
	}
	return s.repository.Get(ctx, id)
}

// Create validates and persists an agent definition.
func (s *Service) Create(ctx context.Context, actor users.User, input Input) (Agent, error) {
	if !actor.IsAdmin {
		return Agent{}, ErrForbidden
	}
	item, err := validated(input)
	if err != nil {
		return Agent{}, err
	}
	if err := s.repository.Create(ctx, &item); err != nil {
		return Agent{}, err
	}
	return item, nil
}

// Update validates and replaces an agent definition.
func (s *Service) Update(ctx context.Context, actor users.User, id uuid.UUID, input Input) (Agent, error) {
	if !actor.IsAdmin {
		return Agent{}, ErrForbidden
	}
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return Agent{}, err
	}
	validated, err := validated(input)
	if err != nil {
		return Agent{}, err
	}
	validated.ID, validated.CreatedAt = current.ID, current.CreatedAt
	if err := s.repository.Update(ctx, &validated); err != nil {
		return Agent{}, err
	}
	return validated, nil
}

// Delete removes an unreferenced agent definition.
func (s *Service) Delete(ctx context.Context, actor users.User, id uuid.UUID) error {
	if !actor.IsAdmin {
		return ErrForbidden
	}
	return s.repository.Delete(ctx, id)
}

// validated trims, validates, and maps an agent input.
func validated(input Input) (Agent, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.RuntimeAdapter = strings.TrimSpace(input.RuntimeAdapter)
	input.RuntimeVersion = strings.TrimSpace(input.RuntimeVersion)
	input.Image = strings.TrimSpace(input.Image)
	input.Provider = strings.TrimSpace(input.Provider)
	input.Model = strings.TrimSpace(input.Model)
	input.SystemPrompt = strings.TrimSpace(input.SystemPrompt)
	if input.Name == "" || len(input.Name) > 120 || input.Description == "" || len(input.Description) > 1000 || input.RuntimeAdapter == "" || input.RuntimeVersion == "" || input.Image == "" || input.Provider == "" || input.Model == "" || input.SystemPrompt == "" || input.CPUMillis <= 0 || input.MemoryMB <= 0 || input.StorageMB <= 0 || input.MaxDurationSeconds <= 0 {
		return Agent{}, ErrInvalid
	}
	capabilities := make([]string, 0, len(input.Capabilities))
	seen := map[string]bool{}
	for _, value := range input.Capabilities {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			capabilities = append(capabilities, value)
		}
	}
	return Agent{Name: input.Name, Description: input.Description, RuntimeAdapter: input.RuntimeAdapter, RuntimeVersion: input.RuntimeVersion, Image: input.Image, Provider: input.Provider, Model: input.Model, SystemPrompt: input.SystemPrompt, Capabilities: capabilities, CPUMillis: input.CPUMillis, MemoryMB: input.MemoryMB, StorageMB: input.StorageMB, MaxDurationSecs: input.MaxDurationSeconds, Active: input.Active}, nil
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
