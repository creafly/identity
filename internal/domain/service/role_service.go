package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/domain/repository"
)

var (
	ErrRoleNotFound      = errors.New("role not found")
	ErrRoleAlreadyExists = errors.New("role with this name already exists")
	ErrCannotDeleteRole  = errors.New("cannot delete role with assigned users")
)

type RoleService interface {
	Create(ctx context.Context, input CreateRoleInput) (*entity.Role, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateRoleInput) (*entity.Role, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.Role, error)
	AssignToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error)
	AssignDefaultUserRole(ctx context.Context, userID uuid.UUID) error
}

type CreateRoleInput struct {
	Name        string
	Description string
}

type UpdateRoleInput struct {
	Name        *string
	Description *string
}

type roleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) Create(ctx context.Context, input CreateRoleInput) (*entity.Role, error) {
	existing, _ := s.repo.GetByName(ctx, input.Name)
	if existing != nil {
		return nil, ErrRoleAlreadyExists
	}

	role := &entity.Role{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *roleService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	return role, nil
}

func (s *roleService) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	role, err := s.repo.GetByName(ctx, name)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	return role, nil
}

func (s *roleService) Update(ctx context.Context, id uuid.UUID, input UpdateRoleInput) (*entity.Role, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrRoleNotFound
	}

	if input.Name != nil {
		existing, _ := s.repo.GetByName(ctx, *input.Name)
		if existing != nil && existing.ID != id {
			return nil, ErrRoleAlreadyExists
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

func (s *roleService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrRoleNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *roleService) List(ctx context.Context, offset, limit int) ([]*entity.Role, error) {
	return s.repo.List(ctx, offset, limit)
}

func (s *roleService) AssignToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, roleID)
	if err != nil {
		return ErrRoleNotFound
	}
	return s.repo.AssignToUser(ctx, userID, roleID)
}

func (s *roleService) RemoveFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.repo.RemoveFromUser(ctx, userID, roleID)
}

func (s *roleService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error) {
	return s.repo.GetUserRoles(ctx, userID)
}

func (s *roleService) AssignDefaultUserRole(ctx context.Context, userID uuid.UUID) error {
	userRoleID, err := uuid.Parse(entity.SystemRoleUserID)
	if err != nil {
		return err
	}
	return s.repo.AssignToUser(ctx, userID, userRoleID)
}
