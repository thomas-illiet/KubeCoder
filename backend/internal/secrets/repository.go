package secrets

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{ db *gorm.DB }

// NewRepository creates the secret persistence adapter.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// List returns one scope-filtered page without decrypting values.
func (r *Repository) List(ctx context.Context, organizationID *uuid.UUID, includePlatform bool, query, scope, status, sortBy, sortOrder string, expiringBefore any, limit, offset int) ([]Secret, int64, error) {
	statement := r.db.WithContext(ctx).Model(&Secret{})
	if organizationID == nil {
		statement = statement.Where("scope = ?", ScopePlatform)
	} else if includePlatform {
		statement = statement.Where("scope = ? OR (organization_id = ? AND scope = ?)", ScopePlatform, *organizationID, ScopeOrganization)
	} else {
		statement = statement.Where("organization_id = ? AND scope = ?", *organizationID, ScopeOrganization)
	}
	if query != "" {
		statement = statement.Where("variable_name ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")
	}
	if scope != "" {
		statement = statement.Where("scope = ?", scope)
	}
	if status != "" {
		switch status {
		case StatusExpired:
			statement = statement.Where("expires_at IS NOT NULL AND expires_at <= now()")
		case StatusExpiringSoon:
			statement = statement.Where("expires_at > now() AND expires_at <= ?", expiringBefore)
		case StatusActive:
			statement = statement.Where("expires_at IS NULL OR expires_at > ?", expiringBefore)
		}
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]Secret, 0)
	order := secretOrder(sortBy, sortOrder, expiringBefore)
	err := statement.Order(order).Limit(limit).Offset(offset).Find(&items).Error
	return items, total, err
}

// UserBelongsToOrganization verifies tenant management access.
func (r *Repository) UserBelongsToOrganization(ctx context.Context, userID, organizationID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.OrganizationMembership{}).Where("user_id = ? AND organization_id = ?", userID, organizationID).Count(&count).Error
	return count > 0, err
}

// secretOrder builds a whitelisted stable database ordering expression.
func secretOrder(sortBy, sortOrder string, expiringBefore any) clause.Expr {
	direction := "ASC"
	if sortOrder == SortDescending {
		direction = "DESC"
	}
	expressions := map[string]string{
		"variable_name":     "variable_name",
		"scope":             "scope",
		"binding_count":     "(SELECT COUNT(*) FROM secret_bindings WHERE secret_bindings.secret_id = secrets.id)",
		"value_replaced_at": "value_replaced_at",
		"expires_at":        "expires_at",
	}
	if expression, ok := expressions[sortBy]; ok {
		return clause.Expr{SQL: expression + " " + direction + ", id " + direction, WithoutParentheses: true}
	}
	return clause.Expr{
		SQL:  "CASE WHEN expires_at IS NOT NULL AND expires_at <= now() THEN 'EXPIRED' WHEN expires_at IS NOT NULL AND expires_at <= ? THEN 'EXPIRING_SOON' ELSE 'ACTIVE' END " + direction + ", id " + direction,
		Vars: []any{expiringBefore}, WithoutParentheses: true,
	}
}

// Get returns one secret constrained to the requested ownership boundary.
func (r *Repository) Get(ctx context.Context, id uuid.UUID, organizationID *uuid.UUID) (Secret, error) {
	var item Secret
	statement := r.db.WithContext(ctx).Where("id = ?", id)
	if organizationID == nil {
		statement = statement.Where("scope = ?", ScopePlatform)
	} else {
		statement = statement.Where("organization_id = ?", *organizationID)
	}
	err := statement.First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Secret{}, ErrNotFound
	}
	return item, err
}

// Create stores a secret.
func (r *Repository) Create(ctx context.Context, secret *Secret) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(secret).Error; err != nil {
			return err
		}
		repositories := make([]models.Repository, 0)
		statement := tx.Where("secret_mode = ?", "ALL")
		if secret.OrganizationID != nil {
			statement = statement.Where("organization_id = ?", *secret.OrganizationID)
		}
		if err := statement.Find(&repositories).Error; err != nil {
			return err
		}
		for _, repository := range repositories {
			repositoryID := repository.ID
			binding := Binding{ID: uuid.New(), SecretID: secret.ID, TargetType: TargetRepository, RepositoryID: &repositoryID, CreatedAt: secret.CreatedAt}
			if err := tx.Create(&binding).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrConflict
	}
	return err
}

// Replace atomically overwrites protected storage on the existing record.
func (r *Repository) Replace(ctx context.Context, secret Secret) error {
	result := r.db.WithContext(ctx).Model(&Secret{}).Where("id = ?", secret.ID).Updates(map[string]any{"encrypted_value": secret.EncryptedValue, "nonce": secret.Nonce, "fingerprint": secret.Fingerprint, "encryption_version": secret.EncryptionVersion, "expires_at": secret.ExpiresAt, "value_replaced_at": secret.ValueReplacedAt, "updated_at": secret.UpdatedAt})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// PlatformVariableExists reports whether a global variable shadows tenant scopes.
func (r *Repository) PlatformVariableExists(ctx context.Context, variableName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Secret{}).Where("scope = ? AND variable_name = ?", ScopePlatform, variableName).Count(&count).Error
	return count > 0, err
}

// OrganizationVariableExists reports whether any tenant already owns a variable name.
func (r *Repository) OrganizationVariableExists(ctx context.Context, variableName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Secret{}).Where("scope = ? AND variable_name = ?", ScopeOrganization, variableName).Count(&count).Error
	return count > 0, err
}

// Bindings returns sanitized binding records for one secret.
func (r *Repository) Bindings(ctx context.Context, secretID uuid.UUID) ([]Binding, error) {
	items := make([]Binding, 0)
	err := r.db.WithContext(ctx).Preload("Agent").Preload("Repository").Where("secret_id = ?", secretID).Order("created_at, id").Find(&items).Error
	return items, err
}

// Binding returns one binding constrained to its parent secret.
func (r *Repository) Binding(ctx context.Context, secretID, bindingID uuid.UUID) (Binding, error) {
	var item Binding
	err := r.db.WithContext(ctx).Where("id = ? AND secret_id = ?", bindingID, secretID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Binding{}, ErrNotFound
	}
	return item, err
}

// BindingsBySecretIDs loads all bindings for one result page in a bounded query set.
func (r *Repository) BindingsBySecretIDs(ctx context.Context, secretIDs []uuid.UUID) (map[uuid.UUID][]Binding, error) {
	result := make(map[uuid.UUID][]Binding, len(secretIDs))
	if len(secretIDs) == 0 {
		return result, nil
	}
	items := make([]Binding, 0)
	if err := r.db.WithContext(ctx).Preload("Agent").Preload("Repository").Where("secret_id IN ?", secretIDs).Order("created_at, id").Find(&items).Error; err != nil {
		return nil, err
	}
	for _, item := range items {
		result[item.SecretID] = append(result[item.SecretID], item)
	}
	return result, nil
}

// AddBinding creates an explicit target assignment.
func (r *Repository) AddBinding(ctx context.Context, binding *Binding) error {
	err := r.db.WithContext(ctx).Create(binding).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrConflict
	}
	return err
}

// RemoveBinding deletes one explicit target assignment.
func (r *Repository) RemoveBinding(ctx context.Context, secretID, bindingID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ? AND secret_id = ?", bindingID, secretID).Delete(&Binding{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// RepositoryBelongs verifies repository tenancy.
func (r *Repository) RepositoryBelongs(ctx context.Context, repositoryID, organizationID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Repository{}).Where("id = ? AND organization_id = ?", repositoryID, organizationID).Count(&count).Error
	return count == 1, err
}

// AgentExists verifies an agent target.
func (r *Repository) AgentExists(ctx context.Context, agentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Agent{}).Where("id = ?", agentID).Count(&count).Error
	return count == 1, err
}

// Targets lists the binding targets available in an administration context.
func (r *Repository) Targets(ctx context.Context, organizationID *uuid.UUID) ([]Target, error) {
	agents := make([]models.Agent, 0)
	if err := r.db.WithContext(ctx).Where("active = true").Order("name, id").Find(&agents).Error; err != nil {
		return nil, err
	}
	result := make([]Target, 0, len(agents))
	for _, agent := range agents {
		result = append(result, Target{ID: agent.ID, Type: TargetAgent, Name: agent.Name})
	}
	return result, nil
}
