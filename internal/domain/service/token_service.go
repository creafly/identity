package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/creafly/identity/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrExpiredToken         = errors.New("token expired")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrTokenRevoked         = errors.New("token has been revoked")
	ErrFingerprintMismatch  = errors.New("token fingerprint mismatch")
)

type TokenFingerprint struct {
	UserAgent string
	IP        string
}

func GenerateFingerprint(fp TokenFingerprint) string {
	if fp.UserAgent == "" && fp.IP == "" {
		return ""
	}
	data := fp.UserAgent + "|" + fp.IP
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

type TokenService interface {
	GenerateAccessToken(userID uuid.UUID, email string, roles []string) (string, error)
	GenerateAccessTokenWithFingerprint(userID uuid.UUID, email string, roles []string, fingerprint string) (string, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	GenerateRefreshTokenWithFingerprint(userID uuid.UUID, fingerprint string) (string, error)
	GenerateTempToken(userID uuid.UUID, email string) (string, error)
	ValidateAccessToken(tokenString string) (*AccessTokenClaims, error)
	ValidateAccessTokenWithFingerprint(tokenString string, fingerprint string) (*AccessTokenClaims, error)
	ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error)
	ValidateRefreshTokenWithFingerprint(tokenString string, fingerprint string) (*RefreshTokenClaims, error)
	ValidateTempToken(tokenString string) (*TempTokenClaims, error)
	RevokeToken(tokenString string, expiresAt time.Time)
	RevokeAllUserTokens(userID uuid.UUID, expiresAt time.Time)
	IsTokenRevoked(tokenString string) bool
	IsUserTokensRevoked(userID uuid.UUID) bool
}

type AccessTokenClaims struct {
	UserID      uuid.UUID `json:"userId"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	Fingerprint string    `json:"fgp,omitempty"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	UserID      uuid.UUID `json:"userId"`
	Fingerprint string    `json:"fgp,omitempty"`
	jwt.RegisteredClaims
}

type TempTokenClaims struct {
	UserID uuid.UUID `json:"userId"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

type tokenService struct {
	cfg       *config.Config
	blacklist TokenBlacklist
}

func NewTokenService(cfg *config.Config, blacklist TokenBlacklist) TokenService {
	return &tokenService{
		cfg:       cfg,
		blacklist: blacklist,
	}
}

func NewTokenServiceWithBlacklist(cfg *config.Config, blacklist TokenBlacklist) TokenService {
	return NewTokenService(cfg, blacklist)
}

func (s *tokenService) GenerateAccessToken(userID uuid.UUID, email string, roles []string) (string, error) {
	return s.GenerateAccessTokenWithFingerprint(userID, email, roles, "")
}

func (s *tokenService) GenerateAccessTokenWithFingerprint(userID uuid.UUID, email string, roles []string, fingerprint string) (string, error) {
	claims := AccessTokenClaims{
		UserID:      userID,
		Email:       email,
		Roles:       roles,
		Fingerprint: fingerprint,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWT.AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "identity-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.AccessTokenSecret))
}

func (s *tokenService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	return s.GenerateRefreshTokenWithFingerprint(userID, "")
}

func (s *tokenService) GenerateRefreshTokenWithFingerprint(userID uuid.UUID, fingerprint string) (string, error) {
	claims := RefreshTokenClaims{
		UserID:      userID,
		Fingerprint: fingerprint,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWT.RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "identity-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.RefreshTokenSecret))
}

func (s *tokenService) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	return s.ValidateAccessTokenWithFingerprint(tokenString, "")
}

func (s *tokenService) ValidateAccessTokenWithFingerprint(tokenString string, fingerprint string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(s.cfg.JWT.AccessTokenSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Fingerprint != "" && fingerprint != "" && claims.Fingerprint != fingerprint {
		return nil, ErrFingerprintMismatch
	}

	return claims, nil
}

func (s *tokenService) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	return s.ValidateRefreshTokenWithFingerprint(tokenString, "")
}

func (s *tokenService) ValidateRefreshTokenWithFingerprint(tokenString string, fingerprint string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(s.cfg.JWT.RefreshTokenSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Fingerprint != "" && fingerprint != "" && claims.Fingerprint != fingerprint {
		return nil, ErrFingerprintMismatch
	}

	return claims, nil
}

func (s *tokenService) GenerateTempToken(userID uuid.UUID, email string) (string, error) {
	claims := TempTokenClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "identity-service-2fa",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.TempTokenSecret))
}

func (s *tokenService) ValidateTempToken(tokenString string) (*TempTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TempTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(s.cfg.JWT.TempTokenSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TempTokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Issuer != "identity-service-2fa" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *tokenService) RevokeToken(tokenString string, expiresAt time.Time) {
	tokenHash := HashToken(tokenString)
	s.blacklist.Add(tokenHash, expiresAt)
}

func (s *tokenService) RevokeAllUserTokens(userID uuid.UUID, expiresAt time.Time) {
	s.blacklist.RevokeAllForUser(userID.String(), expiresAt)
}

func (s *tokenService) IsTokenRevoked(tokenString string) bool {
	tokenHash := HashToken(tokenString)
	return s.blacklist.IsBlacklisted(tokenHash)
}

func (s *tokenService) IsUserTokensRevoked(userID uuid.UUID) bool {
	return s.blacklist.IsUserRevoked(userID.String())
}
