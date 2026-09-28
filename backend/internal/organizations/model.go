// Package organizations implements organizations, memberships, and persistence.
package organizations

import (
	"errors"
	"time"

	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

// Organization is an isolated tenant visible to its members.
type Organization = models.Organization

// OrganizationSummary augments an organization with administration-only aggregate data.
type OrganizationSummary struct {
	Organization `gorm:"embedded"`
	MemberCount  int64 `gorm:"column:member_count" json:"member_count"`
}

// Membership grants a user access to an organization.
type Membership = models.OrganizationMembership

// Member represents a provisioned user and the date their organization access was granted.
type Member struct {
	users.User `gorm:"embedded"`
	JoinedAt   time.Time `gorm:"column:joined_at" json:"joined_at"`
}

var (
	// ErrNotFound indicates that an organization does not exist.
	ErrNotFound = errors.New("organization not found")
	// ErrForbidden indicates that the actor cannot perform an operation.
	ErrForbidden = errors.New("organization access forbidden")
	// ErrConflict indicates a duplicate slug or membership.
	ErrConflict = errors.New("organization conflict")
	// ErrInvalid indicates invalid organization input.
	ErrInvalid = errors.New("invalid organization")
)
