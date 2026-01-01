package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hexaend/identity/internal/domain/entity"
	"github.com/hexaend/identity/internal/domain/repository"
)

var (
	ErrTenantRoleNotFound      = errors.New("tenant role not found")
	ErrTenantRoleAlreadyExists = errors.New("tenant role with this name already exists")
	ErrCannotDeleteDefaultRole = errors.New("cannot delete default role")
)

type TenantRoleService interface {
	Create(ctx context.Context, input CreateTenantRoleInput) (*entity.TenantRole, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.TenantRole, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*entity.TenantRole, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateTenantRoleInput) (*entity.TenantRole, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantRole, error)

	AddClaim(ctx context.Context, tenantRoleID, claimID uuid.UUID) error
	RemoveClaim(ctx context.Context, tenantRoleID, claimID uuid.UUID) error
	GetRoleClaims(ctx context.Context, tenantRoleID uuid.UUID) ([]*entity.Claim, error)
	GetAvailableClaims(ctx context.Context, tenantID uuid.UUID) ([]*entity.Claim, error)
	BatchUpdateClaims(ctx context.Context, tenantRoleID uuid.UUID, assignClaimIDs, removeClaimIDs []uuid.UUID) error

	AssignToUser(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error
	RemoveFromUser(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.TenantRole, error)
	GetUserClaims(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.Claim, error)

	CreateDefaultRoles(ctx context.Context, tenantID uuid.UUID) error
	AssignOwnerRole(ctx context.Context, userID, tenantID uuid.UUID) error
}

type CreateTenantRoleInput struct {
	TenantID    uuid.UUID
	Name        string
	Description string
}

type UpdateTenantRoleInput struct {
	Name        *string
	Description *string
}

type tenantRoleService struct {
	repo      repository.TenantRoleRepository
	claimRepo repository.ClaimRepository
}

func NewTenantRoleService(repo repository.TenantRoleRepository, claimRepo repository.ClaimRepository) TenantRoleService {
	return &tenantRoleService{repo: repo, claimRepo: claimRepo}
}

func (s *tenantRoleService) Create(ctx context.Context, input CreateTenantRoleInput) (*entity.TenantRole, error) {
	existing, _ := s.repo.GetByName(ctx, input.TenantID, input.Name)
	if existing != nil {
		return nil, ErrTenantRoleAlreadyExists
	}

	role := &entity.TenantRole{
		ID:          uuid.New(),
		TenantID:    input.TenantID,
		Name:        input.Name,
		Description: input.Description,
		IsDefault:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *tenantRoleService) GetByID(ctx context.Context, id uuid.UUID) (*entity.TenantRole, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTenantRoleNotFound
	}
	return role, nil
}

func (s *tenantRoleService) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*entity.TenantRole, error) {
	role, err := s.repo.GetByName(ctx, tenantID, name)
	if err != nil {
		return nil, ErrTenantRoleNotFound
	}
	return role, nil
}

func (s *tenantRoleService) Update(ctx context.Context, id uuid.UUID, input UpdateTenantRoleInput) (*entity.TenantRole, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTenantRoleNotFound
	}

	if input.Name != nil {
		existing, _ := s.repo.GetByName(ctx, role.TenantID, *input.Name)
		if existing != nil && existing.ID != id {
			return nil, ErrTenantRoleAlreadyExists
		}
		role.Name = *input.Name
	}
	if input.Description != nil {
		role.Description = *input.Description
	}
	role.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *tenantRoleService) Delete(ctx context.Context, id uuid.UUID) error {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrTenantRoleNotFound
	}
	if role.IsDefault {
		return ErrCannotDeleteDefaultRole
	}
	return s.repo.Delete(ctx, id)
}

func (s *tenantRoleService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantRole, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}

func (s *tenantRoleService) AddClaim(ctx context.Context, tenantRoleID, claimID uuid.UUID) error {
	return s.repo.AddClaim(ctx, tenantRoleID, claimID)
}

func (s *tenantRoleService) RemoveClaim(ctx context.Context, tenantRoleID, claimID uuid.UUID) error {
	return s.repo.RemoveClaim(ctx, tenantRoleID, claimID)
}

func (s *tenantRoleService) GetRoleClaims(ctx context.Context, tenantRoleID uuid.UUID) ([]*entity.Claim, error) {
	return s.repo.GetRoleClaims(ctx, tenantRoleID)
}

func (s *tenantRoleService) GetAvailableClaims(ctx context.Context, tenantID uuid.UUID) ([]*entity.Claim, error) {
	return s.repo.GetTenantAvailableClaims(ctx, tenantID)
}

func (s *tenantRoleService) BatchUpdateClaims(ctx context.Context, tenantRoleID uuid.UUID, assignClaimIDs, removeClaimIDs []uuid.UUID) error {
	return s.repo.BatchUpdateClaims(ctx, tenantRoleID, assignClaimIDs, removeClaimIDs)
}

func (s *tenantRoleService) AssignToUser(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, tenantRoleID)
	if err != nil {
		return ErrTenantRoleNotFound
	}
	return s.repo.AssignToUser(ctx, userID, tenantID, tenantRoleID)
}

func (s *tenantRoleService) RemoveFromUser(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error {
	return s.repo.RemoveFromUser(ctx, userID, tenantID, tenantRoleID)
}

func (s *tenantRoleService) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.TenantRole, error) {
	return s.repo.GetUserRoles(ctx, userID, tenantID)
}

func (s *tenantRoleService) GetUserClaims(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.Claim, error) {
	return s.repo.GetUserClaims(ctx, userID, tenantID)
}

func (s *tenantRoleService) CreateDefaultRoles(ctx context.Context, tenantID uuid.UUID) error {
	return s.repo.CreateDefaultRoles(ctx, tenantID)
}

func (s *tenantRoleService) AssignOwnerRole(ctx context.Context, userID, tenantID uuid.UUID) error {
	ownerRole, err := s.repo.GetByName(ctx, tenantID, "owner")
	if err != nil {
		return ErrTenantRoleNotFound
	}
	return s.repo.AssignToUser(ctx, userID, tenantID, ownerRole.ID)
}
