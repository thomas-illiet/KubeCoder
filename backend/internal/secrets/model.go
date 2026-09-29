// Package secrets implements write-only platform and tenant secrets.
package secrets

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
)

const (
	ScopePlatform      = "PLATFORM"
	ScopeOrganization  = "ORGANIZATION"
	TargetAgent        = "AGENT"
	TargetRepository   = "REPOSITORY"
	StatusActive       = "ACTIVE"
	StatusExpiringSoon = "EXPIRING_SOON"
	StatusExpired      = "EXPIRED"
	StatusNotExpired   = "NOT_EXPIRED"
	SortAscending      = "asc"
	SortDescending     = "desc"
)

var allowedSortFields = map[string]struct{}{
	"variable_name":     {},
	"scope":             {},
	"binding_count":     {},
	"value_replaced_at": {},
	"expires_at":        {},
	"status":            {},
}

type Secret = models.Secret
type Binding = models.SecretBinding

type Input struct {
	Scope        string     `json:"scope"`
	VariableName string     `json:"variable_name"`
	Description  string     `json:"description"`
	Value        string     `json:"value"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

type ReplaceInput struct {
	Value     string     `json:"value"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type BindingInput struct {
	TargetType string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
}

type BindingView struct {
	ID         uuid.UUID `json:"id"`
	TargetType string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
	TargetName string    `json:"target_name"`
}

type Target struct {
	ID   uuid.UUID `json:"id"`
	Type string    `json:"type"`
	Name string    `json:"name"`
}

// View is the only public projection. It intentionally contains no protected storage fields.
type View struct {
	ID              uuid.UUID     `json:"id"`
	Scope           string        `json:"scope"`
	OrganizationID  *uuid.UUID    `json:"organization_id,omitempty"`
	VariableName    string        `json:"variable_name"`
	Description     string        `json:"description"`
	ExpiresAt       *time.Time    `json:"expires_at"`
	ValueReplacedAt time.Time     `json:"value_replaced_at"`
	Status          string        `json:"status"`
	Bindings        []BindingView `json:"bindings"`
	BindingCount    int           `json:"binding_count"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

var (
	ErrNotFound  = errors.New("secret not found")
	ErrForbidden = errors.New("secret access forbidden")
	ErrConflict  = errors.New("secret conflict")
	ErrInvalid   = errors.New("invalid secret")
)
