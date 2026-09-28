// Package agents implements global agent definitions and their public projection.
package agents

import (
	"errors"

	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
)

type Agent = models.Agent

// PublicAgent is the non-sensitive agent catalog entry exposed to organization members.
type PublicAgent struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
	Active       bool     `json:"active"`
}

type Input struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	RuntimeAdapter     string   `json:"runtime_adapter"`
	RuntimeVersion     string   `json:"runtime_version"`
	Image              string   `json:"image"`
	Provider           string   `json:"provider"`
	Model              string   `json:"model"`
	SystemPrompt       string   `json:"system_prompt"`
	Capabilities       []string `json:"capabilities"`
	CPUMillis          int      `json:"cpu_millis"`
	MemoryMB           int      `json:"memory_mb"`
	StorageMB          int      `json:"storage_mb"`
	MaxDurationSeconds int      `json:"max_duration_seconds"`
	Active             bool     `json:"active"`
}

var (
	ErrNotFound  = errors.New("agent not found")
	ErrForbidden = errors.New("agent access forbidden")
	ErrConflict  = errors.New("agent conflict")
	ErrInvalid   = errors.New("invalid agent")
)
