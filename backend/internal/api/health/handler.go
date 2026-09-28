// Package health exposes liveness and readiness operations over HTTP.
package health

import (
	"context"
	"net/http"
	"time"

	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
)

// Database checks whether the database accepts requests.
type Database interface {
	PingContext(context.Context) error
}

// Handler serves the health API category.
type Handler struct {
	database Database
	ready    func(context.Context) error
}

// NewHandler creates a health API handler.
func NewHandler(database Database, ready func(context.Context) error) *Handler {
	return &Handler{database: database, ready: ready}
}

// RegisterRoutes registers liveness and readiness endpoints.
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("GET /healthz", handler.health)
	mux.HandleFunc("GET /readyz", handler.readiness)
}

// health reports that the API process is alive.
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readiness verifies database connectivity and the expected schema version.
func (h *Handler) readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.database.PingContext(ctx); err != nil || h.ready(ctx) != nil {
		httpx.WriteProblem(w, r, http.StatusServiceUnavailable, "Service Unavailable", "The API is not ready.")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
