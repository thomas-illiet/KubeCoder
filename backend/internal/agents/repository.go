package agents

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

// NewRepository creates an agent repository backed by GORM.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// List returns a filtered page of global agent definitions.
func (r *Repository) List(ctx context.Context, query string, active *bool, limit, offset int) ([]Agent, int64, error) {
	statement := r.db.WithContext(ctx).Model(&Agent{})
	if query != "" {
		statement = statement.Where("name ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")
	}
	if active != nil {
		statement = statement.Where("active = ?", *active)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]Agent, 0)
	if err := statement.Order("name, id").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Get returns an agent by its identifier.
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Agent, error) {
	var item Agent
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return Agent{}, ErrNotFound
	} else if err != nil {
		return Agent{}, err
	}
	return item, nil
}

// Create persists an agent definition.
func (r *Repository) Create(ctx context.Context, item *Agent) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrConflict
		}
		return err
	}
	return nil
}

// Update replaces the mutable fields of an agent definition.
func (r *Repository) Update(ctx context.Context, item *Agent) error {
	result := r.db.WithContext(ctx).Save(item)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return ErrConflict
	}
	return result.Error
}

// Delete removes an unreferenced agent definition.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	var bindings int64
	if err := r.db.WithContext(ctx).Model(&models.RepositoryAgentBinding{}).Where("agent_id = ?", id).Count(&bindings).Error; err != nil {
		return err
	}
	if bindings > 0 {
		return ErrConflict
	}
	result := r.db.WithContext(ctx).Delete(&Agent{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("delete agent: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
