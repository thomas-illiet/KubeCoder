// Package organizations exposes authenticated organization operations over HTTP.
package organizations

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	domain "github.com/thomas-illiet/KubeCoder/backend/internal/organizations"
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

// Handler serves the authenticated organization API category.
type Handler struct {
	logger   *slog.Logger
	verifier auth.Verifier
	users    UserService
	service  Service
}

// NewHandler creates an authenticated organization API handler.
func NewHandler(logger *slog.Logger, verifier auth.Verifier, users UserService, service Service) *Handler {
	return &Handler{logger: logger, verifier: verifier, users: users, service: service}
}

// RegisterRoutes registers endpoints for the organization API category.
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("GET /api/v1/organizations", handler.list)
	mux.HandleFunc("GET /api/v1/organizations/{slug}", handler.get)
	mux.HandleFunc("PUT /api/v1/organizations/{slug}/preferred", handler.prefer)
}

type page struct {
	Items  []domain.Organization `json:"items"`
	Total  int64                 `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

// list returns organizations accessible to the current user.
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	limit, offset := pagination(r)
	items, total, err := h.service.ListForUser(r.Context(), actor, r.URL.Query().Get("query"), limit, offset)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if items == nil {
		items = make([]domain.Organization, 0)
	}
	httpx.WriteJSON(w, http.StatusOK, page{Items: items, Total: total, Limit: limit, Offset: offset})
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

// pagination parses bounded offset pagination parameters.
func pagination(r *http.Request) (int, int) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
