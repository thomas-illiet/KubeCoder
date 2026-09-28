// Package repositories exposes organization-scoped repository operations.
package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	domain "github.com/thomas-illiet/KubeCoder/backend/internal/repositories"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

type UserService interface {
	Current(context.Context, auth.Identity) (users.User, error)
}
type Service interface {
	List(context.Context, users.User, string, string, string, *bool, *uuid.UUID, int, int) ([]domain.Item, int64, error)
	Get(context.Context, users.User, string, uuid.UUID) (domain.Item, error)
	Create(context.Context, users.User, string, domain.Input) (domain.Item, error)
	Update(context.Context, users.User, string, uuid.UUID, domain.Input) (domain.Item, error)
	Delete(context.Context, users.User, string, uuid.UUID) error
}

type Handler struct {
	logger   *slog.Logger
	verifier auth.Verifier
	users    UserService
	service  Service
}

// NewHandler creates the organization repository HTTP handler.
func NewHandler(logger *slog.Logger, verifier auth.Verifier, users UserService, service Service) *Handler {
	return &Handler{logger: logger, verifier: verifier, users: users, service: service}
}

// RegisterRoutes registers organization repository endpoints.
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("GET /api/v1/organizations/{slug}/repositories", handler.list)
	mux.HandleFunc("POST /api/v1/organizations/{slug}/repositories", handler.create)
	mux.HandleFunc("GET /api/v1/organizations/{slug}/repositories/{id}", handler.get)
	mux.HandleFunc("PATCH /api/v1/organizations/{slug}/repositories/{id}", handler.update)
	mux.HandleFunc("DELETE /api/v1/organizations/{slug}/repositories/{id}", handler.delete)
}

// list returns a filtered page of repositories.
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	page := httpx.ParsePagination(r)
	configured, ok := optionalBool(w, r, "configured")
	if !ok {
		return
	}
	agentID, ok := optionalUUID(w, r, "agent_id")
	if !ok {
		return
	}
	items, total, err := h.service.List(r.Context(), actor, r.PathValue("slug"), r.URL.Query().Get("query"), r.URL.Query().Get("provider"), configured, agentID, page.Limit, page.Offset)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.NewPage(items, total, page))
}

// get returns one repository in the selected organization.
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.actorAndID(w, r)
	if !ok {
		return
	}
	item, err := h.service.Get(r.Context(), actor, r.PathValue("slug"), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

// create decodes and creates an organization repository.
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	var input domain.Input
	if !decode(w, r, &input) {
		return
	}
	item, err := h.service.Create(r.Context(), actor, r.PathValue("slug"), input)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

// update replaces an organization repository configuration.
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.actorAndID(w, r)
	if !ok {
		return
	}
	var input domain.Input
	if !decode(w, r, &input) {
		return
	}
	item, err := h.service.Update(r.Context(), actor, r.PathValue("slug"), id, input)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

// delete permanently removes an organization repository.
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.actorAndID(w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), actor, r.PathValue("slug"), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// actor authenticates and provisions the request user.
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

// actorAndID authenticates the actor and parses a repository identifier.
func (h *Handler) actorAndID(w http.ResponseWriter, r *http.Request) (users.User, uuid.UUID, bool) {
	actor, ok := h.actor(w, r)
	if !ok {
		return users.User{}, uuid.Nil, false
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		badRequest(w, r, "The repository ID is invalid.")
		return users.User{}, uuid.Nil, false
	}
	return actor, id, true
}

// writeError maps known repository errors to problem responses.
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		httpx.WriteProblem(w, r, http.StatusForbidden, "Forbidden", "You do not have access to this organization.")
	case errors.Is(err, domain.ErrNotFound):
		httpx.WriteProblem(w, r, http.StatusNotFound, "Not Found", "The repository was not found.")
	case errors.Is(err, domain.ErrConflict):
		httpx.WriteProblem(w, r, http.StatusConflict, "Conflict", "This repository URL already exists in the organization.")
	case errors.Is(err, domain.ErrInvalid):
		badRequest(w, r, "The repository configuration or filters are invalid.")
	case errors.Is(err, domain.ErrAgent):
		badRequest(w, r, "The selected agent is not available.")
	default:
		h.internalError(w, r, err)
	}
}

// internalError records and hides an unexpected repository failure.
func (h *Handler) internalError(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.ErrorContext(r.Context(), "repository request failed", "request_id", httpx.RequestID(r.Context()), "error", err)
	httpx.WriteProblem(w, r, http.StatusInternalServerError, "Internal Server Error", "The repository operation failed.")
}

// decode reads one strict bounded JSON body.
func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		badRequest(w, r, "The JSON request body is invalid.")
		return false
	}
	return true
}

// optionalBool parses an optional boolean query filter.
func optionalBool(w http.ResponseWriter, r *http.Request, key string) (*bool, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, true
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		badRequest(w, r, "The "+key+" filter is invalid.")
		return nil, false
	}
	return &value, true
}

// optionalUUID parses an optional UUID query filter.
func optionalUUID(w http.ResponseWriter, r *http.Request, key string) (*uuid.UUID, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, true
	}
	value, err := uuid.Parse(raw)
	if err != nil {
		badRequest(w, r, "The "+key+" filter is invalid.")
		return nil, false
	}
	return &value, true
}

// badRequest writes a standard validation problem.
func badRequest(w http.ResponseWriter, r *http.Request, detail string) {
	httpx.WriteProblem(w, r, http.StatusBadRequest, "Invalid request", detail)
}
