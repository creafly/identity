package handler

import (
	"net/http"

	"github.com/creafly/identity/internal/domain/service"
	"github.com/creafly/identity/internal/i18n"
	"github.com/creafly/identity/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TOTPHandler struct {
	totpService service.TOTPService
	userService service.UserService
}

func NewTOTPHandler(totpService service.TOTPService, userService service.UserService) *TOTPHandler {
	return &TOTPHandler{
		totpService: totpService,
		userService: userService,
	}
}

type EnableTOTPRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

type DisableTOTPRequest struct {
	Password string `json:"password" binding:"required"`
}

type ValidateTOTPRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

func (h *TOTPHandler) Setup(c *gin.Context) {
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

	response, err := h.totpService.GenerateSecret(c.Request.Context(), userID.(uuid.UUID), user.Email, "Creafly")
	if err != nil {
		if err == service.ErrTOTPAlreadyEnabled {
			c.JSON(http.StatusConflict, gin.H{"error": messages.TOTP.AlreadyEnabled})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": messages.TOTP.SetupSuccess,
		"data":    response,
	})
}

func (h *TOTPHandler) Enable(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	var req EnableTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	err := h.totpService.Enable(c.Request.Context(), userID.(uuid.UUID), req.Code)
	if err != nil {
		switch err {
		case service.ErrTOTPAlreadyEnabled:
			c.JSON(http.StatusConflict, gin.H{"error": messages.TOTP.AlreadyEnabled})
		case service.ErrTOTPNotSetup:
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.TOTP.NotSetup})
		case service.ErrTOTPInvalidCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.TOTP.InvalidCode})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": messages.TOTP.EnabledSuccess,
	})
}

func (h *TOTPHandler) Disable(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	var req DisableTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	err := h.totpService.Disable(c.Request.Context(), userID.(uuid.UUID), req.Password)
	if err != nil {
		switch err {
		case service.ErrTOTPNotEnabled:
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.TOTP.NotEnabled})
		case service.ErrInvalidCredentials:
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.InvalidCredentials})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": messages.TOTP.DisabledSuccess,
	})
}

func (h *TOTPHandler) Validate(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	var req ValidateTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	valid, err := h.totpService.ValidateCode(c.Request.Context(), userID.(uuid.UUID), req.Code)
	if err != nil {
		if err == service.ErrTOTPSecretNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.TOTP.NotSetup})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": messages.TOTP.InvalidCode,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
	})
}

func (h *TOTPHandler) Status(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	enabled, err := h.totpService.IsEnabled(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled": enabled,
	})
}
