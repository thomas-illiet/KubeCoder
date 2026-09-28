// Package users implements application users and their persistence.
package users

import (
	"time"

	"github.com/google/uuid"
)

// User is the local application projection of an OIDC identity.
type User struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Issuer      string    `gorm:"not null;uniqueIndex:users_issuer_subject_key" json:"-"`
	Subject     string    `gorm:"not null;uniqueIndex:users_issuer_subject_key" json:"subject"`
	Username    string    `gorm:"not null" json:"username"`
	DisplayName string    `gorm:"not null" json:"display_name"`
	Email       string    `gorm:"not null" json:"email"`
	IsAdmin     bool      `gorm:"not null" json:"is_admin"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`
}
