package users

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
)

type serviceStore struct {
	listedRole      string
	listedOrderBy   string
	listedDirection string
	updatedID       uuid.UUID
	updatedRole     bool
	deletedID       uuid.UUID
}

// Upsert satisfies Store for service tests.
func (s *serviceStore) Upsert(context.Context, auth.Identity) (User, error) { return User{}, nil }

// List records validated list options.
func (s *serviceStore) List(_ context.Context, _ string, role string, _ int, _ int, orderBy, direction string) ([]User, int64, error) {
	s.listedRole, s.listedOrderBy, s.listedDirection = role, orderBy, direction
	return []User{}, 0, nil
}

// UpdateRole records the requested role change.
func (s *serviceStore) UpdateRole(_ context.Context, id uuid.UUID, isAdmin bool) (User, error) {
	s.updatedID, s.updatedRole = id, isAdmin
	return User{ID: id, IsAdmin: isAdmin}, nil
}

// Delete records the requested deletion.
func (s *serviceStore) Delete(_ context.Context, id uuid.UUID) error { s.deletedID = id; return nil }

// TestListValidatesServerSideRoleAndOrder verifies that SQL-facing values are allow-listed.
func TestListValidatesServerSideRoleAndOrder(t *testing.T) {
	t.Parallel()
	store := &serviceStore{}
	service := NewService(store)
	actor := User{ID: uuid.New(), IsAdmin: true}

	if _, _, err := service.List(context.Background(), actor, "alice", "admin", 20, 0, "created_at", "desc"); err != nil {
		t.Fatal(err)
	}
	if store.listedRole != "admin" || store.listedOrderBy != "created_at" || store.listedDirection != "desc" {
		t.Fatalf("unexpected list arguments: role=%q order=%q direction=%q", store.listedRole, store.listedOrderBy, store.listedDirection)
	}
	if _, _, err := service.List(context.Background(), actor, "", "owner", 20, 0, "display_name", "asc"); err != ErrInvalid {
		t.Fatalf("invalid role error = %v, want ErrInvalid", err)
	}
	if _, _, err := service.List(context.Background(), actor, "", "all", 20, 0, "issuer; drop table users", "asc"); err != ErrInvalid {
		t.Fatalf("invalid order error = %v, want ErrInvalid", err)
	}
}

// TestUserAdministrationProtectsCurrentAdministrator verifies self-demotion and self-deletion guards.
func TestUserAdministrationProtectsCurrentAdministrator(t *testing.T) {
	t.Parallel()
	store := &serviceStore{}
	service := NewService(store)
	actor := User{ID: uuid.New(), IsAdmin: true}

	if _, err := service.UpdateRole(context.Background(), actor, actor.ID, false); err != ErrConflict {
		t.Fatalf("self-demotion error = %v, want ErrConflict", err)
	}
	if err := service.Delete(context.Background(), actor, actor.ID); err != ErrConflict {
		t.Fatalf("self-delete error = %v, want ErrConflict", err)
	}
	if store.updatedID != uuid.Nil || store.deletedID != uuid.Nil {
		t.Fatal("protected operations reached the store")
	}
}

// TestUserAdministrationMutationsReachStore verifies authorized role and delete operations.
func TestUserAdministrationMutationsReachStore(t *testing.T) {
	t.Parallel()
	store := &serviceStore{}
	service := NewService(store)
	actor := User{ID: uuid.New(), IsAdmin: true}
	targetID := uuid.New()

	if _, err := service.UpdateRole(context.Background(), actor, targetID, true); err != nil {
		t.Fatal(err)
	}
	if store.updatedID != targetID || !store.updatedRole {
		t.Fatal("role update was not forwarded")
	}
	if err := service.Delete(context.Background(), actor, targetID); err != nil {
		t.Fatal(err)
	}
	if store.deletedID != targetID {
		t.Fatal("delete was not forwarded")
	}
}
