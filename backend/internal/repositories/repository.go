package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/agents"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RepositoryStore struct{ db *gorm.DB }

// NewRepository creates a repository store backed by GORM.
func NewRepository(db *gorm.DB) *RepositoryStore { return &RepositoryStore{db: db} }

// List returns one filtered organization repository page.
func (r *RepositoryStore) List(ctx context.Context, organizationID uuid.UUID, query, provider string, configured *bool, agentID *uuid.UUID, limit, offset int) ([]Item, int64, error) {
	statement := r.db.WithContext(ctx).Model(&Repository{}).Where("organization_id = ?", organizationID)
	if query != "" {
		statement = statement.Where("name ILIKE ? OR clone_url ILIKE ? OR default_branch ILIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}
	if provider != "" {
		statement = statement.Where("provider = ?", provider)
	}
	if configured != nil || agentID != nil {
		statement = statement.Joins("LEFT JOIN repository_agent_bindings ON repository_agent_bindings.repository_id = repositories.id")
		if configured != nil && *configured {
			statement = statement.Where("repository_agent_bindings.repository_id IS NOT NULL")
		}
		if configured != nil && !*configured {
			statement = statement.Where("repository_agent_bindings.repository_id IS NULL")
		}
		if agentID != nil {
			statement = statement.Where("repository_agent_bindings.agent_id = ?", *agentID)
		}
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]Repository, 0)
	if err := statement.Order("repositories.name, repositories.id").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return r.decorate(ctx, rows, total)
}

// Get returns one repository scoped to its organization.
func (r *RepositoryStore) Get(ctx context.Context, organizationID, id uuid.UUID) (Item, error) {
	var row Repository
	if err := r.db.WithContext(ctx).Where("organization_id = ? AND id = ?", organizationID, id).First(&row).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return Item{}, ErrNotFound
	} else if err != nil {
		return Item{}, err
	}
	items, _, err := r.decorate(ctx, []Repository{row}, 1)
	if err != nil {
		return Item{}, err
	}
	return items[0], nil
}

// Create persists a repository and its optional agent binding atomically.
func (r *RepositoryStore) Create(ctx context.Context, row *Repository, agentID *uuid.UUID, secretIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(row).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrConflict
			}
			return err
		}
		if err := replaceBinding(tx, row.ID, agentID); err != nil {
			return err
		}
		return replaceSecretBindings(tx, row.OrganizationID, row.ID, row.SecretMode, secretIDs)
	})
}

// Update changes a repository and its optional agent binding atomically.
func (r *RepositoryStore) Update(ctx context.Context, row *Repository, agentID *uuid.UUID, secretIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&Repository{}).Where("organization_id = ? AND id = ?", row.OrganizationID, row.ID).Updates(map[string]any{
			"name": row.Name, "provider": row.Provider, "clone_url": row.CloneURL, "default_branch": row.DefaultBranch, "include_submodules": row.IncludeSubmodules, "secret_mode": row.SecretMode,
		})
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrConflict
		}
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		if err := replaceBinding(tx, row.ID, agentID); err != nil {
			return err
		}
		return replaceSecretBindings(tx, row.OrganizationID, row.ID, row.SecretMode, secretIDs)
	})
}

// Delete permanently removes an organization repository.
func (r *RepositoryStore) Delete(ctx context.Context, organizationID, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("organization_id = ? AND id = ?", organizationID, id).Delete(&Repository{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// AgentAvailable reports whether an active agent can be assigned.
func (r *RepositoryStore) AgentAvailable(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Agent{}).Where("id = ? AND active = true", id).Count(&count).Error
	return count == 1, err
}

// OrganizationSecretsExist verifies that every selected secret belongs to the tenant.
func (r *RepositoryStore) OrganizationSecretsExist(ctx context.Context, organizationID uuid.UUID, ids []uuid.UUID) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	unique := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Secret{}).Where("(scope = ?) OR (scope = ? AND organization_id = ?)", "PLATFORM", "ORGANIZATION", organizationID).Where("id IN ?", ids).Count(&count).Error
	return count == int64(len(unique)), err
}

// replaceBinding upserts or removes the one-to-one agent assignment.
func replaceBinding(tx *gorm.DB, repositoryID uuid.UUID, agentID *uuid.UUID) error {
	if agentID == nil {
		return tx.Delete(&models.RepositoryAgentBinding{}, "repository_id = ?", repositoryID).Error
	}
	binding := models.RepositoryAgentBinding{RepositoryID: repositoryID, AgentID: *agentID}
	return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "repository_id"}}, DoUpdates: clause.AssignmentColumns([]string{"agent_id", "updated_at"})}).Create(&binding).Error
}

// replaceSecretBindings materializes the repository secret policy atomically.
func replaceSecretBindings(tx *gorm.DB, organizationID, repositoryID uuid.UUID, mode string, selectedIDs []uuid.UUID) error {
	if err := tx.Where("repository_id = ?", repositoryID).Delete(&models.SecretBinding{}).Error; err != nil {
		return err
	}
	secretIDs := selectedIDs
	if mode == SecretModeAll {
		if err := tx.Model(&models.Secret{}).Where("scope = ? OR (scope = ? AND organization_id = ?)", "PLATFORM", "ORGANIZATION", organizationID).Pluck("id", &secretIDs).Error; err != nil {
			return err
		}
	}
	for _, secretID := range secretIDs {
		binding := models.SecretBinding{ID: uuid.New(), SecretID: secretID, TargetType: "REPOSITORY", RepositoryID: &repositoryID, CreatedAt: time.Now().UTC()}
		if err := tx.Create(&binding).Error; err != nil {
			return err
		}
	}
	return nil
}

// decorate adds public agent projections to repository rows.
func (r *RepositoryStore) decorate(ctx context.Context, rows []Repository, total int64) ([]Item, int64, error) {
	items := make([]Item, 0, len(rows))
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	bindings := make([]models.RepositoryAgentBinding, 0)
	if len(ids) > 0 {
		if err := r.db.WithContext(ctx).Preload("Agent").Where("repository_id IN ?", ids).Find(&bindings).Error; err != nil {
			return nil, 0, err
		}
	}
	byRepository := make(map[uuid.UUID]models.Agent, len(bindings))
	for _, binding := range bindings {
		byRepository[binding.RepositoryID] = binding.Agent
	}
	secretBindings := make([]models.SecretBinding, 0)
	if len(ids) > 0 {
		if err := r.db.WithContext(ctx).Where("repository_id IN ?", ids).Order("created_at, id").Find(&secretBindings).Error; err != nil {
			return nil, 0, err
		}
	}
	secretIDsByRepository := make(map[uuid.UUID][]uuid.UUID, len(ids))
	for _, binding := range secretBindings {
		if binding.RepositoryID != nil {
			secretIDsByRepository[*binding.RepositoryID] = append(secretIDsByRepository[*binding.RepositoryID], binding.SecretID)
		}
	}
	for _, row := range rows {
		var public *agents.PublicAgent
		if agent, ok := byRepository[row.ID]; ok {
			public = &agents.PublicAgent{ID: agent.ID.String(), Name: agent.Name, Description: agent.Description, Capabilities: agent.Capabilities, Active: agent.Active}
		}
		items = append(items, Item{ID: row.ID, OrganizationID: row.OrganizationID, Name: row.Name, Provider: row.Provider, CloneURL: row.CloneURL, DefaultBranch: row.DefaultBranch, IncludeSubmodules: row.IncludeSubmodules, SecretMode: row.SecretMode, SecretIDs: secretIDsByRepository[row.ID], Agent: public, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return items, total, nil
}
