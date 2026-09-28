package users

import (
	"context"
	"errors"
	"strings"

	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
)

// Store defines persistence required by the user service.
type Store interface {
	Upsert(context.Context, auth.Identity) (User, error)
	List(context.Context, string, int, int) ([]User, int64, error)
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
func (s *Service) List(ctx context.Context, actor User, query string, limit, offset int) ([]User, int64, error) {
	if !actor.IsAdmin {
		return nil, 0, errors.New("user administration forbidden")
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
	return s.store.List(ctx, strings.TrimSpace(query), limit, offset)
}
