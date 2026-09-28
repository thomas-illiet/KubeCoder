package organizations

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository persists organizations and memberships in PostgreSQL.
type Repository struct{ db *gorm.DB }

// NewRepository creates an organization repository backed by GORM.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// ListForUser returns organizations accessible to a user.
func (r *Repository) ListForUser(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]Organization, int64, error) {
	statement := r.db.WithContext(ctx).Model(&Organization{}).
		Joins("JOIN organization_memberships ON organization_memberships.organization_id = organizations.id").
		Where("organization_memberships.user_id = ?", userID)
	if query != "" {
		statement = statement.Where("organizations.name ILIKE ? OR organizations.slug ILIKE ?", "%"+query+"%", "%"+query+"%")
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count user organizations: %w", err)
	}
	result := make([]Organization, 0)
	if err := statement.Order("organizations.name, organizations.id").Limit(limit).Offset(offset).Find(&result).Error; err != nil {
		return nil, 0, fmt.Errorf("list user organizations: %w", err)
	}
	return result, total, nil
}

// FindForUserBySlug returns an organization only when the user is a member.
func (r *Repository) FindForUserBySlug(ctx context.Context, userID uuid.UUID, slug string) (Organization, error) {
	var result Organization
	err := r.db.WithContext(ctx).Model(&Organization{}).
		Joins("JOIN organization_memberships ON organization_memberships.organization_id = organizations.id").
		Where("organization_memberships.user_id = ? AND organizations.slug = ?", userID, slug).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Organization{}, ErrForbidden
	}
	return result, err
}

// SetPreferred stores an accessible organization as the user preference.
func (r *Repository) SetPreferred(ctx context.Context, userID uuid.UUID, slug string) (Organization, error) {
	var result Organization
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&Organization{}).
			Joins("JOIN organization_memberships ON organization_memberships.organization_id = organizations.id").
			Where("organization_memberships.user_id = ? AND organizations.slug = ?", userID, slug).First(&result).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrForbidden
		}
		if err != nil {
			return err
		}
		return tx.Model(&users.User{}).Where("id = ?", userID).Update("preferred_organization_id", result.ID).Error
	})
	return result, err
}

// List returns all organizations for platform administration.
func (r *Repository) List(ctx context.Context, query string, limit, offset int, orderBy OrganizationOrder, orderDirection OrderDirection) ([]OrganizationSummary, int64, error) {
	statement := r.db.WithContext(ctx).Model(&Organization{})
	if query != "" {
		statement = statement.Where("name ILIKE ? OR slug ILIKE ?", "%"+query+"%", "%"+query+"%")
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	result := make([]OrganizationSummary, 0)
	listStatement := r.db.WithContext(ctx).Model(&Organization{}).
		Select(`organizations.*,
			(SELECT COUNT(*) FROM organization_memberships WHERE organization_memberships.organization_id = organizations.id) AS member_count,
			(SELECT COUNT(*) FROM repositories WHERE repositories.organization_id = organizations.id) AS repository_count`)
	if query != "" {
		listStatement = listStatement.Where("organizations.name ILIKE ? OR organizations.slug ILIKE ?", "%"+query+"%", "%"+query+"%")
	}
	if err := listStatement.Order(organizationOrderClause(orderBy, orderDirection)).Limit(limit).Offset(offset).Scan(&result).Error; err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

// CountRepositories returns the number of repositories across every organization.
func (r *Repository) CountRepositories(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.Repository{}).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count repositories: %w", err)
	}
	return total, nil
}

// organizationOrderClause returns a deterministic SQL order from trusted enum values.
func organizationOrderClause(orderBy OrganizationOrder, direction OrderDirection) string {
	column := "organizations.name"
	switch orderBy {
	case OrganizationOrderCreatedAt:
		column = "organizations.created_at"
	case OrganizationOrderMemberCount:
		column = "member_count"
	case OrganizationOrderRepositoryCount:
		column = "repository_count"
	}
	keyword := "ASC"
	if direction == OrderDescending {
		keyword = "DESC"
	}
	return column + " " + keyword + ", organizations.id ASC"
}

// Create persists a new organization.
func (r *Repository) Create(ctx context.Context, organization *Organization) error {
	if err := r.db.WithContext(ctx).Create(organization).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrConflict
		}
		return err
	}
	return nil
}

// CreateWithSSHKey atomically persists an organization and its encrypted SSH identity.
func (r *Repository) CreateWithSSHKey(ctx context.Context, organization *Organization, key *models.OrganizationSSHKey) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(organization).Error; err != nil {
			return err
		}
		return tx.Create(key).Error
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrConflict
	}
	return err
}

// FindByID returns an organization by its internal identifier.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Organization, error) {
	var result Organization
	if err := r.db.WithContext(ctx).First(&result, "id = ?", id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return Organization{}, ErrNotFound
	} else if err != nil {
		return Organization{}, err
	}
	return result, nil
}

// Rename changes an organization display name without changing its slug.
func (r *Repository) Rename(ctx context.Context, id uuid.UUID, name string) (Organization, error) {
	result, err := r.FindByID(ctx, id)
	if err != nil {
		return Organization{}, err
	}
	if err := r.db.WithContext(ctx).Model(&result).Update("name", name).Error; err != nil {
		return Organization{}, err
	}
	return result, nil
}

// Delete permanently removes an organization and its memberships.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Organization{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListMembers returns provisioned users assigned to an organization.
func (r *Repository) ListMembers(ctx context.Context, organizationID uuid.UUID, query string, limit, offset int, orderBy MemberOrder, orderDirection OrderDirection) ([]Member, int64, error) {
	statement := r.db.WithContext(ctx).Model(&users.User{}).
		Joins("JOIN organization_memberships ON organization_memberships.user_id = users.id").
		Where("organization_memberships.organization_id = ?", organizationID)
	if query != "" {
		pattern := "%" + query + "%"
		statement = statement.Where("users.username ILIKE ? OR users.display_name ILIKE ? OR users.email ILIKE ?", pattern, pattern, pattern)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	result := make([]Member, 0)
	if err := statement.Select("users.*, organization_memberships.created_at AS joined_at").Order(memberOrderClause(orderBy, orderDirection)).Limit(limit).Offset(offset).Scan(&result).Error; err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

// memberOrderClause returns a deterministic SQL order from trusted enum values.
func memberOrderClause(orderBy MemberOrder, direction OrderDirection) string {
	column := "users.display_name"
	switch orderBy {
	case MemberOrderUsername:
		column = "users.username"
	case MemberOrderEmail:
		column = "users.email"
	case MemberOrderJoinedAt:
		column = "organization_memberships.created_at"
	}
	keyword := "ASC"
	if direction == OrderDescending {
		keyword = "DESC"
	}
	return column + " " + keyword + ", users.id ASC"
}

// AddMember creates a membership and rejects duplicates.
func (r *Repository) AddMember(ctx context.Context, organizationID, userID uuid.UUID) error {
	membership := Membership{OrganizationID: organizationID, UserID: userID}
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&membership)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
			return ErrNotFound
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConflict
	}
	return nil
}

// RemoveMember deletes a membership and clears a matching preference atomically.
func (r *Repository) RemoveMember(ctx context.Context, organizationID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&Membership{}, "organization_id = ? AND user_id = ?", organizationID, userID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return tx.Model(&users.User{}).Where("id = ? AND preferred_organization_id = ?", userID, organizationID).
			Update("preferred_organization_id", nil).Error
	})
}
