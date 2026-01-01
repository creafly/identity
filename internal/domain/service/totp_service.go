package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"image/png"

	"github.com/creafly/identity/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var (
	ErrTOTPAlreadyEnabled = errors.New("totp already enabled")
	ErrTOTPNotEnabled     = errors.New("totp not enabled")
	ErrTOTPNotSetup       = errors.New("totp not set up")
	ErrTOTPInvalidCode    = errors.New("invalid totp code")
	ErrTOTPSecretNotFound = errors.New("totp secret not found")
)

type TOTPSetupResponse struct {
	QRCodePNG string `json:"qrCodePng"`
}

type TOTPService interface {
	GenerateSecret(ctx context.Context, userID uuid.UUID, email string, issuer string) (*TOTPSetupResponse, error)
	ValidateCode(ctx context.Context, userID uuid.UUID, code string) (bool, error)
	Enable(ctx context.Context, userID uuid.UUID, code string) error
	Disable(ctx context.Context, userID uuid.UUID, password string) error
	IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error)
}

type totpService struct {
	userRepo    repository.UserRepository
	userService UserService
}

func NewTOTPService(userRepo repository.UserRepository, userService UserService) TOTPService {
	return &totpService{
		userRepo:    userRepo,
		userService: userService,
	}
}

func (s *totpService) GenerateSecret(ctx context.Context, userID uuid.UUID, email string, issuer string) (*TOTPSetupResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if user.TotpEnabled {
		return nil, ErrTOTPAlreadyEnabled
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
		Period:      30,
		SecretSize:  32,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.SetTOTPSecret(ctx, userID, key.Secret()); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, err
	}
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	qrCodePNG := base64.StdEncoding.EncodeToString(buf.Bytes())

	return &TOTPSetupResponse{
		QRCodePNG: qrCodePNG,
	}, nil
}

func (s *totpService) ValidateCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, ErrUserNotFound
	}

	if user.TotpSecret == nil || *user.TotpSecret == "" {
		return false, ErrTOTPSecretNotFound
	}

	valid := totp.Validate(code, *user.TotpSecret)
	return valid, nil
}

func (s *totpService) Enable(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.TotpEnabled {
		return ErrTOTPAlreadyEnabled
	}

	if user.TotpSecret == nil || *user.TotpSecret == "" {
		return ErrTOTPNotSetup
	}

	valid := totp.Validate(code, *user.TotpSecret)
	if !valid {
		return ErrTOTPInvalidCode
	}

	return s.userRepo.EnableTOTP(ctx, userID)
}

func (s *totpService) Disable(ctx context.Context, userID uuid.UUID, password string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if !user.TotpEnabled {
		return ErrTOTPNotEnabled
	}

	_, err = s.userService.ValidateCredentials(ctx, user.Email, password)
	if err != nil {
		return ErrInvalidCredentials
	}

	return s.userRepo.DisableTOTP(ctx, userID)
}

func (s *totpService) IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, ErrUserNotFound
	}

	return user.TotpEnabled, nil
}
