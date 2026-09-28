package repositories

import (
	"testing"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestReplaceSecretBindingsSupportsAllAndSelected verifies both repository policies.
func TestReplaceSecretBindingsSupportsAllAndSelected(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE secrets (id text PRIMARY KEY, scope text, organization_id text)`,
		`CREATE TABLE secret_bindings (id text PRIMARY KEY, secret_id text, target_type text, agent_id text, repository_id text, created_at datetime)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	organizationID, otherOrganizationID, repositoryID := uuid.New(), uuid.New(), uuid.New()
	platformID, organizationSecretID, otherSecretID := uuid.New(), uuid.New(), uuid.New()
	for _, values := range [][]any{{platformID, "PLATFORM", nil}, {organizationSecretID, "ORGANIZATION", organizationID}, {otherSecretID, "ORGANIZATION", otherOrganizationID}} {
		if err := db.Exec("INSERT INTO secrets (id, scope, organization_id) VALUES (?, ?, ?)", values...).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := replaceSecretBindings(db, organizationID, repositoryID, SecretModeAll, nil); err != nil {
		t.Fatal(err)
	}
	var allBindings []models.SecretBinding
	if err := db.Where("repository_id = ?", repositoryID).Find(&allBindings).Error; err != nil {
		t.Fatal(err)
	}
	if len(allBindings) != 2 {
		t.Fatalf("ALL bindings = %d, want platform plus organization", len(allBindings))
	}
	if err := replaceSecretBindings(db, organizationID, repositoryID, SecretModeSelected, []uuid.UUID{organizationSecretID}); err != nil {
		t.Fatal(err)
	}
	var selected []models.SecretBinding
	if err := db.Where("repository_id = ?", repositoryID).Find(&selected).Error; err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || selected[0].SecretID != organizationSecretID {
		t.Fatalf("SELECTED bindings = %#v", selected)
	}
}
