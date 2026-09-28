// Package repositories implements organization-scoped Git repository configuration.
package repositories

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/agents"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
)

type Repository = models.Repository

type Item struct {
	ID                uuid.UUID           `json:"id"`
	OrganizationID    uuid.UUID           `json:"organization_id"`
	Name              string              `json:"name"`
	Provider          string              `json:"provider"`
	CloneURL          string              `json:"clone_url"`
	DefaultBranch     string              `json:"default_branch"`
	IncludeSubmodules bool                `json:"include_submodules"`
	Agent             *agents.PublicAgent `json:"agent"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

type Input struct {
	Name              string     `json:"name"`
	Provider          string     `json:"provider"`
	CloneURL          string     `json:"clone_url"`
	DefaultBranch     string     `json:"default_branch"`
	IncludeSubmodules bool       `json:"include_submodules"`
	AgentID           *uuid.UUID `json:"agent_id"`
}

var (
	ErrNotFound  = errors.New("repository not found")
	ErrForbidden = errors.New("repository access forbidden")
	ErrConflict  = errors.New("repository conflict")
	ErrInvalid   = errors.New("invalid repository")
	ErrAgent     = errors.New("invalid repository agent")
)
