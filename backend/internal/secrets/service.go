package secrets

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

var variablePattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

type Service struct {
	repository   *Repository
	protector    *protector
	expiringSoon time.Duration
	now          func() time.Time
}

// NewService creates the secret domain service.
func NewService(repository *Repository, key []byte, expiringSoon time.Duration) (*Service, error) {
	p, err := newProtector(key)
	if err != nil {
		return nil, err
	}
	return &Service{repository: repository, protector: p, expiringSoon: expiringSoon, now: time.Now}, nil
}

// ListPlatform returns administrator-only platform metadata.
func (s *Service) ListPlatform(ctx context.Context, actor users.User, query, status, sortBy, sortOrder string, limit, offset int) ([]View, int64, error) {
	if !actor.IsAdmin {
		return nil, 0, ErrForbidden
	}
	return s.list(ctx, nil, false, query, "", status, sortBy, sortOrder, limit, offset)
}

// ListOrganizationAdmin returns tenant metadata to a platform administrator.
func (s *Service) ListOrganizationAdmin(ctx context.Context, actor users.User, organizationID uuid.UUID, query, scope, status, sortBy, sortOrder string, limit, offset int) ([]View, int64, error) {
	if !actor.IsAdmin {
		return nil, 0, ErrForbidden
	}
	return s.list(ctx, &organizationID, false, query, scope, status, sortBy, sortOrder, limit, offset)
}

// ListOrganization returns the same sanitized projection to an authorized member.
func (s *Service) ListOrganization(ctx context.Context, organizationID uuid.UUID, query, scope, status, sortBy, sortOrder string, limit, offset int) ([]View, int64, error) {
	return s.list(ctx, &organizationID, true, query, scope, status, sortBy, sortOrder, limit, offset)
}

// Targets returns administrator binding candidates.
func (s *Service) Targets(ctx context.Context, actor users.User, organizationID *uuid.UUID) ([]Target, error) {
	if err := s.authorizeMutation(ctx, actor, organizationID); err != nil {
		return nil, err
	}
	return s.repository.Targets(ctx, organizationID)
}

// list applies common filters and constructs public projections.
func (s *Service) list(ctx context.Context, organizationID *uuid.UUID, includePlatform bool, query, scope, status, sortBy, sortOrder string, limit, offset int) ([]View, int64, error) {
	sortBy, sortOrder = normalizeSort(sortBy, sortOrder)
	if !validFilters(organizationID, includePlatform, scope, status) || sortBy == "" {
		return nil, 0, ErrInvalid
	}
	items, total, err := s.repository.List(ctx, organizationID, includePlatform, strings.TrimSpace(query), scope, status, sortBy, sortOrder, s.now().Add(s.expiringSoon), normalizeLimit(limit), max(offset, 0))
	if err != nil {
		return nil, 0, err
	}
	ids := make([]uuid.UUID, len(items))
	for index := range items {
		ids[index] = items[index].ID
	}
	bindingsBySecret, err := s.repository.BindingsBySecretIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	views := make([]View, 0, len(items))
	for _, item := range items {
		views = append(views, s.viewWithBindings(item, bindingsBySecret[item.ID]))
	}
	return views, total, nil
}

// normalizeSort applies defaults and rejects unsupported ordering parameters.
func normalizeSort(sortBy, sortOrder string) (string, string) {
	if sortBy == "" {
		sortBy = "variable_name"
	}
	if sortOrder == "" {
		sortOrder = SortAscending
	}
	if _, ok := allowedSortFields[sortBy]; !ok || (sortOrder != SortAscending && sortOrder != SortDescending) {
		return "", ""
	}
	return sortBy, sortOrder
}

// CreatePlatform creates one platform-owned secret.
func (s *Service) CreatePlatform(ctx context.Context, actor users.User, input Input) (View, error) {
	return s.create(ctx, actor, nil, input)
}

// CreateOrganization creates one organization-owned secret.
func (s *Service) CreateOrganization(ctx context.Context, actor users.User, organizationID uuid.UUID, input Input) (View, error) {
	return s.create(ctx, actor, &organizationID, input)
}

// create validates, protects and persists a write-only value.
func (s *Service) create(ctx context.Context, actor users.User, organizationID *uuid.UUID, input Input) (View, error) {
	if err := s.authorizeMutation(ctx, actor, organizationID); err != nil {
		return View{}, err
	}
	item, err := s.validatedSecret(ctx, organizationID, input)
	if err != nil {
		return View{}, err
	}
	item.ID = uuid.New()
	now := s.now().UTC()
	item.CreatedAt, item.UpdatedAt, item.ValueReplacedAt = now, now, now
	item.EncryptedValue, item.Nonce, item.Fingerprint, err = s.protector.protect(item.ID, item.Scope, input.Value)
	if err != nil {
		return View{}, err
	}
	item.EncryptionVersion = encryptionVersion
	if err := s.repository.Create(ctx, &item); err != nil {
		return View{}, err
	}
	return s.view(ctx, item)
}

// Replace destructively overwrites the current value.
func (s *Service) Replace(ctx context.Context, actor users.User, id uuid.UUID, organizationID *uuid.UUID, input ReplaceInput) (View, error) {
	if err := s.authorizeMutation(ctx, actor, organizationID); err != nil {
		return View{}, err
	}
	if strings.TrimSpace(input.Value) == "" || len(input.Value) > 65536 || (input.ExpiresAt != nil && !input.ExpiresAt.After(s.now())) {
		return View{}, ErrInvalid
	}
	item, err := s.repository.Get(ctx, id, organizationID)
	if err != nil {
		return View{}, err
	}
	item.EncryptedValue, item.Nonce, item.Fingerprint, err = s.protector.protect(item.ID, item.Scope, input.Value)
	if err != nil {
		return View{}, err
	}
	now := s.now().UTC()
	item.EncryptionVersion, item.ExpiresAt, item.ValueReplacedAt, item.UpdatedAt = encryptionVersion, input.ExpiresAt, now, now
	if err := s.repository.Replace(ctx, item); err != nil {
		return View{}, err
	}
	return s.view(ctx, item)
}

// AddBinding validates tenancy and assigns the current value to a target.
func (s *Service) AddBinding(ctx context.Context, actor users.User, id uuid.UUID, organizationID *uuid.UUID, input BindingInput) (BindingView, error) {
	if err := s.authorizeMutation(ctx, actor, organizationID); err != nil {
		return BindingView{}, err
	}
	if _, err := s.repository.Get(ctx, id, organizationID); err != nil {
		return BindingView{}, err
	}
	binding := Binding{ID: uuid.New(), SecretID: id, TargetType: input.TargetType, CreatedAt: s.now().UTC()}
	switch input.TargetType {
	case TargetAgent:
		exists, err := s.repository.AgentExists(ctx, input.TargetID)
		if err != nil {
			return BindingView{}, err
		}
		if !exists {
			return BindingView{}, ErrInvalid
		}
		binding.AgentID = &input.TargetID
	case TargetRepository:
		return BindingView{}, ErrInvalid
	default:
		return BindingView{}, ErrInvalid
	}
	if err := s.repository.AddBinding(ctx, &binding); err != nil {
		return BindingView{}, err
	}
	bindings, err := s.repository.Bindings(ctx, id)
	if err != nil {
		return BindingView{}, err
	}
	for _, candidate := range bindings {
		if candidate.ID == binding.ID {
			return bindingView(candidate), nil
		}
	}
	return BindingView{}, ErrNotFound
}

// RemoveBinding deletes one explicit target assignment.
func (s *Service) RemoveBinding(ctx context.Context, actor users.User, id, bindingID uuid.UUID, organizationID *uuid.UUID) error {
	if err := s.authorizeMutation(ctx, actor, organizationID); err != nil {
		return err
	}
	if _, err := s.repository.Get(ctx, id, organizationID); err != nil {
		return err
	}
	binding, err := s.repository.Binding(ctx, id, bindingID)
	if err != nil {
		return err
	}
	if binding.TargetType == TargetRepository {
		return ErrInvalid
	}
	return s.repository.RemoveBinding(ctx, id, bindingID)
}

// validatedSecret normalizes input and enforces scope ownership.
func (s *Service) validatedSecret(ctx context.Context, organizationID *uuid.UUID, input Input) (Secret, error) {
	input.VariableName = strings.ToUpper(strings.Join(strings.Fields(input.VariableName), "_"))
	input.Description = strings.TrimSpace(input.Description)
	if !variablePattern.MatchString(input.VariableName) || len(input.VariableName) > 128 || len(input.Description) > 1000 || input.Value == "" || len(input.Value) > 65536 || (input.ExpiresAt != nil && !input.ExpiresAt.After(s.now())) {
		return Secret{}, ErrInvalid
	}
	item := Secret{Scope: input.Scope, OrganizationID: organizationID, VariableName: input.VariableName, Description: input.Description, ExpiresAt: input.ExpiresAt}
	if organizationID != nil {
		exists, err := s.repository.PlatformVariableExists(ctx, input.VariableName)
		if err != nil {
			return Secret{}, err
		}
		if exists {
			return Secret{}, ErrConflict
		}
	}
	if organizationID == nil {
		if input.Scope != ScopePlatform {
			return Secret{}, ErrInvalid
		}
		exists, err := s.repository.OrganizationVariableExists(ctx, input.VariableName)
		if err != nil {
			return Secret{}, err
		}
		if exists {
			return Secret{}, ErrConflict
		}
		return item, nil
	}
	if input.Scope != ScopeOrganization {
		return Secret{}, ErrInvalid
	}
	return item, nil
}

// view constructs the only externally serializable secret representation.
func (s *Service) view(ctx context.Context, item Secret) (View, error) {
	bindings, err := s.repository.Bindings(ctx, item.ID)
	if err != nil {
		return View{}, err
	}
	return s.viewWithBindings(item, bindings), nil
}

// viewWithBindings constructs a public projection from already-loaded bindings.
func (s *Service) viewWithBindings(item Secret, bindings []Binding) View {
	result := View{ID: item.ID, Scope: item.Scope, OrganizationID: item.OrganizationID, VariableName: item.VariableName, Description: item.Description, ExpiresAt: item.ExpiresAt, ValueReplacedAt: item.ValueReplacedAt, Status: s.status(item), Bindings: make([]BindingView, 0, len(bindings)), BindingCount: len(bindings), CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
	for _, binding := range bindings {
		result.Bindings = append(result.Bindings, bindingView(binding))
	}
	return result
}

// status calculates the current lifecycle state.
func (s *Service) status(item Secret) string {
	now := s.now()
	if item.ExpiresAt == nil {
		return StatusActive
	}
	if !item.ExpiresAt.After(now) {
		return StatusExpired
	}
	if !item.ExpiresAt.After(now.Add(s.expiringSoon)) {
		return StatusExpiringSoon
	}
	return StatusActive
}

// bindingView removes persistence-only fields from a binding.
func bindingView(item Binding) BindingView {
	result := BindingView{ID: item.ID, TargetType: item.TargetType}
	if item.AgentID != nil {
		result.TargetID = *item.AgentID
		if item.Agent != nil {
			result.TargetName = item.Agent.Name
		}
	} else if item.RepositoryID != nil {
		result.TargetID = *item.RepositoryID
		if item.Repository != nil {
			result.TargetName = item.Repository.Name
		}
	}
	return result
}

// validFilters rejects unsupported public filter combinations.
func validFilters(organizationID *uuid.UUID, includePlatform bool, scope, status string) bool {
	if scope != "" && (organizationID == nil || (scope != ScopeOrganization && (!includePlatform || scope != ScopePlatform))) {
		return false
	}
	return status == "" || status == StatusActive || status == StatusExpiringSoon || status == StatusExpired
}

// authorizeMutation permits platform administration or explicit tenant membership.
func (s *Service) authorizeMutation(ctx context.Context, actor users.User, organizationID *uuid.UUID) error {
	if organizationID == nil {
		if actor.IsAdmin {
			return nil
		}
		return ErrForbidden
	}
	if actor.IsAdmin {
		return nil
	}
	belongs, err := s.repository.UserBelongsToOrganization(ctx, actor.ID, *organizationID)
	if err != nil {
		return err
	}
	if !belongs {
		return ErrForbidden
	}
	return nil
}

// normalizeLimit applies public pagination bounds.
func normalizeLimit(value int) int {
	if value <= 0 {
		return 20
	}
	if value > 100 {
		return 100
	}
	return value
}
