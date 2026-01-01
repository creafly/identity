package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hexaend/identity/internal/domain/service"
	"github.com/hexaend/identity/internal/i18n"
	"github.com/hexaend/identity/internal/middleware"
)

type TenantRoleHandler struct {
	tenantRoleService service.TenantRoleService
}

func NewTenantRoleHandler(tenantRoleService service.TenantRoleService) *TenantRoleHandler {
	return &TenantRoleHandler{tenantRoleService: tenantRoleService}
}

type CreateTenantRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateTenantRoleRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type AssignTenantRoleClaimRequest struct {
	ClaimID string `json:"claimId" binding:"required,uuid"`
}

type BatchUpdateClaimsRequest struct {
	AssignClaimIDs []string `json:"assignClaimIds"`
	RemoveClaimIDs []string `json:"removeClaimIds"`
}

func (h *TenantRoleHandler) Create(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req CreateTenantRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	role, err := h.tenantRoleService.Create(c.Request.Context(), service.CreateTenantRoleInput{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if err == service.ErrTenantRoleAlreadyExists {
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

func (h *TenantRoleHandler) GetByID(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	role, err := h.tenantRoleService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": messages.Role.NotFound})
		return
	}

	c.JSON(http.StatusOK, role)
}

func (h *TenantRoleHandler) List(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	roles, err := h.tenantRoleService.ListByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, roles)
}

func (h *TenantRoleHandler) Update(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req UpdateTenantRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	role, err := h.tenantRoleService.Update(c.Request.Context(), id, service.UpdateTenantRoleInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if err == service.ErrTenantRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Role.NotFound})
			return
		}
		if err == service.ErrTenantRoleAlreadyExists {
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

func (h *TenantRoleHandler) Delete(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.tenantRoleService.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrTenantRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Role.NotFound})
			return
		}
		if err == service.ErrCannotDeleteDefaultRole {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete default role"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Role.Deleted})
}

func (h *TenantRoleHandler) AssignClaim(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req AssignTenantRoleClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	claimID, _ := uuid.Parse(req.ClaimID)

	if err := h.tenantRoleService.AddClaim(c.Request.Context(), roleID, claimID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Claim assigned successfully"})
}

func (h *TenantRoleHandler) RemoveClaim(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	claimID, err := uuid.Parse(c.Param("claimId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.tenantRoleService.RemoveClaim(c.Request.Context(), roleID, claimID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Claim removed successfully"})
}

func (h *TenantRoleHandler) GetRoleClaims(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	claims, err := h.tenantRoleService.GetRoleClaims(c.Request.Context(), roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, claims)
}

func (h *TenantRoleHandler) GetAvailableClaims(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	claims, err := h.tenantRoleService.GetAvailableClaims(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, claims)
}

func (h *TenantRoleHandler) BatchUpdateClaims(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req BatchUpdateClaimsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var assignClaimUUIDs []uuid.UUID
	for _, claimIDStr := range req.AssignClaimIDs {
		claimID, err := uuid.Parse(claimIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
			return
		}
		assignClaimUUIDs = append(assignClaimUUIDs, claimID)
	}

	var removeClaimUUIDs []uuid.UUID
	for _, claimIDStr := range req.RemoveClaimIDs {
		claimID, err := uuid.Parse(claimIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
			return
		}
		removeClaimUUIDs = append(removeClaimUUIDs, claimID)
	}

	if err := h.tenantRoleService.BatchUpdateClaims(c.Request.Context(), roleID, assignClaimUUIDs, removeClaimUUIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Claims updated successfully"})
}
