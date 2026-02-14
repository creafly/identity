package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/domain/repository"
	"github.com/creafly/identity/internal/utils"
)

var (
	ErrClaimNotFound      = errors.New("claim not found")
	ErrClaimAlreadyExists = errors.New("claim with this value already exists")
)

type ClaimService interface {
	Create(ctx context.Context, input CreateClaimInput) (*entity.Claim, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Claim, error)
	GetByValue(ctx context.Context, value string) (*entity.Claim, error)
	GetOrCreate(ctx context.Context, value string) (*entity.Claim, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.Claim, error)
	AssignToUser(ctx context.Context, userID, claimID uuid.UUID) error
	RemoveFromUser(ctx context.Context, userID, claimID uuid.UUID) error
	GetUserClaims(ctx context.Context, userID uuid.UUID) ([]*entity.Claim, error)
	GetUserAllClaims(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) ([]*entity.Claim, error)
	AssignToRole(ctx context.Context, roleID, claimID uuid.UUID) error
	RemoveFromRole(ctx context.Context, roleID, claimID uuid.UUID) error
	GetRoleClaims(ctx context.Context, roleID uuid.UUID) ([]*entity.Claim, error)
}

type CreateClaimInput struct {
	Value string
}

type claimService struct {
	claimRepo      repository.ClaimRepository
	roleRepo       repository.RoleRepository
	tenantRoleRepo repository.TenantRoleRepository
}

func NewClaimService(
	claimRepo repository.ClaimRepository,
	roleRepo repository.RoleRepository,
	tenantRoleRepo repository.TenantRoleRepository,
) ClaimService {
	return &claimService{
		claimRepo:      claimRepo,
		roleRepo:       roleRepo,
		tenantRoleRepo: tenantRoleRepo,
	}
}

func (s *claimService) Create(ctx context.Context, input CreateClaimInput) (*entity.Claim, error) {
	existing, _ := s.claimRepo.GetByValue(ctx, input.Value)
	if existing != nil {
		return nil, ErrClaimAlreadyExists
	}

	claim := &entity.Claim{
		ID:        utils.GenerateUUID(),
		Value:     input.Value,
		CreatedAt: time.Now(),
	}

	if err := s.claimRepo.Create(ctx, claim); err != nil {
		return nil, err
	}

	return claim, nil
}

func (s *claimService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Claim, error) {
	claim, err := s.claimRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrClaimNotFound
	}
	return claim, nil
}

func (s *claimService) GetByValue(ctx context.Context, value string) (*entity.Claim, error) {
	claim, err := s.claimRepo.GetByValue(ctx, value)
	if err != nil {
		return nil, ErrClaimNotFound
	}
	return claim, nil
}

func (s *claimService) GetOrCreate(ctx context.Context, value string) (*entity.Claim, error) {
	existing, err := s.claimRepo.GetByValue(ctx, value)
	if err == nil && existing != nil {
		return existing, nil
	}

	claim := &entity.Claim{
		ID:        utils.GenerateUUID(),
		Value:     value,
		CreatedAt: time.Now(),
	}

	if err := s.claimRepo.Create(ctx, claim); err != nil {
		return nil, err
	}

	return claim, nil
}

func (s *claimService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.claimRepo.GetByID(ctx, id)
	if err != nil {
		return ErrClaimNotFound
	}
	return s.claimRepo.Delete(ctx, id)
}

func (s *claimService) List(ctx context.Context, offset, limit int) ([]*entity.Claim, error) {
	return s.claimRepo.List(ctx, offset, limit)
}

func (s *claimService) AssignToUser(ctx context.Context, userID, claimID uuid.UUID) error {
	_, err := s.claimRepo.GetByID(ctx, claimID)
	if err != nil {
		return ErrClaimNotFound
	}
	return s.claimRepo.AssignToUser(ctx, userID, claimID)
}

func (s *claimService) RemoveFromUser(ctx context.Context, userID, claimID uuid.UUID) error {
	return s.claimRepo.RemoveFromUser(ctx, userID, claimID)
}

func (s *claimService) GetUserClaims(ctx context.Context, userID uuid.UUID) ([]*entity.Claim, error) {
	return s.claimRepo.GetUserClaims(ctx, userID)
}

func (s *claimService) GetUserAllClaims(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) ([]*entity.Claim, error) {
	claimMap := make(map[uuid.UUID]*entity.Claim)

	userClaims, err := s.claimRepo.GetUserClaims(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, claim := range userClaims {
		claimMap[claim.ID] = claim
	}

	systemRoles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err == nil {
		for _, role := range systemRoles {
			roleClaims, err := s.claimRepo.GetRoleClaims(ctx, role.ID)
			if err != nil {
				continue
			}
			for _, claim := range roleClaims {
				claimMap[claim.ID] = claim
			}
		}
	}

	if tenantID != nil {
		tenantClaims, err := s.tenantRoleRepo.GetUserClaims(ctx, userID, *tenantID)
		if err == nil {
			for _, claim := range tenantClaims {
				claimMap[claim.ID] = claim
			}
		}
	}

	allClaims := make([]*entity.Claim, 0, len(claimMap))
	for _, claim := range claimMap {
		allClaims = append(allClaims, claim)
	}

	return allClaims, nil
}

func (s *claimService) AssignToRole(ctx context.Context, roleID, claimID uuid.UUID) error {
	_, err := s.claimRepo.GetByID(ctx, claimID)
	if err != nil {
		return ErrClaimNotFound
	}
	_, err = s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return ErrRoleNotFound
	}
	return s.claimRepo.AssignToRole(ctx, roleID, claimID)
}

func (s *claimService) RemoveFromRole(ctx context.Context, roleID, claimID uuid.UUID) error {
	return s.claimRepo.RemoveFromRole(ctx, roleID, claimID)
}

func (s *claimService) GetRoleClaims(ctx context.Context, roleID uuid.UUID) ([]*entity.Claim, error) {
	return s.claimRepo.GetRoleClaims(ctx, roleID)
}
