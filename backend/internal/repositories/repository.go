package repositories

import (
	"context"
	"errors"

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
func (r *RepositoryStore) Create(ctx context.Context, row *Repository, agentID *uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(row).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrConflict
			}
			return err
		}
		return replaceBinding(tx, row.ID, agentID)
	})
}

// Update changes a repository and its optional agent binding atomically.
func (r *RepositoryStore) Update(ctx context.Context, row *Repository, agentID *uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&Repository{}).Where("organization_id = ? AND id = ?", row.OrganizationID, row.ID).Updates(map[string]any{
			"name": row.Name, "provider": row.Provider, "clone_url": row.CloneURL, "default_branch": row.DefaultBranch, "include_submodules": row.IncludeSubmodules,
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
		return replaceBinding(tx, row.ID, agentID)
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

// replaceBinding upserts or removes the one-to-one agent assignment.
func replaceBinding(tx *gorm.DB, repositoryID uuid.UUID, agentID *uuid.UUID) error {
	if agentID == nil {
		return tx.Delete(&models.RepositoryAgentBinding{}, "repository_id = ?", repositoryID).Error
	}
	binding := models.RepositoryAgentBinding{RepositoryID: repositoryID, AgentID: *agentID}
	return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "repository_id"}}, DoUpdates: clause.AssignmentColumns([]string{"agent_id", "updated_at"})}).Create(&binding).Error
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
	for _, row := range rows {
		var public *agents.PublicAgent
		if agent, ok := byRepository[row.ID]; ok {
			public = &agents.PublicAgent{ID: agent.ID.String(), Name: agent.Name, Description: agent.Description, Capabilities: agent.Capabilities, Active: agent.Active}
		}
		items = append(items, Item{ID: row.ID, OrganizationID: row.OrganizationID, Name: row.Name, Provider: row.Provider, CloneURL: row.CloneURL, DefaultBranch: row.DefaultBranch, IncludeSubmodules: row.IncludeSubmodules, Agent: public, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return items, total, nil
}
