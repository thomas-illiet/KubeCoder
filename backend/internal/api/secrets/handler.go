// Package secrets exposes platform-managed, write-only secret APIs.
package secrets

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
	"github.com/thomas-illiet/KubeCoder/backend/internal/organizations"
	domain "github.com/thomas-illiet/KubeCoder/backend/internal/secrets"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

type UserService interface {
	Current(context.Context, auth.Identity) (users.User, error)
}
type OrganizationService interface {
	Get(context.Context, users.User, uuid.UUID) (organizations.Organization, error)
	GetForUser(context.Context, users.User, string) (organizations.Organization, error)
}
type SecretService interface {
	ListPlatform(context.Context, users.User, string, string, string, string, int, int) ([]domain.View, int64, error)
	ListOrganizationAdmin(context.Context, users.User, uuid.UUID, string, string, string, string, string, int, int) ([]domain.View, int64, error)
	ListOrganization(context.Context, uuid.UUID, string, string, string, string, string, int, int) ([]domain.View, int64, error)
	CreatePlatform(context.Context, users.User, domain.Input) (domain.View, error)
	CreateOrganization(context.Context, users.User, uuid.UUID, domain.Input) (domain.View, error)
	Replace(context.Context, users.User, uuid.UUID, *uuid.UUID, domain.ReplaceInput) (domain.View, error)
	AddBinding(context.Context, users.User, uuid.UUID, *uuid.UUID, domain.BindingInput) (domain.BindingView, error)
	RemoveBinding(context.Context, users.User, uuid.UUID, uuid.UUID, *uuid.UUID) error
	Targets(context.Context, users.User, *uuid.UUID) ([]domain.Target, error)
}

type Handler struct {
	logger        *slog.Logger
	verifier      auth.Verifier
	users         UserService
	organizations OrganizationService
	service       SecretService
}

// NewHandler implements the corresponding sanitized secret HTTP operation.
func NewHandler(logger *slog.Logger, verifier auth.Verifier, users UserService, organizations OrganizationService, service SecretService) *Handler {
	return &Handler{logger: logger, verifier: verifier, users: users, organizations: organizations, service: service}
}

// RegisterRoutes implements the corresponding sanitized secret HTTP operation.
func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /api/v1/admin/secrets", h.listPlatform)
	mux.HandleFunc("POST /api/v1/admin/secrets", h.createPlatform)
	mux.HandleFunc("POST /api/v1/admin/secrets/{id}/replace", h.replacePlatform)
	mux.HandleFunc("POST /api/v1/admin/secrets/{id}/bindings", h.addPlatformBinding)
	mux.HandleFunc("GET /api/v1/admin/secrets/targets", h.platformTargets)
	mux.HandleFunc("DELETE /api/v1/admin/secrets/{id}/bindings/{bindingID}", h.removePlatformBinding)
	mux.HandleFunc("GET /api/v1/admin/organizations/{organizationID}/secrets", h.listOrganizationAdmin)
	mux.HandleFunc("POST /api/v1/admin/organizations/{organizationID}/secrets", h.createOrganization)
	mux.HandleFunc("POST /api/v1/admin/organizations/{organizationID}/secrets/{id}/replace", h.replaceOrganization)
	mux.HandleFunc("POST /api/v1/admin/organizations/{organizationID}/secrets/{id}/bindings", h.addOrganizationBinding)
	mux.HandleFunc("GET /api/v1/admin/organizations/{organizationID}/secrets/targets", h.organizationTargets)
	mux.HandleFunc("DELETE /api/v1/admin/organizations/{organizationID}/secrets/{id}/bindings/{bindingID}", h.removeOrganizationBinding)
	mux.HandleFunc("GET /api/v1/organizations/{slug}/secrets", h.listOrganizationReadonly)
	mux.HandleFunc("POST /api/v1/organizations/{slug}/secrets", h.createOrganizationMember)
	mux.HandleFunc("GET /api/v1/organizations/{slug}/secrets/targets", h.organizationMemberTargets)
	mux.HandleFunc("POST /api/v1/organizations/{slug}/secrets/{id}/replace", h.replaceOrganizationMember)
	mux.HandleFunc("POST /api/v1/organizations/{slug}/secrets/{id}/bindings", h.addOrganizationMemberBinding)
	mux.HandleFunc("DELETE /api/v1/organizations/{slug}/secrets/{id}/bindings/{bindingID}", h.removeOrganizationMemberBinding)
}

// listPlatform implements the corresponding sanitized secret HTTP operation.
func (h *Handler) listPlatform(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	p := httpx.ParsePagination(r)
	items, total, err := h.service.ListPlatform(r.Context(), actor, r.URL.Query().Get("query"), r.URL.Query().Get("status"), r.URL.Query().Get("sort_by"), r.URL.Query().Get("sort_order"), p.Limit, p.Offset)
	h.page(w, r, items, total, p, err)
}

// createPlatform implements the corresponding sanitized secret HTTP operation.
func (h *Handler) createPlatform(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	var input domain.Input
	if !decode(w, r, &input) {
		return
	}
	item, err := h.service.CreatePlatform(r.Context(), actor, input)
	h.item(w, r, item, http.StatusCreated, err)
}

// replacePlatform implements the corresponding sanitized secret HTTP operation.
func (h *Handler) replacePlatform(w http.ResponseWriter, r *http.Request) { h.replace(w, r, nil) }

// addPlatformBinding implements the corresponding sanitized secret HTTP operation.
func (h *Handler) addPlatformBinding(w http.ResponseWriter, r *http.Request) { h.addBinding(w, r, nil) }

// removePlatformBinding implements the corresponding sanitized secret HTTP operation.
func (h *Handler) removePlatformBinding(w http.ResponseWriter, r *http.Request) {
	h.removeBinding(w, r, nil)
}

// platformTargets implements the corresponding sanitized secret HTTP operation.
func (h *Handler) platformTargets(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	items, err := h.service.Targets(r.Context(), actor, nil)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

// organizationAdmin implements the corresponding sanitized secret HTTP operation.
func (h *Handler) organizationAdmin(w http.ResponseWriter, r *http.Request) (users.User, uuid.UUID, bool) {
	actor, ok := h.actor(w, r)
	if !ok {
		return users.User{}, uuid.Nil, false
	}
	id, err := uuid.Parse(r.PathValue("organizationID"))
	if err != nil {
		badRequest(w, r, "The organization ID is invalid.")
		return users.User{}, uuid.Nil, false
	}
	if _, err = h.organizations.Get(r.Context(), actor, id); err != nil {
		h.writeError(w, r, err)
		return users.User{}, uuid.Nil, false
	}
	return actor, id, true
}

// listOrganizationAdmin implements the corresponding sanitized secret HTTP operation.
func (h *Handler) listOrganizationAdmin(w http.ResponseWriter, r *http.Request) {
	actor, oid, ok := h.organizationAdmin(w, r)
	if !ok {
		return
	}
	p := httpx.ParsePagination(r)
	items, total, err := h.service.ListOrganizationAdmin(r.Context(), actor, oid, r.URL.Query().Get("query"), r.URL.Query().Get("scope"), r.URL.Query().Get("status"), r.URL.Query().Get("sort_by"), r.URL.Query().Get("sort_order"), p.Limit, p.Offset)
	h.page(w, r, items, total, p, err)
}

// createOrganization implements the corresponding sanitized secret HTTP operation.
func (h *Handler) createOrganization(w http.ResponseWriter, r *http.Request) {
	actor, oid, ok := h.organizationAdmin(w, r)
	if !ok {
		return
	}
	var input domain.Input
	if !decode(w, r, &input) {
		return
	}
	item, err := h.service.CreateOrganization(r.Context(), actor, oid, input)
	h.item(w, r, item, http.StatusCreated, err)
}

// replaceOrganization implements the corresponding sanitized secret HTTP operation.
func (h *Handler) replaceOrganization(w http.ResponseWriter, r *http.Request) {
	_, oid, ok := h.organizationAdmin(w, r)
	if !ok {
		return
	}
	h.replace(w, r, &oid)
}

// addOrganizationBinding implements the corresponding sanitized secret HTTP operation.
func (h *Handler) addOrganizationBinding(w http.ResponseWriter, r *http.Request) {
	_, oid, ok := h.organizationAdmin(w, r)
	if !ok {
		return
	}
	h.addBinding(w, r, &oid)
}

// removeOrganizationBinding implements the corresponding sanitized secret HTTP operation.
func (h *Handler) removeOrganizationBinding(w http.ResponseWriter, r *http.Request) {
	_, oid, ok := h.organizationAdmin(w, r)
	if !ok {
		return
	}
	h.removeBinding(w, r, &oid)
}

// organizationTargets implements the corresponding sanitized secret HTTP operation.
func (h *Handler) organizationTargets(w http.ResponseWriter, r *http.Request) {
	actor, oid, ok := h.organizationAdmin(w, r)
	if !ok {
		return
	}
	items, err := h.service.Targets(r.Context(), actor, &oid)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

// listOrganizationReadonly implements the corresponding sanitized secret HTTP operation.
func (h *Handler) listOrganizationReadonly(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	organization, err := h.organizations.GetForUser(r.Context(), actor, r.PathValue("slug"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	p := httpx.ParsePagination(r)
	items, total, err := h.service.ListOrganization(r.Context(), organization.ID, r.URL.Query().Get("query"), r.URL.Query().Get("scope"), r.URL.Query().Get("status"), r.URL.Query().Get("sort_by"), r.URL.Query().Get("sort_order"), p.Limit, p.Offset)
	h.page(w, r, items, total, p, err)
}

// organizationMember resolves an organization through explicit membership.
func (h *Handler) organizationMember(w http.ResponseWriter, r *http.Request) (users.User, uuid.UUID, bool) {
	actor, ok := h.actor(w, r)
	if !ok {
		return users.User{}, uuid.Nil, false
	}
	organization, err := h.organizations.GetForUser(r.Context(), actor, r.PathValue("slug"))
	if err != nil {
		h.writeError(w, r, err)
		return users.User{}, uuid.Nil, false
	}
	return actor, organization.ID, true
}

// createOrganizationMember creates a tenant-owned secret for an organization member.
func (h *Handler) createOrganizationMember(w http.ResponseWriter, r *http.Request) {
	actor, oid, ok := h.organizationMember(w, r)
	if !ok {
		return
	}
	var input domain.Input
	if !decode(w, r, &input) {
		return
	}
	item, err := h.service.CreateOrganization(r.Context(), actor, oid, input)
	h.item(w, r, item, http.StatusCreated, err)
}

// organizationMemberTargets lists targets constrained to the member organization.
func (h *Handler) organizationMemberTargets(w http.ResponseWriter, r *http.Request) {
	actor, oid, ok := h.organizationMember(w, r)
	if !ok {
		return
	}
	items, err := h.service.Targets(r.Context(), actor, &oid)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

// replaceOrganizationMember replaces a tenant secret value.
func (h *Handler) replaceOrganizationMember(w http.ResponseWriter, r *http.Request) {
	_, oid, ok := h.organizationMember(w, r)
	if ok {
		h.replace(w, r, &oid)
	}
}

// addOrganizationMemberBinding adds a tenant-constrained binding.
func (h *Handler) addOrganizationMemberBinding(w http.ResponseWriter, r *http.Request) {
	_, oid, ok := h.organizationMember(w, r)
	if ok {
		h.addBinding(w, r, &oid)
	}
}

// removeOrganizationMemberBinding removes a tenant-constrained binding.
func (h *Handler) removeOrganizationMemberBinding(w http.ResponseWriter, r *http.Request) {
	_, oid, ok := h.organizationMember(w, r)
	if ok {
		h.removeBinding(w, r, &oid)
	}
}

// replace implements the corresponding sanitized secret HTTP operation.
func (h *Handler) replace(w http.ResponseWriter, r *http.Request, oid *uuid.UUID) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	var input domain.ReplaceInput
	if !decode(w, r, &input) {
		return
	}
	item, err := h.service.Replace(r.Context(), actor, id, oid, input)
	h.item(w, r, item, http.StatusOK, err)
}

// addBinding implements the corresponding sanitized secret HTTP operation.
func (h *Handler) addBinding(w http.ResponseWriter, r *http.Request, oid *uuid.UUID) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	var input domain.BindingInput
	if !decode(w, r, &input) {
		return
	}
	item, err := h.service.AddBinding(r.Context(), actor, id, oid, input)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

// removeBinding implements the corresponding sanitized secret HTTP operation.
func (h *Handler) removeBinding(w http.ResponseWriter, r *http.Request, oid *uuid.UUID) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	bindingID, ok := parseID(w, r, "bindingID")
	if !ok {
		return
	}
	if err := h.service.RemoveBinding(r.Context(), actor, id, bindingID, oid); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// actor implements the corresponding sanitized secret HTTP operation.
func (h *Handler) actor(w http.ResponseWriter, r *http.Request) (users.User, bool) {
	identity, err := auth.FromRequest(r.Context(), h.verifier, r)
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusUnauthorized, "Unauthorized", "A valid bearer token is required.")
		return users.User{}, false
	}
	actor, err := h.users.Current(r.Context(), identity)
	if err != nil {
		h.internal(w, r, err)
		return users.User{}, false
	}
	return actor, true
}

// page implements the corresponding sanitized secret HTTP operation.
func (h *Handler) page(w http.ResponseWriter, r *http.Request, items []domain.View, total int64, p httpx.Pagination, err error) {
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.NewPage(items, total, p))
}

// item implements the corresponding sanitized secret HTTP operation.
func (h *Handler) item(w http.ResponseWriter, r *http.Request, item domain.View, status int, err error) {
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, status, item)
}

// writeError implements the corresponding sanitized secret HTTP operation.
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, organizations.ErrForbidden):
		httpx.WriteProblem(w, r, http.StatusForbidden, "Forbidden", "Platform administrator access is required for this operation.")
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, organizations.ErrNotFound):
		httpx.WriteProblem(w, r, http.StatusNotFound, "Not Found", "The requested resource was not found.")
	case errors.Is(err, domain.ErrConflict):
		httpx.WriteProblem(w, r, http.StatusConflict, "Conflict", "The secret or binding already exists.")
	case errors.Is(err, domain.ErrInvalid), errors.Is(err, organizations.ErrInvalid):
		badRequest(w, r, "The secret request is invalid.")
	default:
		h.internal(w, r, err)
	}
}

// internal implements the corresponding sanitized secret HTTP operation.
func (h *Handler) internal(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.ErrorContext(r.Context(), "secret request failed", "request_id", httpx.RequestID(r.Context()), "error", err)
	httpx.WriteProblem(w, r, http.StatusInternalServerError, "Internal Server Error", "The secret operation failed.")
}

// decode implements the corresponding sanitized secret HTTP operation.
func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		badRequest(w, r, "The JSON request body is invalid.")
		return false
	}
	return true
}

// parseID implements the corresponding sanitized secret HTTP operation.
func parseID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		badRequest(w, r, "The resource ID is invalid.")
		return uuid.Nil, false
	}
	return id, true
}

// badRequest implements the corresponding sanitized secret HTTP operation.
func badRequest(w http.ResponseWriter, r *http.Request, detail string) {
	httpx.WriteProblem(w, r, http.StatusBadRequest, "Invalid request", detail)
}
