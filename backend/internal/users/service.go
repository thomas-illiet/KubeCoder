package users

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
)

// Store defines persistence required by the user service.
type Store interface {
	Upsert(context.Context, auth.Identity) (User, error)
	List(context.Context, string, string, int, int, string, string) ([]User, int64, error)
	UpdateRole(context.Context, uuid.UUID, bool) (User, error)
	Delete(context.Context, uuid.UUID) error
}

// Service implements application-level user operations.
type Service struct{ store Store }

// NewService creates a user service backed by the supplied store.
func NewService(store Store) *Service { return &Service{store: store} }

// Current provisions or refreshes the user represented by an OIDC identity.
func (s *Service) Current(ctx context.Context, identity auth.Identity) (User, error) {
	return s.store.Upsert(ctx, identity)
}

// List returns provisioned users for a platform administrator.
func (s *Service) List(ctx context.Context, actor User, query, role string, limit, offset int, orderBy, orderDirection string) ([]User, int64, error) {
	if !actor.IsAdmin {
		return nil, 0, ErrForbidden
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		role = "all"
	}
	if role != "all" && role != "admin" && role != "user" {
		return nil, 0, ErrInvalid
	}
	orderBy = strings.ToLower(strings.TrimSpace(orderBy))
	if orderBy == "" {
		orderBy = "display_name"
	}
	orderDirection = strings.ToLower(strings.TrimSpace(orderDirection))
	if orderDirection == "" {
		orderDirection = "asc"
	}
	allowedOrder := map[string]bool{"display_name": true, "username": true, "email": true, "is_admin": true, "created_at": true, "updated_at": true}
	if !allowedOrder[orderBy] || (orderDirection != "asc" && orderDirection != "desc") {
		return nil, 0, ErrInvalid
	}
	return s.store.List(ctx, strings.TrimSpace(query), role, limit, offset, orderBy, orderDirection)
}

// UpdateRole changes a provisioned user's platform administrator access.
func (s *Service) UpdateRole(ctx context.Context, actor User, targetID uuid.UUID, isAdmin bool) (User, error) {
	if !actor.IsAdmin {
		return User{}, ErrForbidden
	}
	if actor.ID == targetID && !isAdmin {
		return User{}, ErrConflict
	}
	return s.store.UpdateRole(ctx, targetID, isAdmin)
}

// Delete permanently removes a provisioned user and their memberships.
func (s *Service) Delete(ctx context.Context, actor User, targetID uuid.UUID) error {
	if !actor.IsAdmin {
		return ErrForbidden
	}
	if actor.ID == targetID {
		return ErrConflict
	}
	return s.store.Delete(ctx, targetID)
}
