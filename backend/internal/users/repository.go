package users

import (
	"context"
	"fmt"

	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const firstAdminLockID int64 = 5227809561277816065

type Repository struct{ db *gorm.DB }

// NewRepository creates a PostgreSQL-backed user repository.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Upsert provisions or refreshes a user from a verified OIDC identity.
func (r *Repository) Upsert(ctx context.Context, identity auth.Identity) (User, error) {
	var result User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", firstAdminLockID).Error; err != nil {
			return fmt.Errorf("lock first administrator selection: %w", err)
		}
		var count int64
		if err := tx.Model(&User{}).Count(&count).Error; err != nil {
			return fmt.Errorf("count users: %w", err)
		}
		candidate := User{
			Issuer: identity.Issuer, Subject: identity.Subject, Username: identity.Username,
			DisplayName: identity.DisplayName, Email: identity.Email,
			IsAdmin: count == 0,
		}
		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "issuer"}, {Name: "subject"}},
			DoUpdates: clause.Assignments(map[string]any{
				"username": candidate.Username, "display_name": candidate.DisplayName,
				"email": candidate.Email, "updated_at": gorm.Expr("now()"),
			}),
		}, clause.Returning{}).Create(&candidate).Error
		if err != nil {
			return fmt.Errorf("upsert user: %w", err)
		}
		if err := tx.Preload("PreferredOrganization").First(&result, "id = ?", candidate.ID).Error; err != nil {
			return fmt.Errorf("load user preference: %w", err)
		}
		return nil
	})
	if err != nil {
		return User{}, err
	}
	return result, nil
}

// List returns provisioned users whose identifying fields match the query.
func (r *Repository) List(ctx context.Context, query string, limit, offset int) ([]User, int64, error) {
	statement := r.db.WithContext(ctx).Model(&User{})
	if query != "" {
		pattern := "%" + query + "%"
		statement = statement.Where("username ILIKE ? OR display_name ILIKE ? OR email ILIKE ?", pattern, pattern, pattern)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	result := make([]User, 0)
	if err := statement.Order("display_name, id").Limit(limit).Offset(offset).Find(&result).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return result, total, nil
}
