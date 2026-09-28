package secrets

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testService constructs an isolated in-memory domain fixture.
func testService(t *testing.T) (*Service, *gorm.DB, users.User) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE users (id text PRIMARY KEY, issuer text, subject text, username text, display_name text, email text, is_admin boolean, preferred_organization_id text, created_at datetime, updated_at datetime)`,
		`CREATE TABLE organizations (id text PRIMARY KEY, name text, slug text, created_at datetime, updated_at datetime)`,
		`CREATE TABLE organization_memberships (organization_id text, user_id text, created_at datetime, PRIMARY KEY (organization_id, user_id))`,
		`CREATE TABLE agents (id text PRIMARY KEY, name text, active boolean)`,
		`CREATE TABLE repositories (id text PRIMARY KEY, organization_id text, name text)`,
		`CREATE TABLE secrets (id text PRIMARY KEY, scope text, organization_id text, variable_name text UNIQUE, description text, encrypted_value blob, nonce blob, fingerprint text, encryption_version text, expires_at datetime, value_replaced_at datetime, created_at datetime, updated_at datetime)`,
		`CREATE TABLE secret_bindings (id text PRIMARY KEY, secret_id text, target_type text, agent_id text, repository_id text, created_at datetime)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	actor := users.User{ID: uuid.New(), Issuer: "test", Subject: uuid.NewString(), Username: "admin", DisplayName: "Admin", Email: "admin@example.test", IsAdmin: true}
	if err := db.Create(&actor).Error; err != nil {
		t.Fatal(err)
	}
	service, err := NewService(NewRepository(db), make([]byte, 32), 30*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return service, db, actor
}

// TestReplacePermanentlyOverwrites proves the no-versioning contract.
func TestReplacePermanentlyOverwrites(t *testing.T) {
	service, db, actor := testService(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	created, err := service.CreatePlatform(ctx, actor, Input{Scope: ScopePlatform, VariableName: "MODEL_TOKEN", Description: "provider", Value: "old-value"})
	if err != nil {
		t.Fatal(err)
	}
	var before models.Secret
	if err := db.First(&before, "id = ?", created.ID).Error; err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Hour)
	replaced, err := service.Replace(ctx, actor, created.ID, nil, ReplaceInput{Value: "new-value"})
	if err != nil {
		t.Fatal(err)
	}
	var after models.Secret
	if err := db.First(&after, "id = ?", created.ID).Error; err != nil {
		t.Fatal(err)
	}
	if string(before.EncryptedValue) == string(after.EncryptedValue) {
		t.Fatal("ciphertext was not replaced")
	}
	plaintext, err := service.protector.open(after)
	if err != nil {
		t.Fatal(err)
	}
	if string(plaintext) != "new-value" {
		t.Fatalf("got plaintext %q", plaintext)
	}
	if replaced.Status != StatusActive {
		t.Fatal("replacement did not remain active")
	}
	if db.Migrator().HasTable("secret_versions") {
		t.Fatal("unexpected secret_versions table")
	}
}

// TestStatusTransitions covers computed lifecycle states.
func TestStatusTransitions(t *testing.T) {
	service, _, _ := testService(t)
	now := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	expired := now.Add(-time.Minute)
	soon := now.Add(24 * time.Hour)
	later := now.Add(60 * 24 * time.Hour)
	for name, item := range map[string]struct {
		item Secret
		want string
	}{"active": {Secret{}, StatusActive}, "later": {Secret{ExpiresAt: &later}, StatusActive}, "soon": {Secret{ExpiresAt: &soon}, StatusExpiringSoon}, "expired": {Secret{ExpiresAt: &expired}, StatusExpired}} {
		t.Run(name, func(t *testing.T) {
			if got := service.status(item.item); got != item.want {
				t.Fatalf("got %s want %s", got, item.want)
			}
		})
	}
}

// TestMemberCannotMutate reserves writes for platform administrators.
func TestMemberCannotMutate(t *testing.T) {
	service, _, _ := testService(t)
	member := users.User{ID: uuid.New()}
	if _, err := service.CreatePlatform(context.Background(), member, Input{}); err != ErrForbidden {
		t.Fatalf("got %v", err)
	}
}

// TestCreateNormalizesVariableName keeps the API contract aligned with the UI.
func TestCreateNormalizesVariableName(t *testing.T) {
	service, _, actor := testService(t)
	created, err := service.CreatePlatform(context.Background(), actor, Input{Scope: ScopePlatform, VariableName: "model api token", Value: "protected"})
	if err != nil {
		t.Fatal(err)
	}
	if created.VariableName != "MODEL_API_TOKEN" {
		t.Fatalf("got %q", created.VariableName)
	}
}

// TestNormalizeSortRejectsUnknownValues verifies the SQL ordering whitelist.
func TestNormalizeSortRejectsUnknownValues(t *testing.T) {
	if field, order := normalizeSort("unknown", SortAscending); field != "" || order != "" {
		t.Fatalf("unexpected sort %q %q", field, order)
	}
	if field, order := normalizeSort("expires_at", SortDescending); field != "expires_at" || order != SortDescending {
		t.Fatalf("unexpected sort %q %q", field, order)
	}
}

// TestPlatformVariableCannotBeOverriddenByTenant protects global precedence.
func TestPlatformVariableCannotBeOverriddenByTenant(t *testing.T) {
	service, _, actor := testService(t)
	ctx := context.Background()
	_, err := service.CreatePlatform(ctx, actor, Input{Scope: ScopePlatform, VariableName: "GLOBAL_TOKEN", Value: "protected"})
	if err != nil {
		t.Fatal(err)
	}
	organizationID := uuid.New()
	if _, err := service.CreateOrganization(ctx, actor, organizationID, Input{Scope: ScopeOrganization, VariableName: "GLOBAL_TOKEN", Value: "tenant"}); err != ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
}

// TestPlatformVariableCannotOverrideTenant enforces precedence regardless of creation order.
func TestPlatformVariableCannotOverrideTenant(t *testing.T) {
	service, _, actor := testService(t)
	ctx := context.Background()
	organizationID := uuid.New()
	if _, err := service.CreateOrganization(ctx, actor, organizationID, Input{Scope: ScopeOrganization, VariableName: "SHARED_TOKEN", Value: "tenant"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreatePlatform(ctx, actor, Input{Scope: ScopePlatform, VariableName: "SHARED_TOKEN", Value: "platform"}); err != ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
}

// TestOrganizationMemberCanManageTenantSecret verifies member writes stay tenant-scoped.
func TestOrganizationMemberCanManageTenantSecret(t *testing.T) {
	service, db, actor := testService(t)
	actor.IsAdmin = false
	organizationID := uuid.New()
	if err := db.Exec("INSERT INTO organization_memberships (organization_id, user_id, created_at) VALUES (?, ?, ?)", organizationID, actor.ID, time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateOrganization(context.Background(), actor, organizationID, Input{Scope: ScopeOrganization, VariableName: "TENANT_TOKEN", Value: "protected"})
	if err != nil {
		t.Fatal(err)
	}
	if created.Scope != ScopeOrganization || created.OrganizationID == nil || *created.OrganizationID != organizationID {
		t.Fatalf("unexpected tenant projection: %#v", created)
	}
}

// TestOrganizationMutationCannotReachPlatformSecret protects the locked global scope.
func TestOrganizationMutationCannotReachPlatformSecret(t *testing.T) {
	service, db, actor := testService(t)
	ctx := context.Background()
	created, err := service.CreatePlatform(ctx, actor, Input{Scope: ScopePlatform, VariableName: "LOCKED_TOKEN", Value: "platform"})
	if err != nil {
		t.Fatal(err)
	}
	actor.IsAdmin = false
	organizationID := uuid.New()
	if err := db.Exec("INSERT INTO organization_memberships (organization_id, user_id, created_at) VALUES (?, ?, ?)", organizationID, actor.ID, time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.Replace(ctx, actor, created.ID, &organizationID, ReplaceInput{Value: "tenant"}); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

// TestRepositoryBindingRejectsAnotherOrganization enforces tenant isolation on targets.
func TestRepositoryBindingRejectsAnotherOrganization(t *testing.T) {
	service, db, actor := testService(t)
	ctx := context.Background()
	organizationID, otherOrganizationID := uuid.New(), uuid.New()
	secret, err := service.CreateOrganization(ctx, actor, organizationID, Input{Scope: ScopeOrganization, VariableName: "TENANT_TOKEN", Value: "protected"})
	if err != nil {
		t.Fatal(err)
	}
	repositoryID := uuid.New()
	if err := db.Exec("INSERT INTO repositories (id, organization_id, name) VALUES (?, ?, ?)", repositoryID, otherOrganizationID, "other").Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddBinding(ctx, actor, secret.ID, &organizationID, BindingInput{TargetType: TargetRepository, TargetID: repositoryID}); err != ErrInvalid {
		t.Fatalf("expected invalid target, got %v", err)
	}
}

// TestRepositoryScopeIsRejected keeps repositories as binding targets only.
func TestRepositoryScopeIsRejected(t *testing.T) {
	service, _, actor := testService(t)
	_, err := service.CreateOrganization(context.Background(), actor, uuid.New(), Input{Scope: "REPOSITORY", VariableName: "REPOSITORY_TOKEN", Value: "protected"})
	if err != ErrInvalid {
		t.Fatalf("expected invalid scope, got %v", err)
	}
}
