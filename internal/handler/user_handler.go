package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hexaend/identity/internal/domain/service"
	"github.com/hexaend/identity/internal/i18n"
	"github.com/hexaend/identity/internal/middleware"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type BlockUserRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type ListUsersResponse struct {
	Users  interface{} `json:"users"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
}

func (h *UserHandler) List(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	users, err := h.userService.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, ListUsersResponse{
		Users:  users,
		Offset: offset,
		Limit:  limit,
	})
}

func (h *UserHandler) GetByID(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Errors.UserNotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Block(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req BlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	blockedBy, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	blockedByID, ok := blockedBy.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	if err := h.userService.BlockUser(c.Request.Context(), id, req.Reason, blockedByID); err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Errors.UserNotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.User.Blocked})
}

func (h *UserHandler) Unblock(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.userService.UnblockUser(c.Request.Context(), id); err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Errors.UserNotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.User.Unblocked})
}
