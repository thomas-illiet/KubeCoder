package sshkeys

import (
	"bytes"
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/config"
	"github.com/thomas-illiet/KubeCoder/backend/internal/database"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
)

// TestRegenerateIsConcurrentAndCascades verifies concurrent replacement and the organization ownership constraint.
func TestRegenerateIsConcurrentAndCascades(t *testing.T) {
	dsn := os.Getenv("KUBECODER_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("KUBECODER_TEST_DATABASE_DSN is not set")
	}
	db, err := database.Open(context.Background(), config.DatabaseConfig{DSN: dsn, MaxOpenConns: 20, MaxIdleConns: 5, ConnMaxLifetime: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	defer db.SQL.Close()

	organization := models.Organization{ID: uuid.New(), Name: "SSH Key Integration", Slug: "ssh-key-" + uuid.NewString()}
	if err := db.GORM.Create(&organization).Error; err != nil {
		t.Fatal(err)
	}
	defer db.GORM.Delete(&organization)
	service, err := NewService(db.GORM, bytes.Repeat([]byte{3}, 32))
	if err != nil {
		t.Fatal(err)
	}
	initial, err := service.Generate(organization.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.GORM.Create(&initial).Error; err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := service.Regenerate(context.Background(), organization.ID); err != nil {
				t.Errorf("regenerate: %v", err)
			}
		}()
	}
	group.Wait()
	var count int64
	if err := db.GORM.Model(&models.OrganizationSSHKey{}).Where("organization_id = ?", organization.ID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("SSH key rows = %d", count)
	}
	if err := db.GORM.Delete(&organization).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.GORM.Model(&models.OrganizationSSHKey{}).Where("organization_id = ?", organization.ID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("SSH key rows after organization deletion = %d", count)
	}
}
