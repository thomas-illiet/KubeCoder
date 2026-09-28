// Package organizations exposes platform organization administration over HTTP.
package organizations

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	domain "github.com/thomas-illiet/KubeCoder/backend/internal/organizations"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

// UserService resolves and lists provisioned application users.
type UserService interface {
	Current(context.Context, auth.Identity) (users.User, error)
	List(context.Context, users.User, string, int, int) ([]users.User, int64, error)
}

// Service defines organization administration operations required by the handler.
type Service interface {
	List(context.Context, users.User, string, int, int, string, string) ([]domain.OrganizationSummary, int64, error)
	CountRepositories(context.Context, users.User) (int64, error)
	Create(context.Context, users.User, string, string) (domain.Organization, error)
	Get(context.Context, users.User, uuid.UUID) (domain.Organization, error)
	Rename(context.Context, users.User, uuid.UUID, string) (domain.Organization, error)
	Delete(context.Context, users.User, uuid.UUID) error
	ListMembers(context.Context, users.User, uuid.UUID, string, int, int, string, string) ([]domain.Member, int64, error)
	AddMember(context.Context, users.User, uuid.UUID, uuid.UUID) error
	RemoveMember(context.Context, users.User, uuid.UUID, uuid.UUID) error
}

// Handler serves platform organization administration.
type Handler struct {
	logger   *slog.Logger
	verifier auth.Verifier
	users    UserService
	service  Service
}

// NewHandler creates a platform organization administration handler.
func NewHandler(logger *slog.Logger, verifier auth.Verifier, users UserService, service Service) *Handler {
	return &Handler{logger: logger, verifier: verifier, users: users, service: service}
}

// RegisterRoutes registers platform organization administration endpoints.
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("GET /api/v1/admin/organizations", handler.list)
	mux.HandleFunc("POST /api/v1/admin/organizations", handler.create)
	mux.HandleFunc("GET /api/v1/admin/organizations/{id}", handler.get)
	mux.HandleFunc("PATCH /api/v1/admin/organizations/{id}", handler.rename)
	mux.HandleFunc("DELETE /api/v1/admin/organizations/{id}", handler.delete)
	mux.HandleFunc("GET /api/v1/admin/organizations/{id}/members", handler.listMembers)
	mux.HandleFunc("POST /api/v1/admin/organizations/{id}/members", handler.addMember)
	mux.HandleFunc("DELETE /api/v1/admin/organizations/{id}/members/{userID}", handler.removeMember)
	mux.HandleFunc("GET /api/v1/admin/users", handler.listUsers)
}

type organizationInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type memberInput struct {
	UserID uuid.UUID `json:"user_id"`
}

// list returns all organizations to a platform administrator.
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	pagination := httpx.ParsePagination(r)
	items, total, err := h.service.List(
		r.Context(),
		actor,
		r.URL.Query().Get("query"),
		pagination.Limit,
		pagination.Offset,
		r.URL.Query().Get("order_by"),
		r.URL.Query().Get("order_direction"),
	)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	repositoryTotal, err := h.service.CountRepositories(r.Context(), actor)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	page := httpx.NewPage(items, total, pagination)
	httpx.WriteJSON(w, http.StatusOK, struct {
		httpx.Page[domain.OrganizationSummary]
		RepositoryTotal int64 `json:"repository_total"`
	}{Page: page, RepositoryTotal: repositoryTotal})
}

// create validates and creates an organization.
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	var input organizationInput
	if !decode(w, r, &input) {
		return
	}
	item, err := h.service.Create(r.Context(), actor, input.Name, input.Slug)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

// get returns one organization to a platform administrator.
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.actorAndID(w, r)
	if !ok {
		return
	}
	item, err := h.service.Get(r.Context(), actor, id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

// rename updates an organization display name.
func (h *Handler) rename(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.actorAndID(w, r)
	if !ok {
		return
	}
	var input organizationInput
	if !decode(w, r, &input) {
		return
	}
	item, err := h.service.Rename(r.Context(), actor, id, input.Name)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

// delete permanently removes an organization.
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.actorAndID(w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), actor, id); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listMembers returns provisioned users assigned to an organization.
func (h *Handler) listMembers(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.actorAndID(w, r)
	if !ok {
		return
	}
	pagination := httpx.ParsePagination(r)
	items, total, err := h.service.ListMembers(
		r.Context(),
		actor,
		id,
		r.URL.Query().Get("query"),
		pagination.Limit,
		pagination.Offset,
		r.URL.Query().Get("order_by"),
		r.URL.Query().Get("order_direction"),
	)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.NewPage(items, total, pagination))
}

// addMember grants organization access to a provisioned user.
func (h *Handler) addMember(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.actorAndID(w, r)
	if !ok {
		return
	}
	var input memberInput
	if !decode(w, r, &input) {
		return
	}
	if err := h.service.AddMember(r.Context(), actor, id, input.UserID); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// removeMember revokes organization access from a user.
func (h *Handler) removeMember(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.actorAndID(w, r)
	if !ok {
		return
	}
	userID, err := uuid.Parse(r.PathValue("userID"))
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "Invalid request", "The user ID is invalid.")
		return
	}
	if err := h.service.RemoveMember(r.Context(), actor, id, userID); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listUsers returns provisioned users available for membership assignment.
func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	pagination := httpx.ParsePagination(r)
	items, total, err := h.users.List(r.Context(), actor, r.URL.Query().Get("query"), pagination.Limit, pagination.Offset)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.NewPage(items, total, pagination))
}

// actorAndID authenticates the actor and parses the organization ID.
func (h *Handler) actorAndID(w http.ResponseWriter, r *http.Request) (users.User, uuid.UUID, bool) {
	actor, ok := h.actor(w, r)
	if !ok {
		return users.User{}, uuid.Nil, false
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "Invalid request", "The organization ID is invalid.")
		return users.User{}, uuid.Nil, false
	}
	return actor, id, true
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
	if !actor.IsAdmin {
		httpx.WriteProblem(w, r, http.StatusForbidden, "Forbidden", "Platform administrator access is required.")
		return users.User{}, false
	}
	return actor, true
}

// writeError converts domain errors to problem responses.
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		httpx.WriteProblem(w, r, http.StatusForbidden, "Forbidden", "Platform administrator access is required.")
	case errors.Is(err, domain.ErrNotFound):
		httpx.WriteProblem(w, r, http.StatusNotFound, "Not Found", "The organization or membership was not found.")
	case errors.Is(err, domain.ErrConflict):
		httpx.WriteProblem(w, r, http.StatusConflict, "Conflict", "The organization slug or membership already exists.")
	case errors.Is(err, domain.ErrInvalid):
		httpx.WriteProblem(w, r, http.StatusBadRequest, "Invalid request", "The organization name or slug is invalid.")
	default:
		h.internalError(w, r, err)
	}
}

// internalError logs an unexpected failure without exposing sensitive data.
func (h *Handler) internalError(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.ErrorContext(r.Context(), "organization administration failed", "request_id", httpx.RequestID(r.Context()), "error", err)
	httpx.WriteProblem(w, r, http.StatusInternalServerError, "Internal Server Error", "The organization operation failed.")
}

// decode reads one strict JSON request body.
func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "Invalid request", "The JSON request body is invalid.")
		return false
	}
	return true
}
