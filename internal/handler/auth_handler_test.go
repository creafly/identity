package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/creafly/identity/internal/config"
	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/domain/service"
	"github.com/creafly/identity/internal/testutil"
	"github.com/creafly/identity/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	validator.Init()
}

type TokenServiceMock struct {
	GenerateAccessTokenFunc                 func(userID uuid.UUID, email string, roles []string) (string, error)
	GenerateRefreshTokenFunc                func(userID uuid.UUID) (string, error)
	GenerateTempTokenFunc                   func(userID uuid.UUID, email string) (string, error)
	ValidateAccessTokenFunc                 func(tokenString string) (*service.AccessTokenClaims, error)
	ValidateRefreshTokenFunc                func(tokenString string) (*service.RefreshTokenClaims, error)
	ValidateTempTokenFunc                   func(tokenString string) (*service.TempTokenClaims, error)
	GenerateAccessTokenWithFingerprintFunc  func(userID uuid.UUID, email string, roles []string, fingerprint string) (string, error)
	GenerateRefreshTokenWithFingerprintFunc func(userID uuid.UUID, fingerprint string) (string, error)
	ValidateAccessTokenWithFingerprintFunc  func(tokenString string, fingerprint string) (*service.AccessTokenClaims, error)
	ValidateRefreshTokenWithFingerprintFunc func(tokenString string, fingerprint string) (*service.RefreshTokenClaims, error)
}

func (m *TokenServiceMock) GenerateAccessToken(userID uuid.UUID, email string, roles []string) (string, error) {
	if m.GenerateAccessTokenFunc != nil {
		return m.GenerateAccessTokenFunc(userID, email, roles)
	}
	return "access-token", nil
}

func (m *TokenServiceMock) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	if m.GenerateRefreshTokenFunc != nil {
		return m.GenerateRefreshTokenFunc(userID)
	}
	return "refresh-token", nil
}

func (m *TokenServiceMock) GenerateTempToken(userID uuid.UUID, email string) (string, error) {
	if m.GenerateTempTokenFunc != nil {
		return m.GenerateTempTokenFunc(userID, email)
	}
	return "temp-token", nil
}

func (m *TokenServiceMock) ValidateAccessToken(tokenString string) (*service.AccessTokenClaims, error) {
	if m.ValidateAccessTokenFunc != nil {
		return m.ValidateAccessTokenFunc(tokenString)
	}
	return nil, nil
}

func (m *TokenServiceMock) ValidateRefreshToken(tokenString string) (*service.RefreshTokenClaims, error) {
	if m.ValidateRefreshTokenFunc != nil {
		return m.ValidateRefreshTokenFunc(tokenString)
	}
	return nil, nil
}

func (m *TokenServiceMock) ValidateTempToken(tokenString string) (*service.TempTokenClaims, error) {
	if m.ValidateTempTokenFunc != nil {
		return m.ValidateTempTokenFunc(tokenString)
	}
	return nil, nil
}

func (m *TokenServiceMock) RevokeToken(token string, expiresAt time.Time) {}

func (m *TokenServiceMock) RevokeAllUserTokens(userID uuid.UUID, expiresAt time.Time) {}

func (m *TokenServiceMock) IsTokenRevoked(token string) bool {
	return false
}

func (m *TokenServiceMock) IsUserTokensRevoked(userID uuid.UUID) bool {
	return false
}

func (m *TokenServiceMock) GenerateAccessTokenWithFingerprint(userID uuid.UUID, email string, roles []string, fingerprint string) (string, error) {
	if m.GenerateAccessTokenWithFingerprintFunc != nil {
		return m.GenerateAccessTokenWithFingerprintFunc(userID, email, roles, fingerprint)
	}
	return "access-token", nil
}

func (m *TokenServiceMock) GenerateRefreshTokenWithFingerprint(userID uuid.UUID, fingerprint string) (string, error) {
	if m.GenerateRefreshTokenWithFingerprintFunc != nil {
		return m.GenerateRefreshTokenWithFingerprintFunc(userID, fingerprint)
	}
	return "refresh-token", nil
}

func (m *TokenServiceMock) ValidateAccessTokenWithFingerprint(tokenString string, fingerprint string) (*service.AccessTokenClaims, error) {
	if m.ValidateAccessTokenWithFingerprintFunc != nil {
		return m.ValidateAccessTokenWithFingerprintFunc(tokenString, fingerprint)
	}
	return nil, nil
}

func (m *TokenServiceMock) ValidateRefreshTokenWithFingerprint(tokenString string, fingerprint string) (*service.RefreshTokenClaims, error) {
	if m.ValidateRefreshTokenWithFingerprintFunc != nil {
		return m.ValidateRefreshTokenWithFingerprintFunc(tokenString, fingerprint)
	}
	return nil, nil
}

type TOTPServiceMock struct {
	IsEnabledFunc    func(ctx context.Context, userID uuid.UUID) (bool, error)
	ValidateCodeFunc func(ctx context.Context, userID uuid.UUID, code string) (bool, error)
}

func (m *TOTPServiceMock) GenerateSecret(ctx context.Context, userID uuid.UUID, email string, issuer string) (*service.TOTPSetupResponse, error) {
	return nil, nil
}

func (m *TOTPServiceMock) Enable(ctx context.Context, userID uuid.UUID, code string) error {
	return nil
}

func (m *TOTPServiceMock) Disable(ctx context.Context, userID uuid.UUID, password string) error {
	return nil
}

func (m *TOTPServiceMock) IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	if m.IsEnabledFunc != nil {
		return m.IsEnabledFunc(ctx, userID)
	}
	return false, nil
}

func (m *TOTPServiceMock) ValidateCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	if m.ValidateCodeFunc != nil {
		return m.ValidateCodeFunc(ctx, userID, code)
	}
	return true, nil
}

type PasswordResetServiceMock struct {
	RequestPasswordResetFunc func(ctx context.Context, email string) error
	ResetPasswordFunc        func(ctx context.Context, token, newPassword string) error
}

func (m *PasswordResetServiceMock) RequestPasswordReset(ctx context.Context, email string) error {
	if m.RequestPasswordResetFunc != nil {
		return m.RequestPasswordResetFunc(ctx, email)
	}
	return nil
}

func (m *PasswordResetServiceMock) ResetPassword(ctx context.Context, token, newPassword string) error {
	if m.ResetPasswordFunc != nil {
		return m.ResetPasswordFunc(ctx, token, newPassword)
	}
	return nil
}

type EmailVerificationServiceMock struct {
	RequestVerificationFunc func(ctx context.Context, userID uuid.UUID) error
	VerifyEmailFunc         func(ctx context.Context, userID uuid.UUID, code string) error
	ResendVerificationFunc  func(ctx context.Context, userID uuid.UUID) error
}

func (m *EmailVerificationServiceMock) RequestVerification(ctx context.Context, userID uuid.UUID) error {
	if m.RequestVerificationFunc != nil {
		return m.RequestVerificationFunc(ctx, userID)
	}
	return nil
}

func (m *EmailVerificationServiceMock) VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error {
	if m.VerifyEmailFunc != nil {
		return m.VerifyEmailFunc(ctx, userID, code)
	}
	return nil
}

func (m *EmailVerificationServiceMock) ResendVerification(ctx context.Context, userID uuid.UUID) error {
	if m.ResendVerificationFunc != nil {
		return m.ResendVerificationFunc(ctx, userID)
	}
	return nil
}

type AuthRoleServiceMock struct {
	AssignDefaultUserRoleFunc func(ctx context.Context, userID uuid.UUID) error
}

func (m *AuthRoleServiceMock) Create(ctx context.Context, input service.CreateRoleInput) (*entity.Role, error) {
	return nil, nil
}
func (m *AuthRoleServiceMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	return nil, nil
}
func (m *AuthRoleServiceMock) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	return nil, nil
}
func (m *AuthRoleServiceMock) Update(ctx context.Context, id uuid.UUID, input service.UpdateRoleInput) (*entity.Role, error) {
	return nil, nil
}
func (m *AuthRoleServiceMock) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *AuthRoleServiceMock) List(ctx context.Context, offset, limit int) ([]*entity.Role, error) {
	return nil, nil
}
func (m *AuthRoleServiceMock) AssignToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return nil
}
func (m *AuthRoleServiceMock) RemoveFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return nil
}
func (m *AuthRoleServiceMock) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error) {
	return nil, nil
}
func (m *AuthRoleServiceMock) AssignDefaultUserRole(ctx context.Context, userID uuid.UUID) error {
	if m.AssignDefaultUserRoleFunc != nil {
		return m.AssignDefaultUserRoleFunc(ctx, userID)
	}
	return nil
}

type LoginAttemptTrackerMock struct {
	RecordFailedAttemptFunc func(identifier string)
	GetAttemptsFunc         func(identifier string) int
	IsLockedFunc            func(identifier string) bool
	GetLockoutRemainingFunc func(identifier string) time.Duration
	ClearAttemptsFunc       func(identifier string)
}

func (m *LoginAttemptTrackerMock) RecordFailedAttempt(identifier string) {
	if m.RecordFailedAttemptFunc != nil {
		m.RecordFailedAttemptFunc(identifier)
	}
}

func (m *LoginAttemptTrackerMock) GetAttempts(identifier string) int {
	if m.GetAttemptsFunc != nil {
		return m.GetAttemptsFunc(identifier)
	}
	return 0
}

func (m *LoginAttemptTrackerMock) IsLocked(identifier string) bool {
	if m.IsLockedFunc != nil {
		return m.IsLockedFunc(identifier)
	}
	return false
}

func (m *LoginAttemptTrackerMock) GetLockoutRemaining(identifier string) time.Duration {
	if m.GetLockoutRemainingFunc != nil {
		return m.GetLockoutRemainingFunc(identifier)
	}
	return 0
}

func (m *LoginAttemptTrackerMock) ClearAttempts(identifier string) {
	if m.ClearAttemptsFunc != nil {
		m.ClearAttemptsFunc(identifier)
	}
}

type TOTPAttemptTrackerMock struct {
	RecordFailedAttemptFunc func(identifier string)
	GetAttemptsFunc         func(identifier string) int
	IsLockedFunc            func(identifier string) bool
	GetLockoutRemainingFunc func(identifier string) time.Duration
	ClearAttemptsFunc       func(identifier string)
}

func (m *TOTPAttemptTrackerMock) RecordFailedAttempt(identifier string) {
	if m.RecordFailedAttemptFunc != nil {
		m.RecordFailedAttemptFunc(identifier)
	}
}

func (m *TOTPAttemptTrackerMock) GetAttempts(identifier string) int {
	if m.GetAttemptsFunc != nil {
		return m.GetAttemptsFunc(identifier)
	}
	return 0
}

func (m *TOTPAttemptTrackerMock) IsLocked(identifier string) bool {
	if m.IsLockedFunc != nil {
		return m.IsLockedFunc(identifier)
	}
	return false
}

func (m *TOTPAttemptTrackerMock) GetLockoutRemaining(identifier string) time.Duration {
	if m.GetLockoutRemainingFunc != nil {
		return m.GetLockoutRemainingFunc(identifier)
	}
	return 0
}

func (m *TOTPAttemptTrackerMock) ClearAttempts(identifier string) {
	if m.ClearAttemptsFunc != nil {
		m.ClearAttemptsFunc(identifier)
	}
}

type authMocks struct {
	cfg                  *config.Config
	userSvc              *UserServiceMock
	tokenSvc             *TokenServiceMock
	roleSvc              *AuthRoleServiceMock
	totpSvc              *TOTPServiceMock
	passwordResetSvc     *PasswordResetServiceMock
	emailVerificationSvc *EmailVerificationServiceMock
	claimSvc             *ClaimServiceMock
	loginAttemptTracker  *LoginAttemptTrackerMock
	totpAttemptTracker   *TOTPAttemptTrackerMock
}

func newAuthMocks() *authMocks {
	return &authMocks{
		cfg:                  &config.Config{},
		userSvc:              &UserServiceMock{},
		tokenSvc:             &TokenServiceMock{},
		roleSvc:              &AuthRoleServiceMock{},
		totpSvc:              &TOTPServiceMock{},
		passwordResetSvc:     &PasswordResetServiceMock{},
		emailVerificationSvc: &EmailVerificationServiceMock{},
		claimSvc:             &ClaimServiceMock{},
		loginAttemptTracker:  &LoginAttemptTrackerMock{},
		totpAttemptTracker:   &TOTPAttemptTrackerMock{},
	}
}

func setupAuthRouter(m *authMocks) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewAuthHandler(m.cfg, m.userSvc, m.tokenSvc, m.roleSvc, m.totpSvc, m.passwordResetSvc, m.emailVerificationSvc, m.claimSvc, m.loginAttemptTracker, m.totpAttemptTracker)

	router.POST("/auth/register", handler.Register)
	router.POST("/auth/login", handler.Login)
	router.POST("/auth/refresh", handler.Refresh)
	router.POST("/auth/forgot-password", handler.ForgotPassword)
	router.POST("/auth/reset-password", handler.ResetPassword)
	router.GET("/auth/me", handler.Me)

	return router
}

func TestAuthHandler_Register(t *testing.T) {
	m := newAuthMocks()
	router := setupAuthRouter(m)

	t.Run("valid registration", func(t *testing.T) {
		user := testutil.NewTestUser()
		m.userSvc.RegisterFunc = func(ctx context.Context, input service.RegisterInput) (*entity.User, error) {
			return user, nil
		}

		body := `{"email": "test@example.com", "password": "Test1234!", "firstName": "John", "lastName": "Doe"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Register() status = %d, want %d", w.Code, http.StatusCreated)
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		body := `{"email": "test@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Register() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("user already exists", func(t *testing.T) {
		m.userSvc.RegisterFunc = func(ctx context.Context, input service.RegisterInput) (*entity.User, error) {
			return nil, service.ErrUserAlreadyExists
		}

		body := `{"email": "test@example.com", "password": "Test1234!", "firstName": "John", "lastName": "Doe"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Register() status = %d, want %d", w.Code, http.StatusConflict)
		}
	})
}

func TestAuthHandler_Login(t *testing.T) {
	m := newAuthMocks()
	router := setupAuthRouter(m)

	t.Run("valid login", func(t *testing.T) {
		user := testutil.NewTestUser()
		m.userSvc.ValidateCredentialsFunc = func(ctx context.Context, email, password string) (*entity.User, error) {
			return user, nil
		}
		m.totpSvc.IsEnabledFunc = func(ctx context.Context, userID uuid.UUID) (bool, error) {
			return false, nil
		}

		body := `{"email": "test@example.com", "password": "Test1234!"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		m.userSvc.ValidateCredentialsFunc = func(ctx context.Context, email, password string) (*entity.User, error) {
			return nil, service.ErrInvalidCredentials
		}

		body := `{"email": "test@example.com", "password": "wrongpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Login() status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("user blocked", func(t *testing.T) {
		m.userSvc.ValidateCredentialsFunc = func(ctx context.Context, email, password string) (*entity.User, error) {
			return nil, service.ErrUserBlocked
		}

		body := `{"email": "test@example.com", "password": "Test1234!"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Login() status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})
}

func TestAuthHandler_ForgotPassword(t *testing.T) {
	m := newAuthMocks()
	router := setupAuthRouter(m)

	t.Run("valid request", func(t *testing.T) {
		body := `{"email": "test@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("ForgotPassword() status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestAuthHandler_ResetPassword(t *testing.T) {
	m := newAuthMocks()
	router := setupAuthRouter(m)

	t.Run("valid reset", func(t *testing.T) {
		m.passwordResetSvc.ResetPasswordFunc = func(ctx context.Context, token, newPassword string) error {
			return nil
		}

		body := `{"token": "valid-token", "newPassword": "NewPass123!"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("ResetPassword() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("token not found", func(t *testing.T) {
		m.passwordResetSvc.ResetPasswordFunc = func(ctx context.Context, token, newPassword string) error {
			return service.ErrPasswordResetTokenNotFound
		}

		body := `{"token": "invalid-token", "newPassword": "NewPass123!"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("ResetPassword() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}
