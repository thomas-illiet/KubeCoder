package organizations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/config"
	"github.com/thomas-illiet/KubeCoder/backend/internal/database"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

// TestOrganizationMembershipLifecycle verifies persistence, preference cleanup, and membership uniqueness.
func TestOrganizationMembershipLifecycle(t *testing.T) {
	dsn := os.Getenv("KUBECODER_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("KUBECODER_TEST_DATABASE_DSN is not set")
	}
	db, err := database.Open(context.Background(), config.DatabaseConfig{
		DSN: dsn, MaxOpenConns: 20, MaxIdleConns: 5, ConnMaxLifetime: time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.SQL.Close()

	ctx := context.Background()
	repository := NewRepository(db.GORM)
	user := users.User{
		ID: uuid.New(), Issuer: "integration", Subject: uuid.NewString(), Username: "member-" + uuid.NewString(),
		DisplayName: "Organization Member", Email: uuid.NewString() + "@example.test",
	}
	if err := db.GORM.WithContext(ctx).Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	defer db.GORM.WithContext(ctx).Delete(&user)

	organization := Organization{Name: "Integration Organization", Slug: "integration-" + uuid.NewString()}
	if err := repository.Create(ctx, &organization); err != nil {
		t.Fatal(err)
	}
	defer repository.Delete(ctx, organization.ID)

	var created atomic.Int32
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			err := repository.AddMember(ctx, organization.ID, user.ID)
			if err == nil {
				created.Add(1)
			} else if !errors.Is(err, ErrConflict) {
				t.Errorf("add membership: %v", err)
			}
		}()
	}
	group.Wait()
	if created.Load() != 1 {
		t.Fatalf("created memberships = %d", created.Load())
	}
	if _, err := repository.SetPreferred(ctx, user.ID, organization.Slug); err != nil {
		t.Fatal(err)
	}
	if err := repository.RemoveMember(ctx, organization.ID, user.ID); err != nil {
		t.Fatal(err)
	}
	var preferred sql.NullString
	if err := db.GORM.Model(&users.User{}).Select("preferred_organization_id").Where("id = ?", user.ID).Scan(&preferred).Error; err != nil {
		t.Fatal(fmt.Errorf("read cleared preference: %w", err))
	}
	if preferred.Valid {
		t.Fatalf("preferred organization was not cleared: %s", preferred.String)
	}
}
