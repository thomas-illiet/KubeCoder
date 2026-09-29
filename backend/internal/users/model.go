// Package users implements application users and their persistence.
package users

import (
	"errors"

	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
)

// User is the local application projection of an OIDC identity.
type User = models.User

var (
	ErrForbidden = errors.New("user administration forbidden")
	ErrNotFound  = errors.New("user not found")
	ErrConflict  = errors.New("user administration conflict")
	ErrInvalid   = errors.New("invalid user administration input")
)
