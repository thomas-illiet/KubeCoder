// Package users exposes user operations over HTTP.
package users

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	domain "github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

// Service defines user operations required by the HTTP handler.
type Service interface {
	Current(context.Context, auth.Identity) (domain.User, error)
}

// Handler serves the user API category.
type Handler struct {
	logger   *slog.Logger
	verifier auth.Verifier
	service  Service
}

// NewHandler creates a user API handler.
func NewHandler(logger *slog.Logger, verifier auth.Verifier, service Service) *Handler {
	return &Handler{logger: logger, verifier: verifier, service: service}
}

// RegisterRoutes registers endpoints for the user API category.
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("GET /api/v1/users/me", handler.currentUser)
}

// currentUser authenticates, provisions, and returns the current application user.
func (h *Handler) currentUser(w http.ResponseWriter, r *http.Request) {
	identity, err := auth.FromRequest(r.Context(), h.verifier, r)
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusUnauthorized, "Unauthorized", "A valid bearer token is required.")
		return
	}
	current, err := h.service.Current(r.Context(), identity)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "user persistence failed", "request_id", httpx.RequestID(r.Context()), "error", err)
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "Internal Server Error", "The user could not be loaded.")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, current)
}
