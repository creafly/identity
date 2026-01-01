package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/domain/repository"
	"github.com/google/uuid"
)

var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantAlreadyExists = errors.New("tenant with this slug already exists")
	ErrInvalidSlug         = errors.New("invalid slug format")
	ErrTenantBlocked       = errors.New("tenant is blocked")
)

type TenantService interface {
	Create(ctx context.Context, input CreateTenantInput) (*entity.Tenant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateTenantInput) (*entity.Tenant, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.Tenant, error)
	AddMember(ctx context.Context, tenantID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error
	GetMembers(ctx context.Context, tenantID uuid.UUID) ([]*entity.User, error)
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*entity.Tenant, error)
	IsMember(ctx context.Context, tenantID, userID uuid.UUID) (bool, error)
	BlockTenant(ctx context.Context, id uuid.UUID, reason string, blockedBy uuid.UUID) error
	UnblockTenant(ctx context.Context, id uuid.UUID) error
}

type CreateTenantInput struct {
	Name        string
	DisplayName string
	Slug        string
}

type UpdateTenantInput struct {
	Name        *string
	DisplayName *string
	Slug        *string
	IsActive    *bool
}

type tenantService struct {
	repo repository.TenantRepository
}

func NewTenantService(repo repository.TenantRepository) TenantService {
	return &tenantService{repo: repo}
}

func (s *tenantService) Create(ctx context.Context, input CreateTenantInput) (*entity.Tenant, error) {
	slug := input.Slug
	if slug == "" {
		slug = generateSlug(input.Name)
	}

	if !isValidSlug(slug) {
		return nil, ErrInvalidSlug
	}

	existing, _ := s.repo.GetBySlug(ctx, slug)
	if existing != nil {
		return nil, ErrTenantAlreadyExists
	}

	displayName := input.DisplayName
	if displayName == "" {
		displayName = input.Name
	}

	tenant := &entity.Tenant{
		ID:          uuid.New(),
		Name:        input.Name,
		DisplayName: displayName,
		Slug:        slug,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (s *tenantService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTenantNotFound
	}
	return tenant, nil
}

func (s *tenantService) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	tenant, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrTenantNotFound
	}
	return tenant, nil
}

func (s *tenantService) Update(ctx context.Context, id uuid.UUID, input UpdateTenantInput) (*entity.Tenant, error) {
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	if input.Name != nil {
		tenant.Name = *input.Name
	}
	if input.DisplayName != nil {
		tenant.DisplayName = *input.DisplayName
	}
	if input.Slug != nil {
		if !isValidSlug(*input.Slug) {
			return nil, ErrInvalidSlug
		}
		existing, _ := s.repo.GetBySlug(ctx, *input.Slug)
		if existing != nil && existing.ID != id {
			return nil, ErrTenantAlreadyExists
		}
		tenant.Slug = *input.Slug
	}
	if input.IsActive != nil {
		tenant.IsActive = *input.IsActive
	}
	tenant.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (s *tenantService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrTenantNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *tenantService) List(ctx context.Context, offset, limit int) ([]*entity.Tenant, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.List(ctx, offset, limit)
}

func (s *tenantService) AddMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, tenantID)
	if err != nil {
		return ErrTenantNotFound
	}
	return s.repo.AddMember(ctx, tenantID, userID)
}

func (s *tenantService) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, tenantID)
	if err != nil {
		return ErrTenantNotFound
	}
	return s.repo.RemoveMember(ctx, tenantID, userID)
}

func (s *tenantService) GetMembers(ctx context.Context, tenantID uuid.UUID) ([]*entity.User, error) {
	_, err := s.repo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, ErrTenantNotFound
	}
	return s.repo.GetMembers(ctx, tenantID)
}

func (s *tenantService) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*entity.Tenant, error) {
	return s.repo.GetUserTenants(ctx, userID)
}

func (s *tenantService) IsMember(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	return s.repo.IsMember(ctx, tenantID, userID)
}

func (s *tenantService) BlockTenant(ctx context.Context, id uuid.UUID, reason string, blockedBy uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrTenantNotFound
	}
	return s.repo.UpdateBlocked(ctx, id, true, &reason, &blockedBy)
}

func (s *tenantService) UnblockTenant(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrTenantNotFound
	}
	return s.repo.UpdateBlocked(ctx, id, false, nil, nil)
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")
	return slug
}

func isValidSlug(slug string) bool {
	if len(slug) < 2 || len(slug) > 50 {
		return false
	}
	matched, _ := regexp.MatchString("^[a-z0-9][a-z0-9-]*[a-z0-9]$", slug)
	return matched
}
