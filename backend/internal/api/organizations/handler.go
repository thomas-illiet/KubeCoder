// Package organizations exposes authenticated organization operations over HTTP.
package organizations

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	domain "github.com/thomas-illiet/KubeCoder/backend/internal/organizations"
	"github.com/thomas-illiet/KubeCoder/backend/internal/sshkeys"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

// UserService resolves the local user represented by a verified identity.
type UserService interface {
	Current(context.Context, auth.Identity) (users.User, error)
}

// Service defines organization operations required by the HTTP handler.
type Service interface {
	ListForUser(context.Context, users.User, string, int, int) ([]domain.Organization, int64, error)
	GetForUser(context.Context, users.User, string) (domain.Organization, error)
	SetPreferred(context.Context, users.User, string) (domain.Organization, error)
}

// SSHKeyService exposes organization SSH key public metadata and rotation.
type SSHKeyService interface {
	Get(context.Context, uuid.UUID) (models.OrganizationSSHKey, error)
	Regenerate(context.Context, uuid.UUID) (models.OrganizationSSHKey, error)
}

// Handler serves the authenticated organization API category.
type Handler struct {
	logger   *slog.Logger
	verifier auth.Verifier
	users    UserService
	service  Service
	sshKeys  SSHKeyService
}

// NewHandler creates an authenticated organization API handler.
func NewHandler(logger *slog.Logger, verifier auth.Verifier, users UserService, service Service, sshKeys SSHKeyService) *Handler {
	return &Handler{logger: logger, verifier: verifier, users: users, service: service, sshKeys: sshKeys}
}

// RegisterRoutes registers endpoints for the organization API category.
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("GET /api/v1/organizations", handler.list)
	mux.HandleFunc("GET /api/v1/organizations/{slug}", handler.get)
	mux.HandleFunc("PUT /api/v1/organizations/{slug}/preferred", handler.prefer)
	mux.HandleFunc("GET /api/v1/organizations/{slug}/ssh-key", handler.getSSHKey)
	mux.HandleFunc("POST /api/v1/organizations/{slug}/ssh-key/regenerate", handler.regenerateSSHKey)
}

// getSSHKey returns only the public SSH identity to an organization member.
func (h *Handler) getSSHKey(w http.ResponseWriter, r *http.Request) {
	h.withOrganizationSSHKey(w, r, false)
}

// regenerateSSHKey atomically creates a missing identity or replaces an existing one.
func (h *Handler) regenerateSSHKey(w http.ResponseWriter, r *http.Request) {
	h.withOrganizationSSHKey(w, r, true)
}

// withOrganizationSSHKey checks membership before reading or rotating public key metadata.
func (h *Handler) withOrganizationSSHKey(w http.ResponseWriter, r *http.Request, regenerate bool) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	organization, err := h.service.GetForUser(r.Context(), actor, r.PathValue("slug"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	var key models.OrganizationSSHKey
	if regenerate {
		key, err = h.sshKeys.Regenerate(r.Context(), organization.ID)
	} else {
		key, err = h.sshKeys.Get(r.Context(), organization.ID)
	}
	if errors.Is(err, sshkeys.ErrNotFound) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "Not Found", "This organization does not have an SSH key yet.")
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, key)
}

// list returns organizations accessible to the current user.
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	pagination := httpx.ParsePagination(r)
	items, total, err := h.service.ListForUser(r.Context(), actor, r.URL.Query().Get("query"), pagination.Limit, pagination.Offset)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.NewPage(items, total, pagination))
}

// get returns one organization when the current user is a member.
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	item, err := h.service.GetForUser(r.Context(), actor, r.PathValue("slug"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

// prefer records one accessible organization as the current user preference.
func (h *Handler) prefer(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	item, err := h.service.SetPreferred(r.Context(), actor, r.PathValue("slug"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

// actor authenticates the request and provisions its local user.
func (h *Handler) actor(w http.ResponseWriter, r *http.Request) (users.User, bool) {
	identity, err := auth.FromRequest(r.Context(), h.verifier, r)
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusUnauthorized, "Unauthorized", "A valid bearer token is required.")
		return users.User{}, false
	}
	actor, err := h.users.Current(r.Context(), identity)
	if err != nil {
		h.internalError(w, r, err)
		return users.User{}, false
	}
	return actor, true
}

// writeError converts known organization errors to problem responses.
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, domain.ErrForbidden) {
		httpx.WriteProblem(w, r, http.StatusForbidden, "Forbidden", "You do not have access to this organization.")
		return
	}
	h.internalError(w, r, err)
}

// internalError logs an unexpected failure without exposing sensitive data.
func (h *Handler) internalError(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.ErrorContext(r.Context(), "organization request failed", "request_id", httpx.RequestID(r.Context()), "error", err)
	httpx.WriteProblem(w, r, http.StatusInternalServerError, "Internal Server Error", "The organization operation failed.")
}
