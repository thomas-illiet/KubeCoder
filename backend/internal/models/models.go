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
