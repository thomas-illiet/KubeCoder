package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
func (r *Repository) List(ctx context.Context, query, role string, limit, offset int, orderBy, orderDirection string) ([]User, int64, error) {
	statement := r.db.WithContext(ctx).Model(&User{})
	if query != "" {
		pattern := "%" + query + "%"
		statement = statement.Where("username ILIKE ? OR display_name ILIKE ? OR email ILIKE ?", pattern, pattern, pattern)
	}
	if role == "admin" {
		statement = statement.Where("is_admin = ?", true)
	} else if role == "user" {
		statement = statement.Where("is_admin = ?", false)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	result := make([]User, 0)
	orderColumns := map[string]string{"display_name": "display_name", "username": "username", "email": "email", "is_admin": "is_admin", "created_at": "created_at", "updated_at": "updated_at"}
	order := orderColumns[orderBy] + " " + orderDirection + ", id " + orderDirection
	if err := statement.Preload("PreferredOrganization").Order(order).Limit(limit).Offset(offset).Find(&result).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return result, total, nil
}

// UpdateRole changes one user's platform role while preserving at least one administrator.
func (r *Repository) UpdateRole(ctx context.Context, id uuid.UUID, isAdmin bool) (User, error) {
	var result User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", firstAdminLockID).Error; err != nil {
			return fmt.Errorf("lock administrator update: %w", err)
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&result, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrNotFound
			}
			return fmt.Errorf("load user: %w", err)
		}
		if result.IsAdmin && !isAdmin {
			var count int64
			if err := tx.Model(&User{}).Where("is_admin = ?", true).Count(&count).Error; err != nil {
				return fmt.Errorf("count administrators: %w", err)
			}
			if count <= 1 {
				return ErrConflict
			}
		}
		if err := tx.Model(&result).Update("is_admin", isAdmin).Error; err != nil {
			return fmt.Errorf("update user role: %w", err)
		}
		result.IsAdmin = isAdmin
		return tx.Preload("PreferredOrganization").First(&result, "id = ?", id).Error
	})
	return result, err
}

// Delete removes one user and cascades their organization memberships.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", firstAdminLockID).Error; err != nil {
			return fmt.Errorf("lock user deletion: %w", err)
		}
		var target User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&target, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrNotFound
			}
			return fmt.Errorf("load user: %w", err)
		}
		if target.IsAdmin {
			var count int64
			if err := tx.Model(&User{}).Where("is_admin = ?", true).Count(&count).Error; err != nil {
				return fmt.Errorf("count administrators: %w", err)
			}
			if count <= 1 {
				return ErrConflict
			}
		}
		if err := tx.Delete(&target).Error; err != nil {
			return fmt.Errorf("delete user: %w", err)
		}
		return nil
	})
}
