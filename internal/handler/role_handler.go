package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/creafly/identity/internal/domain/service"
	"github.com/creafly/identity/internal/i18n"
	"github.com/creafly/identity/internal/middleware"
)

type RoleHandler struct {
	roleService service.RoleService
}

func NewRoleHandler(roleService service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type AssignRoleRequest struct {
	RoleID string `json:"roleId" binding:"required,uuid"`
}

func (h *RoleHandler) Create(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	role, err := h.roleService.Create(c.Request.Context(), service.CreateRoleInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if err == service.ErrRoleAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": messages.Role.AlreadyExists})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": messages.Role.Created,
		"role":    role,
	})
}

func (h *RoleHandler) GetByID(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	role, err := h.roleService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": messages.Role.NotFound})
		return
	}

	c.JSON(http.StatusOK, role)
}

func (h *RoleHandler) List(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	offset := 0
	limit := 10

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	roles, err := h.roleService.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles":  roles,
		"offset": offset,
		"limit":  limit,
	})
}

func (h *RoleHandler) Update(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	role, err := h.roleService.Update(c.Request.Context(), id, service.UpdateRoleInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if err == service.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Role.NotFound})
			return
		}
		if err == service.ErrRoleAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": messages.Role.AlreadyExists})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": messages.Role.Updated,
		"role":    role,
	})
}

func (h *RoleHandler) Delete(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.roleService.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Role.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Role.Deleted})
}

func (h *RoleHandler) AssignToUser(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	roleID, _ := uuid.Parse(req.RoleID)

	if err := h.roleService.AssignToUser(c.Request.Context(), userID, roleID); err != nil {
		if err == service.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Role.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Role.Assigned})
}

func (h *RoleHandler) RemoveFromUser(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.roleService.RemoveFromUser(c.Request.Context(), userID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Role.Unassigned})
}

func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	roles, err := h.roleService.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, roles)
}
