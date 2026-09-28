package organizations

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// OrganizationOrder identifies the supported administration list sort columns.
type OrganizationOrder string

const (
	OrganizationOrderName            OrganizationOrder = "name"
	OrganizationOrderCreatedAt       OrganizationOrder = "created_at"
	OrganizationOrderMemberCount     OrganizationOrder = "member_count"
	OrganizationOrderRepositoryCount OrganizationOrder = "repository_count"
)

// MemberOrder identifies the supported organization member sort columns.
type MemberOrder string

const (
	MemberOrderDisplayName MemberOrder = "display_name"
	MemberOrderUsername    MemberOrder = "username"
	MemberOrderEmail       MemberOrder = "email"
	MemberOrderJoinedAt    MemberOrder = "joined_at"
)

// OrderDirection identifies the supported list sort directions.
type OrderDirection string

const (
	OrderAscending  OrderDirection = "asc"
	OrderDescending OrderDirection = "desc"
)

// SSHKeyGenerator creates an encrypted SSH identity for a new organization.
type SSHKeyGenerator interface {
	Generate(uuid.UUID) (models.OrganizationSSHKey, error)
}

// Service implements organization access and administration rules.
type Service struct {
	repository *Repository
	sshKeys    SSHKeyGenerator
}

// NewService creates an organization service.
func NewService(repository *Repository, sshKeys SSHKeyGenerator) *Service {
	return &Service{repository: repository, sshKeys: sshKeys}
}

// ListForUser returns organizations accessible to the actor.
func (s *Service) ListForUser(ctx context.Context, actor users.User, query string, limit, offset int) ([]Organization, int64, error) {
	return s.repository.ListForUser(ctx, actor.ID, strings.TrimSpace(query), normalizeLimit(limit), max(offset, 0))
}

// GetForUser returns an organization only when the actor is a member.
func (s *Service) GetForUser(ctx context.Context, actor users.User, slug string) (Organization, error) {
	return s.repository.FindForUserBySlug(ctx, actor.ID, slug)
}

// SetPreferred makes an accessible organization the actor preference.
func (s *Service) SetPreferred(ctx context.Context, actor users.User, slug string) (Organization, error) {
	return s.repository.SetPreferred(ctx, actor.ID, slug)
}

// List returns all organizations for a platform administrator.
func (s *Service) List(ctx context.Context, actor users.User, query string, limit, offset int, orderBy, orderDirection string) ([]OrganizationSummary, int64, error) {
	if !actor.IsAdmin {
		return nil, 0, ErrForbidden
	}
	return s.repository.List(
		ctx,
		strings.TrimSpace(query),
		normalizeLimit(limit),
		max(offset, 0),
		normalizeOrganizationOrder(orderBy),
		normalizeOrderDirection(orderDirection),
	)
}

// CountRepositories returns the platform repository inventory to an administrator.
func (s *Service) CountRepositories(ctx context.Context, actor users.User) (int64, error) {
	if !actor.IsAdmin {
		return 0, ErrForbidden
	}
	return s.repository.CountRepositories(ctx)
}

// Create validates and creates an organization for a platform administrator.
func (s *Service) Create(ctx context.Context, actor users.User, name, slug string) (Organization, error) {
	if !actor.IsAdmin {
		return Organization{}, ErrForbidden
	}
	name, slug = strings.TrimSpace(name), strings.ToLower(strings.TrimSpace(slug))
	if name == "" || len(name) > 120 || len(slug) > 63 || !slugPattern.MatchString(slug) {
		return Organization{}, ErrInvalid
	}
	result := Organization{ID: uuid.New(), Name: name, Slug: slug}
	key, err := s.sshKeys.Generate(result.ID)
	if err != nil {
		return Organization{}, err
	}
	if err := s.repository.CreateWithSSHKey(ctx, &result, &key); err != nil {
		return Organization{}, err
	}
	return result, nil
}

// Get returns an organization by ID for a platform administrator.
func (s *Service) Get(ctx context.Context, actor users.User, id uuid.UUID) (Organization, error) {
	if !actor.IsAdmin {
		return Organization{}, ErrForbidden
	}
	return s.repository.FindByID(ctx, id)
}

// Rename updates an organization display name for a platform administrator.
func (s *Service) Rename(ctx context.Context, actor users.User, id uuid.UUID, name string) (Organization, error) {
	if !actor.IsAdmin {
		return Organization{}, ErrForbidden
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 120 {
		return Organization{}, ErrInvalid
	}
	return s.repository.Rename(ctx, id, name)
}

// Delete permanently removes an organization for a platform administrator.
func (s *Service) Delete(ctx context.Context, actor users.User, id uuid.UUID) error {
	if !actor.IsAdmin {
		return ErrForbidden
	}
	return s.repository.Delete(ctx, id)
}

// ListMembers returns organization members for a platform administrator.
func (s *Service) ListMembers(ctx context.Context, actor users.User, id uuid.UUID, query string, limit, offset int, orderBy, orderDirection string) ([]Member, int64, error) {
	if !actor.IsAdmin {
		return nil, 0, ErrForbidden
	}
	if _, err := s.repository.FindByID(ctx, id); err != nil {
		return nil, 0, err
	}
	return s.repository.ListMembers(
		ctx,
		id,
		strings.TrimSpace(query),
		normalizeLimit(limit),
		max(offset, 0),
		normalizeMemberOrder(orderBy),
		normalizeOrderDirection(orderDirection),
	)
}

// AddMember grants organization access to a provisioned user.
func (s *Service) AddMember(ctx context.Context, actor users.User, organizationID, userID uuid.UUID) error {
	if !actor.IsAdmin {
		return ErrForbidden
	}
	if _, err := s.repository.FindByID(ctx, organizationID); err != nil {
		return err
	}
	return s.repository.AddMember(ctx, organizationID, userID)
}

// RemoveMember revokes organization access from a user.
func (s *Service) RemoveMember(ctx context.Context, actor users.User, organizationID, userID uuid.UUID) error {
	if !actor.IsAdmin {
		return ErrForbidden
	}
	return s.repository.RemoveMember(ctx, organizationID, userID)
}

// normalizeLimit applies the public pagination bounds.
func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

// normalizeOrganizationOrder restricts client input to supported columns.
func normalizeOrganizationOrder(value string) OrganizationOrder {
	switch OrganizationOrder(value) {
	case OrganizationOrderCreatedAt, OrganizationOrderMemberCount, OrganizationOrderRepositoryCount:
		return OrganizationOrder(value)
	default:
		return OrganizationOrderName
	}
}

// normalizeOrderDirection restricts client input to supported directions.
func normalizeOrderDirection(value string) OrderDirection {
	if OrderDirection(value) == OrderDescending {
		return OrderDescending
	}
	return OrderAscending
}

// normalizeMemberOrder restricts client input to supported member columns.
func normalizeMemberOrder(value string) MemberOrder {
	switch MemberOrder(value) {
	case MemberOrderUsername, MemberOrderEmail, MemberOrderJoinedAt:
		return MemberOrder(value)
	default:
		return MemberOrderDisplayName
	}
}
