package organizations

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeSSHKeyGenerator struct{ err error }

// Generate returns deterministic protected key material for organization service tests.
func (f fakeSSHKeyGenerator) Generate(id uuid.UUID) (models.OrganizationSSHKey, error) {
	return models.OrganizationSSHKey{OrganizationID: id, PublicKey: "ssh-ed25519 public", EncryptedPrivateKey: []byte("ciphertext"), Nonce: []byte("nonce"), KeyAlgorithm: "ssh-ed25519", EncryptionVersion: "aes-256-gcm-v1", Fingerprint: "SHA256:test"}, f.err
}

// TestCreatePersistsOrganizationAndSSHKeyAtomically verifies both records are created together.
func TestCreatePersistsOrganizationAndSSHKeyAtomically(t *testing.T) {
	t.Parallel()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal(err)
	}
	createSSHKeyTestSchema(t, db)
	service := NewService(NewRepository(db), fakeSSHKeyGenerator{})
	organization, err := service.Create(context.Background(), users.User{IsAdmin: true}, "Northstar", "northstar")
	if err != nil {
		t.Fatal(err)
	}
	var key models.OrganizationSSHKey
	if err := db.First(&key, "organization_id = ?", organization.ID).Error; err != nil {
		t.Fatal(err)
	}
}

// TestCreateDoesNotPersistOrganizationWhenKeyGenerationFails verifies generation precedes persistence.
func TestCreateDoesNotPersistOrganizationWhenKeyGenerationFails(t *testing.T) {
	t.Parallel()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal(err)
	}
	createSSHKeyTestSchema(t, db)
	expected := errors.New("generation failed")
	service := NewService(NewRepository(db), fakeSSHKeyGenerator{err: expected})
	if _, err := service.Create(context.Background(), users.User{IsAdmin: true}, "Northstar", "northstar"); !errors.Is(err, expected) {
		t.Fatalf("create error = %v", err)
	}
	var count int64
	if err := db.Model(&models.Organization{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("persisted organizations = %d", count)
	}
}

// createSSHKeyTestSchema installs a SQLite-compatible subset of the PostgreSQL schema.
func createSSHKeyTestSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`CREATE TABLE organizations (id uuid PRIMARY KEY, name text NOT NULL, slug text NOT NULL UNIQUE, created_at datetime NOT NULL, updated_at datetime NOT NULL)`,
		`CREATE TABLE organization_ssh_keys (organization_id uuid PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE, public_key text NOT NULL, encrypted_private_key blob NOT NULL, nonce blob NOT NULL, key_algorithm text NOT NULL, encryption_version text NOT NULL, fingerprint text NOT NULL, created_at datetime NOT NULL, updated_at datetime NOT NULL)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
}
