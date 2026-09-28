package users

import (
	"context"

	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
)

// Store defines persistence required by the user service.
type Store interface {
	Upsert(context.Context, auth.Identity) (User, error)
}

// Service implements application-level user operations.
type Service struct{ store Store }

// NewService creates a user service backed by the supplied store.
func NewService(store Store) *Service { return &Service{store: store} }

// Current provisions or refreshes the user represented by an OIDC identity.
func (s *Service) Current(ctx context.Context, identity auth.Identity) (User, error) {
	return s.store.Upsert(ctx, identity)
}
