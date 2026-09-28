package users

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	"github.com/thomas-illiet/KubeCoder/backend/internal/config"
	"github.com/thomas-illiet/KubeCoder/backend/internal/database"
)

// TestConcurrentFirstUserProvisioning proves that only one initial user becomes administrator.
func TestConcurrentFirstUserProvisioning(t *testing.T) {
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
	repository := NewRepository(db.GORM)

	const users = 10
	errorsChannel := make(chan error, users)
	var group sync.WaitGroup
	for index := range users {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := repository.Upsert(context.Background(), auth.Identity{
				Issuer: "https://issuer.example", Subject: fmt.Sprintf("subject-%d", index),
				Username: fmt.Sprintf("user-%d", index), DisplayName: "Integration User",
			})
			errorsChannel <- err
		}()
	}
	group.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Fatal(err)
		}
	}

	var total, administrators int64
	if err := db.GORM.Model(&User{}).Count(&total).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.GORM.Model(&User{}).Where("is_admin = true").Count(&administrators).Error; err != nil {
		t.Fatal(err)
	}
	if total != users || administrators != 1 {
		t.Fatalf("users = %d, platform administrators = %d", total, administrators)
	}
}
