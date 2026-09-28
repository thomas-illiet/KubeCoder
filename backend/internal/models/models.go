// Package models declares the GORM models that define the PostgreSQL schema.
package models

import (
	"time"

	"github.com/google/uuid"
)

// User is the local application projection of an OIDC identity.
type User struct {
	ID                      uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Issuer                  string        `gorm:"not null;uniqueIndex:users_issuer_subject_key" json:"-"`
	Subject                 string        `gorm:"not null;uniqueIndex:users_issuer_subject_key" json:"subject"`
	Username                string        `gorm:"not null" json:"username"`
	DisplayName             string        `gorm:"not null" json:"display_name"`
	Email                   string        `gorm:"not null" json:"email"`
	IsAdmin                 bool          `gorm:"not null" json:"is_admin"`
	PreferredOrganizationID *uuid.UUID    `gorm:"type:uuid;index" json:"-"`
	PreferredOrganization   *Organization `gorm:"foreignKey:PreferredOrganizationID;constraint:OnDelete:SET NULL" json:"preferred_organization"`
	CreatedAt               time.Time     `gorm:"not null" json:"created_at"`
	UpdatedAt               time.Time     `gorm:"not null" json:"updated_at"`
}

// Organization is an isolated tenant that users can access through memberships.
type Organization struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Slug      string    `gorm:"not null;uniqueIndex" json:"slug"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// OrganizationMembership grants a user access to an organization.
type OrganizationMembership struct {
	OrganizationID uuid.UUID    `gorm:"type:uuid;primaryKey" json:"organization_id"`
	UserID         uuid.UUID    `gorm:"type:uuid;primaryKey" json:"user_id"`
	Organization   Organization `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	User           User         `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt      time.Time    `gorm:"not null" json:"created_at"`
}

// OrganizationSSHKey is the encrypted SSH identity owned by one organization.
// The private key and encryption metadata are never serialized outside the backend.
type OrganizationSSHKey struct {
	OrganizationID      uuid.UUID    `gorm:"type:uuid;primaryKey" json:"-"`
	Organization        Organization `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	PublicKey           string       `gorm:"not null" json:"public_key"`
	EncryptedPrivateKey []byte       `gorm:"not null" json:"-"`
	Nonce               []byte       `gorm:"not null" json:"-"`
	KeyAlgorithm        string       `gorm:"not null" json:"algorithm"`
	EncryptionVersion   string       `gorm:"not null" json:"-"`
	Fingerprint         string       `gorm:"not null" json:"fingerprint"`
	CreatedAt           time.Time    `gorm:"not null" json:"created_at"`
	UpdatedAt           time.Time    `gorm:"not null" json:"updated_at"`
}

// Agent is a global platform-managed runtime definition.
type Agent struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name            string    `gorm:"not null;uniqueIndex" json:"name"`
	Description     string    `gorm:"not null" json:"description"`
	RuntimeAdapter  string    `gorm:"not null" json:"runtime_adapter"`
	RuntimeVersion  string    `gorm:"not null" json:"runtime_version"`
	Image           string    `gorm:"not null" json:"image"`
	Provider        string    `gorm:"not null" json:"provider"`
	Model           string    `gorm:"not null" json:"model"`
	SystemPrompt    string    `gorm:"not null" json:"system_prompt"`
	Capabilities    []string  `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"capabilities"`
	CPUMillis       int       `gorm:"not null" json:"cpu_millis"`
	MemoryMB        int       `gorm:"not null" json:"memory_mb"`
	StorageMB       int       `gorm:"not null" json:"storage_mb"`
	MaxDurationSecs int       `gorm:"not null" json:"max_duration_seconds"`
	Active          bool      `gorm:"not null;default:true" json:"active"`
	CreatedAt       time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null" json:"updated_at"`
}

// Repository is an organization-owned Git repository configuration.
type Repository struct {
	ID                uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID    uuid.UUID    `gorm:"type:uuid;not null;uniqueIndex:repositories_organization_clone_url_key" json:"organization_id"`
	Organization      Organization `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Name              string       `gorm:"not null" json:"name"`
	Provider          string       `gorm:"not null" json:"provider"`
	CloneURL          string       `gorm:"not null;uniqueIndex:repositories_organization_clone_url_key" json:"clone_url"`
	DefaultBranch     string       `gorm:"not null" json:"default_branch"`
	IncludeSubmodules bool         `gorm:"not null;default:false" json:"include_submodules"`
	CreatedAt         time.Time    `gorm:"not null" json:"created_at"`
	UpdatedAt         time.Time    `gorm:"not null" json:"updated_at"`
}

// RepositoryAgentBinding assigns at most one current agent to a repository.
type RepositoryAgentBinding struct {
	RepositoryID uuid.UUID  `gorm:"type:uuid;primaryKey" json:"repository_id"`
	AgentID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"agent_id"`
	Repository   Repository `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Agent        Agent      `gorm:"constraint:OnDelete:RESTRICT" json:"agent"`
	CreatedAt    time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"not null" json:"updated_at"`
}

// Secret is a write-only protected value. EncryptedValue, Nonce and Fingerprint
// are deliberately excluded from every JSON projection.
type Secret struct {
	ID                uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Scope             string        `gorm:"not null" json:"scope"`
	OrganizationID    *uuid.UUID    `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	Organization      *Organization `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	VariableName      string        `gorm:"not null" json:"variable_name"`
	Description       string        `gorm:"not null" json:"description"`
	EncryptedValue    []byte        `gorm:"not null" json:"-"`
	Nonce             []byte        `gorm:"not null" json:"-"`
	Fingerprint       string        `gorm:"not null" json:"-"`
	EncryptionVersion string        `gorm:"not null" json:"-"`
	ExpiresAt         *time.Time    `json:"expires_at"`
	ValueReplacedAt   time.Time     `gorm:"not null" json:"value_replaced_at"`
	CreatedAt         time.Time     `gorm:"not null" json:"created_at"`
	UpdatedAt         time.Time     `gorm:"not null" json:"updated_at"`
}

// SecretBinding makes the current value available to one approved target.
type SecretBinding struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SecretID     uuid.UUID   `gorm:"type:uuid;not null;index" json:"secret_id"`
	Secret       Secret      `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	TargetType   string      `gorm:"not null" json:"target_type"`
	AgentID      *uuid.UUID  `gorm:"type:uuid;index" json:"agent_id,omitempty"`
	Agent        *Agent      `gorm:"constraint:OnDelete:CASCADE" json:"agent,omitempty"`
	RepositoryID *uuid.UUID  `gorm:"type:uuid;index" json:"repository_id,omitempty"`
	Repository   *Repository `gorm:"constraint:OnDelete:CASCADE" json:"repository,omitempty"`
	CreatedAt    time.Time   `gorm:"not null" json:"created_at"`
}
