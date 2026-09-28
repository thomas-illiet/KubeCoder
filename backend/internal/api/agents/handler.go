// Package agents exposes agent administration and the member-safe catalog.
package agents

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/agents"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	"github.com/thomas-illiet/KubeCoder/backend/internal/organizations"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

type UserService interface {
	Current(context.Context, auth.Identity) (users.User, error)
}
type OrganizationService interface {
	GetForUser(context.Context, users.User, string) (organizations.Organization, error)
}
type AgentService interface {
	List(context.Context, users.User, string, *bool, int, int) ([]agents.Agent, int64, error)
	Catalog(context.Context, string, int, int) ([]agents.PublicAgent, int64, error)
	Get(context.Context, users.User, uuid.UUID) (agents.Agent, error)
	Create(context.Context, users.User, agents.Input) (agents.Agent, error)
	Update(context.Context, users.User, uuid.UUID, agents.Input) (agents.Agent, error)
	Delete(context.Context, users.User, uuid.UUID) error
}

type Handler struct {
	logger        *slog.Logger
	verifier      auth.Verifier
	users         UserService
	organizations OrganizationService
	service       AgentService
}

// NewHandler creates the combined administration and catalog agent handler.
func NewHandler(logger *slog.Logger, verifier auth.Verifier, users UserService, organizations OrganizationService, service AgentService) *Handler {
	return &Handler{logger: logger, verifier: verifier, users: users, organizations: organizations, service: service}
}

// RegisterRoutes registers agent administration and member catalog endpoints.
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("GET /api/v1/admin/agents", handler.list)
	mux.HandleFunc("POST /api/v1/admin/agents", handler.create)
	mux.HandleFunc("GET /api/v1/admin/agents/{id}", handler.get)
	mux.HandleFunc("PATCH /api/v1/admin/agents/{id}", handler.update)
	mux.HandleFunc("DELETE /api/v1/admin/agents/{id}", handler.delete)
	mux.HandleFunc("GET /api/v1/organizations/{slug}/agents", handler.catalog)
}

// list returns full agent definitions to platform administrators.
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	page := httpx.ParsePagination(r)
	var active *bool
	if raw := r.URL.Query().Get("active"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			badRequest(w, r, "The active filter is invalid.")
			return
		}
		active = &value
	}
	items, total, err := h.service.List(r.Context(), actor, r.URL.Query().Get("query"), active, page.Limit, page.Offset)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.NewPage(items, total, page))
}

// catalog returns active public agents to organization members.
func (h *Handler) catalog(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	if _, err := h.organizations.GetForUser(r.Context(), actor, r.PathValue("slug")); err != nil {
		if errors.Is(err, organizations.ErrForbidden) {
			httpx.WriteProblem(w, r, http.StatusForbidden, "Forbidden", "You do not have access to this organization.")
			return
		}
		h.internalError(w, r, err)
		return
	}
	page := httpx.ParsePagination(r)
	items, total, err := h.service.Catalog(r.Context(), r.URL.Query().Get("query"), page.Limit, page.Offset)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.NewPage(items, total, page))
}

// get returns one full agent definition.
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

// create decodes and creates an agent definition.
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	var input agents.Input
	if !decode(w, r, &input) {
		return
	}
	item, err := h.service.Create(r.Context(), actor, input)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

// update replaces an agent definition.
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.actorAndID(w, r)
	if !ok {
		return
	}
	var input agents.Input
	if !decode(w, r, &input) {
		return
	}
	item, err := h.service.Update(r.Context(), actor, id, input)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

// delete removes an unreferenced agent definition.
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

// actorAndID authenticates the actor and parses an agent identifier.
func (h *Handler) actorAndID(w http.ResponseWriter, r *http.Request) (users.User, uuid.UUID, bool) {
	actor, ok := h.actor(w, r)
	if !ok {
		return users.User{}, uuid.Nil, false
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		badRequest(w, r, "The agent ID is invalid.")
		return users.User{}, uuid.Nil, false
	}
	return actor, id, true
}

// writeError maps known agent errors to problem responses.
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, agents.ErrForbidden):
		httpx.WriteProblem(w, r, http.StatusForbidden, "Forbidden", "Platform administrator access is required.")
	case errors.Is(err, agents.ErrNotFound):
		httpx.WriteProblem(w, r, http.StatusNotFound, "Not Found", "The agent was not found.")
	case errors.Is(err, agents.ErrConflict):
		httpx.WriteProblem(w, r, http.StatusConflict, "Conflict", "The agent name already exists or the agent is assigned to a repository.")
	case errors.Is(err, agents.ErrInvalid):
		badRequest(w, r, "The agent definition is invalid.")
	default:
		h.internalError(w, r, err)
	}
}

// internalError records and hides an unexpected agent failure.
func (h *Handler) internalError(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.ErrorContext(r.Context(), "agent request failed", "request_id", httpx.RequestID(r.Context()), "error", err)
	httpx.WriteProblem(w, r, http.StatusInternalServerError, "Internal Server Error", "The agent operation failed.")
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

// badRequest writes a standard validation problem.
func badRequest(w http.ResponseWriter, r *http.Request, detail string) {
	httpx.WriteProblem(w, r, http.StatusBadRequest, "Invalid request", detail)
}
