package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/domain/repository"
	"github.com/creafly/identity/internal/utils"
)

const BcryptCost = 12

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUsernameAlreadyUsed = errors.New("username already used")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserBlocked         = errors.New("user is blocked")
)

type UserService interface {
	Register(ctx context.Context, input RegisterInput) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*entity.User, error)
	BlockUser(ctx context.Context, id uuid.UUID, reason string, blockedBy uuid.UUID) error
	UnblockUser(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.User, error)
	ValidateCredentials(ctx context.Context, email, password string) (*entity.User, error)
	ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error
}

type RegisterInput struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
	Locale    string
}

type UpdateUserInput struct {
	Email     *string
	Username  *string
	FirstName *string
	LastName  *string
	AvatarURL *string
	Locale    *string
	IsActive  *bool
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, input RegisterInput) (*entity.User, error) {
	existing, _ := s.repo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	if input.Username != "" {
		existingUsername, _ := s.repo.GetByUsername(ctx, input.Username)
		if existingUsername != nil {
			return nil, ErrUsernameAlreadyUsed
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), BcryptCost)
	if err != nil {
		return nil, err
	}

	var username *string
	if input.Username != "" {
		username = &input.Username
	}

	user := &entity.User{
		ID:           utils.GenerateUUID(),
		Email:        input.Email,
		Username:     username,
		PasswordHash: string(hashedPassword),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Locale:       input.Locale,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*entity.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Username != nil {
		if *input.Username != "" {
			existing, _ := s.repo.GetByUsername(ctx, *input.Username)
			if existing != nil && existing.ID != user.ID {
				return nil, ErrUsernameAlreadyUsed
			}
		}
		user.Username = input.Username
	}
	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}
	if input.Locale != nil {
		user.Locale = *input.Locale
	}
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) BlockUser(ctx context.Context, id uuid.UUID, reason string, blockedBy uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}
	return s.repo.UpdateBlocked(ctx, id, true, &reason, &blockedBy)
}

func (s *userService) UnblockUser(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}
	return s.repo.UpdateBlocked(ctx, id, false, nil, nil)
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *userService) List(ctx context.Context, offset, limit int) ([]*entity.User, error) {
	return s.repo.List(ctx, offset, limit)
}

func (s *userService) ValidateCredentials(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrInvalidCredentials
	}

	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	return user, nil
}

func (s *userService) ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return s.repo.Update(ctx, user)
}
