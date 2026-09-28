// Package server assembles the HTTP server from independent API categories.
package server

import (
	"log/slog"
	"net/http"

	"github.com/thomas-illiet/KubeCoder/backend/internal/api/health"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	userapi "github.com/thomas-illiet/KubeCoder/backend/internal/api/users"
)

// New builds the complete HTTP handler and registers each API category.
func New(logger *slog.Logger, healthHandler *health.Handler, userHandler *userapi.Handler, allowedOrigins []string) http.Handler {
	mux := http.NewServeMux()
	health.RegisterRoutes(mux, healthHandler)
	userapi.RegisterRoutes(mux, userHandler)
	registerDocumentationRoutes(mux)
	return httpx.Middleware(logger, httpx.CORS(allowedOrigins, mux))
}
