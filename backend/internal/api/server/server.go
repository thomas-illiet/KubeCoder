// Package server assembles the HTTP server from independent API categories.
package server

import (
	"log/slog"
	"net/http"

	adminorganizations "github.com/thomas-illiet/KubeCoder/backend/internal/api/admin/organizations"
	agentapi "github.com/thomas-illiet/KubeCoder/backend/internal/api/agents"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/health"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
	organizationapi "github.com/thomas-illiet/KubeCoder/backend/internal/api/organizations"
	repositoryapi "github.com/thomas-illiet/KubeCoder/backend/internal/api/repositories"
	secretapi "github.com/thomas-illiet/KubeCoder/backend/internal/api/secrets"
	userapi "github.com/thomas-illiet/KubeCoder/backend/internal/api/users"
)

// New builds the complete HTTP handler and registers each API category.
func New(logger *slog.Logger, healthHandler *health.Handler, userHandler *userapi.Handler, organizationHandler *organizationapi.Handler, adminOrganizationHandler *adminorganizations.Handler, agentHandler *agentapi.Handler, repositoryHandler *repositoryapi.Handler, secretHandler *secretapi.Handler, allowedOrigins []string) http.Handler {
	mux := http.NewServeMux()
	health.RegisterRoutes(mux, healthHandler)
	userapi.RegisterRoutes(mux, userHandler)
	organizationapi.RegisterRoutes(mux, organizationHandler)
	adminorganizations.RegisterRoutes(mux, adminOrganizationHandler)
	agentapi.RegisterRoutes(mux, agentHandler)
	repositoryapi.RegisterRoutes(mux, repositoryHandler)
	secretapi.RegisterRoutes(mux, secretHandler)
	registerDocumentationRoutes(mux)
	return httpx.Middleware(logger, httpx.CORS(allowedOrigins, mux))
}
