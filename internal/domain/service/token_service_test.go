package service

import (
	"testing"
	"time"

	"github.com/creafly/identity/internal/config"
	"github.com/creafly/identity/internal/utils"
)

type mockBlacklist struct {
	tokens      map[string]time.Time
	userRevokes map[string]time.Time
}

func newMockBlacklist() *mockBlacklist {
	return &mockBlacklist{
		tokens:      make(map[string]time.Time),
		userRevokes: make(map[string]time.Time),
	}
}

func (m *mockBlacklist) Add(tokenHash string, expiresAt time.Time) {
	m.tokens[tokenHash] = expiresAt
}

func (m *mockBlacklist) IsBlacklisted(tokenHash string) bool {
	exp, exists := m.tokens[tokenHash]
	return exists && time.Now().Before(exp)
}

func (m *mockBlacklist) RevokeAllForUser(userID string, expiresAt time.Time) {
	m.userRevokes[userID] = expiresAt
}

func (m *mockBlacklist) IsUserRevoked(userID string) bool {
	exp, exists := m.userRevokes[userID]
	return exists && time.Now().Before(exp)
}

func newTestTokenService() TokenService {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:               "test-secret-key-at-least-32-chars",
			AccessTokenSecret:    "test-access-secret-key-32-chars!",
			RefreshTokenSecret:   "test-refresh-secret-key-32chars!",
			TempTokenSecret:      "test-temp-secret-key-32-chars!!",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 24 * time.Hour,
		},
	}
	return NewTokenService(cfg, newMockBlacklist())
}

func TestGenerateAccessToken(t *testing.T) {
	svc := newTestTokenService()

	userID := utils.GenerateUUID()
	email := "test@example.com"
	roles := []string{"admin", "user"}

	token, err := svc.GenerateAccessToken(userID, email, roles)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("userID mismatch: got %v, want %v", claims.UserID, userID)
	}

	if claims.Email != email {
		t.Errorf("email mismatch: got %v, want %v", claims.Email, email)
	}

	if len(claims.Roles) != len(roles) {
		t.Errorf("roles length mismatch: got %d, want %d", len(claims.Roles), len(roles))
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	svc := newTestTokenService()

	userID := utils.GenerateUUID()

	token, err := svc.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	claims, err := svc.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("failed to validate refresh token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("userID mismatch: got %v, want %v", claims.UserID, userID)
	}
}

func TestGenerateTempToken(t *testing.T) {
	svc := newTestTokenService()

	userID := utils.GenerateUUID()
	email := "test@example.com"

	token, err := svc.GenerateTempToken(userID, email)
	if err != nil {
		t.Fatalf("failed to generate temp token: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	claims, err := svc.ValidateTempToken(token)
	if err != nil {
		t.Fatalf("failed to validate temp token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("userID mismatch: got %v, want %v", claims.UserID, userID)
	}

	if claims.Email != email {
		t.Errorf("email mismatch: got %v, want %v", claims.Email, email)
	}

	if claims.Issuer != "identity-service-2fa" {
		t.Errorf("issuer mismatch: got %v, want identity-service-2fa", claims.Issuer)
	}
}

func TestValidateAccessTokenInvalid(t *testing.T) {
	svc := newTestTokenService()

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid token", "not-a-valid-token"},
		{"malformed jwt", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ValidateAccessToken(tt.token)
			if err == nil {
				t.Error("expected error for invalid token")
			}
			if err != ErrInvalidToken {
				t.Errorf("expected ErrInvalidToken, got %v", err)
			}
		})
	}
}

func TestValidateRefreshTokenInvalid(t *testing.T) {
	svc := newTestTokenService()

	_, err := svc.ValidateRefreshToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateTempTokenInvalid(t *testing.T) {
	svc := newTestTokenService()

	_, err := svc.ValidateTempToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAccessTokenWithDifferentSecret(t *testing.T) {
	svc1 := newTestTokenService()

	cfg2 := &config.Config{
		JWT: config.JWTConfig{
			Secret:               "different-secret-key-for-testing",
			AccessTokenSecret:    "different-access-secret-key-32!!",
			RefreshTokenSecret:   "different-refresh-secret-key-32!",
			TempTokenSecret:      "different-temp-secret-key-32!!!!",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 24 * time.Hour,
		},
	}
	svc2 := NewTokenService(cfg2, newMockBlacklist())

	token, err := svc1.GenerateAccessToken(utils.GenerateUUID(), "test@example.com", []string{"user"})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = svc2.ValidateAccessToken(token)
	if err == nil {
		t.Error("expected error when validating with different secret")
	}
}

func TestTokenClaims(t *testing.T) {
	svc := newTestTokenService()

	userID := utils.GenerateUUID()
	email := "user@example.com"
	roles := []string{"admin", "moderator"}

	token, _ := svc.GenerateAccessToken(userID, email, roles)
	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.Issuer != "identity-service" {
		t.Errorf("issuer mismatch: got %v, want identity-service", claims.Issuer)
	}

	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("token should not be expired")
	}

	if claims.IssuedAt.After(time.Now()) {
		t.Error("issued at should not be in the future")
	}
}
