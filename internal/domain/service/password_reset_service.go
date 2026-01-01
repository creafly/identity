package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/domain/repository"
	"github.com/creafly/identity/internal/infra/outbox"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordResetTokenNotFound = errors.New("password reset token not found or expired")
	ErrPasswordResetTokenExpired  = errors.New("password reset token expired")
)

type PasswordResetService interface {
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type passwordResetService struct {
	userRepo          repository.UserRepository
	passwordResetRepo repository.PasswordResetRepository
	outboxRepo        repository.OutboxRepository
}

func NewPasswordResetService(
	userRepo repository.UserRepository,
	passwordResetRepo repository.PasswordResetRepository,
	outboxRepo repository.OutboxRepository,
) PasswordResetService {
	return &passwordResetService{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		outboxRepo:        outboxRepo,
	}
}

func (s *passwordResetService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	_ = s.passwordResetRepo.DeleteByUserID(ctx, user.ID)

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	rawToken := hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	token := &entity.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.passwordResetRepo.Create(ctx, token); err != nil {
		return err
	}

	payload, err := outbox.CreatePayload(map[string]any{
		"type":      "password_reset",
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"token":     rawToken,
		"expiresAt": token.ExpiresAt.Format(time.RFC3339),
		"locale":    user.Locale,
	})
	if err != nil {
		return err
	}

	event := &entity.OutboxEvent{
		ID:        uuid.New(),
		EventType: "notifications.email",
		Payload:   payload,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	return s.outboxRepo.Create(ctx, event)
}

func (s *passwordResetService) ResetPassword(ctx context.Context, token, newPassword string) error {
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	resetToken, err := s.passwordResetRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return ErrPasswordResetTokenNotFound
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return ErrPasswordResetTokenExpired
	}

	user, err := s.userRepo.GetByID(ctx, resetToken.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	if err := s.passwordResetRepo.MarkAsUsed(ctx, resetToken.ID); err != nil {
		return err
	}

	_ = s.passwordResetRepo.DeleteByUserID(ctx, user.ID)

	return nil
}
