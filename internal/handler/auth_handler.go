package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/creafly/identity/internal/config"
	"github.com/creafly/identity/internal/domain/service"
	"github.com/creafly/identity/internal/i18n"
	"github.com/creafly/identity/internal/middleware"
	"github.com/creafly/identity/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	cfg                      *config.Config
	userService              service.UserService
	tokenService             service.TokenService
	roleService              service.RoleService
	totpService              service.TOTPService
	passwordResetService     service.PasswordResetService
	emailVerificationService service.EmailVerificationService
	claimService             service.ClaimService
}

func NewAuthHandler(
	cfg *config.Config,
	userService service.UserService,
	tokenService service.TokenService,
	roleService service.RoleService,
	totpService service.TOTPService,
	passwordResetService service.PasswordResetService,
	emailVerificationService service.EmailVerificationService,
	claimService service.ClaimService,
) *AuthHandler {
	return &AuthHandler{
		cfg:                      cfg,
		userService:              userService,
		tokenService:             tokenService,
		roleService:              roleService,
		totpService:              totpService,
		passwordResetService:     passwordResetService,
		emailVerificationService: emailVerificationService,
		claimService:             claimService,
	}
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"omitempty,min=3,max=30,alphanum"`
	Password  string `json:"password" binding:"required,password"`
	FirstName string `json:"firstName" binding:"required,min=1,max=100"`
	LastName  string `json:"lastName" binding:"required,min=1,max=100"`
	Locale    string `json:"locale"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginVerifyTOTPRequest struct {
	TempToken string `json:"tempToken" binding:"required"`
	Code      string `json:"code" binding:"required,len=6"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,password"`
}

type UpdateProfileRequest struct {
	FirstName *string `json:"firstName" binding:"omitempty,min=1,max=100"`
	LastName  *string `json:"lastName" binding:"omitempty,min=1,max=100"`
	Username  *string `json:"username" binding:"omitempty,min=3,max=30,alphanum"`
	AvatarURL *string `json:"avatarUrl" binding:"omitempty,url"`
	Locale    *string `json:"locale" binding:"omitempty,oneof=en-US ru-RU"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,password"`
}

type VerifyEmailRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validator.HandleBindError(c, err) {
			return
		}
	}

	if req.Locale == "" {
		req.Locale = string(locale)
	}

	user, err := h.userService.Register(c.Request.Context(), service.RegisterInput{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Locale:    req.Locale,
	})
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": messages.Errors.UserAlreadyExists})
			return
		}
		if err == service.ErrUsernameAlreadyUsed {
			c.JSON(http.StatusConflict, gin.H{"error": messages.Errors.UsernameAlreadyUsed})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	if err := h.roleService.AssignDefaultUserRole(c.Request.Context(), user.ID); err != nil {
		log.Printf("Warning: Failed to assign default user role to user %s: %v", user.ID, err)
	}

	if err := h.emailVerificationService.RequestVerification(c.Request.Context(), user.ID); err != nil {
		log.Printf("Warning: Failed to send verification email to user %s: %v", user.ID, err)
	}

	accessToken, err := h.tokenService.GenerateAccessToken(user.ID, user.Email, []string{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": messages.Auth.RegisterSuccess,
		"user":    user,
		"tokens": TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(h.cfg.JWT.AccessTokenDuration).Unix(),
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validator.HandleBindError(c, err) {
			return
		}
	}

	user, err := h.userService.ValidateCredentials(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrUserBlocked {
			c.JSON(http.StatusForbidden, gin.H{"error": messages.Errors.UserBlocked})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.InvalidCredentials})
		return
	}

	totpEnabled, err := h.totpService.IsEnabled(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	if totpEnabled {
		tempToken, err := h.tokenService.GenerateTempToken(user.ID, user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"totpRequired": true,
			"tempToken":    tempToken,
			"message":      messages.TOTP.VerificationRequired,
		})
		return
	}

	accessToken, err := h.tokenService.GenerateAccessToken(user.ID, user.Email, []string{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": messages.Auth.LoginSuccess,
		"user":    user,
		"tokens": TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(h.cfg.JWT.AccessTokenDuration).Unix(),
		},
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validator.HandleBindError(c, err) {
			return
		}
	}

	claims, err := h.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.InvalidToken})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.UserNotFound})
		return
	}

	if user.IsBlocked {
		c.JSON(http.StatusForbidden, gin.H{"error": messages.Errors.UserBlocked})
		return
	}

	accessToken, err := h.tokenService.GenerateAccessToken(user.ID, user.Email, []string{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": messages.Auth.LoginSuccess,
		"user":    user,
		"tokens": TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(h.cfg.JWT.AccessTokenDuration).Unix(),
		},
	})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validator.HandleBindError(c, err) {
			return
		}
	}

	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	_ = h.passwordResetService.RequestPasswordReset(c.Request.Context(), req.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": messages.PasswordReset.RequestSent,
	})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validator.HandleBindError(c, err) {
			return
		}
	}

	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	err := h.passwordResetService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		if err == service.ErrPasswordResetTokenNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.PasswordReset.TokenNotFound})
			return
		}
		if err == service.ErrPasswordResetTokenExpired {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.PasswordReset.TokenExpired})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": messages.PasswordReset.PasswordReset,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": messages.Errors.UserNotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validator.HandleBindError(c, err) {
			return
		}
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID.(uuid.UUID), req.OldPassword, req.NewPassword)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.InvalidCredentials})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Auth.PasswordChanged})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validator.HandleBindError(c, err) {
			return
		}
	}

	user, err := h.userService.Update(c.Request.Context(), userID.(uuid.UUID), service.UpdateUserInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		AvatarURL: req.AvatarURL,
		Locale:    req.Locale,
	})
	if err != nil {
		if err == service.ErrUsernameAlreadyUsed {
			c.JSON(http.StatusConflict, gin.H{"error": messages.Errors.UsernameAlreadyUsed})
			return
		}
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Errors.UserNotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

type VerifyTokenResponse struct {
	Valid       bool       `json:"valid"`
	UserID      uuid.UUID  `json:"userId,omitempty"`
	Email       string     `json:"email,omitempty"`
	TenantID    *string    `json:"tenantId,omitempty"`
	Claims      []string   `json:"claims,omitempty"`
	IsBlocked   bool       `json:"isBlocked,omitempty"`
	BlockReason *string    `json:"blockReason,omitempty"`
	BlockedAt   *time.Time `json:"blockedAt,omitempty"`
}

func (h *AuthHandler) Verify(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid": false,
			"error": messages.Errors.Unauthorized,
		})
		return
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid": false,
			"error": messages.Errors.Unauthorized,
		})
		return
	}
	tokenString := authHeader[len(bearerPrefix):]

	claims, err := h.tokenService.ValidateAccessToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid": false,
			"error": messages.Errors.InvalidToken,
		})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid": false,
			"error": messages.Errors.UserNotFound,
		})
		return
	}

	if user.IsBlocked {
		c.JSON(http.StatusForbidden, VerifyTokenResponse{
			Valid:       false,
			UserID:      user.ID,
			Email:       user.Email,
			IsBlocked:   true,
			BlockReason: user.BlockReason,
			BlockedAt:   user.BlockedAt,
		})
		return
	}

	userClaims, err := h.claimService.GetUserAllClaims(c.Request.Context(), claims.UserID, nil)
	var claimValues []string
	if err == nil && userClaims != nil {
		claimValues = make([]string, 0, len(userClaims))
		for _, claim := range userClaims {
			claimValues = append(claimValues, claim.Value)
		}
	}

	c.JSON(http.StatusOK, VerifyTokenResponse{
		Valid:  true,
		UserID: claims.UserID,
		Email:  claims.Email,
		Claims: claimValues,
	})
}

func (h *AuthHandler) LoginVerifyTOTP(c *gin.Context) {
	var req LoginVerifyTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validator.HandleBindError(c, err) {
			return
		}
	}

	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	claims, err := h.tokenService.ValidateTempToken(req.TempToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.InvalidToken})
		return
	}

	valid, err := h.totpService.ValidateCode(c.Request.Context(), claims.UserID, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.TOTP.InvalidCode})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.UserNotFound})
		return
	}

	accessToken, err := h.tokenService.GenerateAccessToken(user.ID, user.Email, []string{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": messages.Auth.LoginSuccess,
		"user":    user,
		"tokens": TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(h.cfg.JWT.AccessTokenDuration).Unix(),
		},
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validator.HandleBindError(c, err) {
			return
		}
	}

	err := h.emailVerificationService.VerifyEmail(c.Request.Context(), userID.(uuid.UUID), req.Code)
	if err != nil {
		switch err {
		case service.ErrEmailAlreadyVerified:
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.EmailVerification.AlreadyVerified})
			return
		case service.ErrEmailVerificationCodeInvalid:
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.EmailVerification.InvalidCode})
			return
		case service.ErrEmailVerificationCodeExpired:
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.EmailVerification.CodeExpired})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": messages.EmailVerification.VerificationSuccess,
	})
}

func (h *AuthHandler) ResendVerificationEmail(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	err := h.emailVerificationService.ResendVerification(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		switch err {
		case service.ErrEmailAlreadyVerified:
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.EmailVerification.AlreadyVerified})
			return
		case service.ErrTooManyVerificationRequests:
			c.JSON(http.StatusTooManyRequests, gin.H{"error": messages.EmailVerification.TooManyRequests})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": messages.EmailVerification.CodeSent,
	})
}
