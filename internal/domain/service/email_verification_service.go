package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/domain/repository"
	"github.com/creafly/outbox"
	"github.com/google/uuid"
)

var (
	ErrEmailAlreadyVerified          = errors.New("email already verified")
	ErrEmailVerificationCodeNotFound = errors.New("verification code not found or expired")
	ErrEmailVerificationCodeInvalid  = errors.New("invalid verification code")
	ErrEmailVerificationCodeExpired  = errors.New("verification code expired")
	ErrTooManyVerificationRequests   = errors.New("too many verification requests, please wait")
)

const (
	VerificationCodeLength = 6
	VerificationCodeExpiry = 15 * time.Minute
	ResendCooldown         = 60 * time.Second
)

type EmailVerificationService interface {
	RequestVerification(ctx context.Context, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error
	ResendVerification(ctx context.Context, userID uuid.UUID) error
}

type emailVerificationService struct {
	userRepo              repository.UserRepository
	emailVerificationRepo repository.EmailVerificationRepository
	outboxRepo            outbox.Repository
}

func NewEmailVerificationService(
	userRepo repository.UserRepository,
	emailVerificationRepo repository.EmailVerificationRepository,
	outboxRepo outbox.Repository,
) EmailVerificationService {
	return &emailVerificationService{
		userRepo:              userRepo,
		emailVerificationRepo: emailVerificationRepo,
		outboxRepo:            outboxRepo,
	}
}

func (s *emailVerificationService) RequestVerification(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	code, err := s.generateVerificationCode()
	if err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(code))
	codeHash := hex.EncodeToString(hash[:])

	_ = s.emailVerificationRepo.DeleteByUserID(ctx, userID)

	token := &entity.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		CodeHash:  codeHash,
		ExpiresAt: time.Now().Add(VerificationCodeExpiry),
		CreatedAt: time.Now(),
	}

	if err := s.emailVerificationRepo.Create(ctx, token); err != nil {
		return err
	}

	payload, err := outbox.CreatePayload(map[string]any{
		"type":      "email_verification",
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"code":      code,
		"expiresAt": token.ExpiresAt.Format(time.RFC3339),
		"locale":    user.Locale,
	})
	if err != nil {
		return err
	}

	event := outbox.NewEvent("notifications.email", payload)

	return s.outboxRepo.Create(ctx, event)
}

func (s *emailVerificationService) VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	hash := sha256.Sum256([]byte(code))
	codeHash := hex.EncodeToString(hash[:])

	token, err := s.emailVerificationRepo.GetByCodeHash(ctx, userID, codeHash)
	if err != nil {
		return ErrEmailVerificationCodeInvalid
	}

	if time.Now().After(token.ExpiresAt) {
		return ErrEmailVerificationCodeExpired
	}

	if err := s.emailVerificationRepo.MarkAsUsed(ctx, token.ID); err != nil {
		return err
	}

	if err := s.userRepo.MarkEmailVerified(ctx, userID); err != nil {
		return err
	}

	_ = s.emailVerificationRepo.DeleteByUserID(ctx, userID)

	return nil
}

func (s *emailVerificationService) ResendVerification(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	latestToken, err := s.emailVerificationRepo.GetLatestByUserID(ctx, userID)
	if err == nil && latestToken != nil {
		if time.Since(latestToken.CreatedAt) < ResendCooldown {
			return ErrTooManyVerificationRequests
		}
	}

	return s.RequestVerification(ctx, userID)
}

func (s *emailVerificationService) generateVerificationCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
