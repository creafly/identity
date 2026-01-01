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

type ClaimHandler struct {
	claimService  service.ClaimService
	tenantService service.TenantService
}

func NewClaimHandler(claimService service.ClaimService, tenantService service.TenantService) *ClaimHandler {
	return &ClaimHandler{
		claimService:  claimService,
		tenantService: tenantService,
	}
}

type CreateClaimRequest struct {
	Value string `json:"value" binding:"required"`
}

type AssignClaimRequest struct {
	ClaimID string `json:"claimId" binding:"required,uuid"`
}

func (h *ClaimHandler) Create(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	var req CreateClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	claim, err := h.claimService.Create(c.Request.Context(), service.CreateClaimInput{
		Value: req.Value,
	})
	if err != nil {
		if err == service.ErrClaimAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": messages.Claim.AlreadyExists})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": messages.Claim.Created,
		"claim":   claim,
	})
}

func (h *ClaimHandler) GetByID(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	claim, err := h.claimService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": messages.Claim.NotFound})
		return
	}

	c.JSON(http.StatusOK, claim)
}

func (h *ClaimHandler) List(c *gin.Context) {
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

	claims, err := h.claimService.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"claims": claims,
		"offset": offset,
		"limit":  limit,
	})
}

func (h *ClaimHandler) Delete(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.claimService.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrClaimNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Claim.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Claim.Deleted})
}

func (h *ClaimHandler) AssignToUser(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req AssignClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	claimID, _ := uuid.Parse(req.ClaimID)

	if err := h.claimService.AssignToUser(c.Request.Context(), userID, claimID); err != nil {
		if err == service.ErrClaimNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Claim.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Claim.Assigned})
}

func (h *ClaimHandler) RemoveFromUser(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	claimID, err := uuid.Parse(c.Param("claimId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.claimService.RemoveFromUser(c.Request.Context(), userID, claimID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Claim.Unassigned})
}

func (h *ClaimHandler) GetUserClaims(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	claims, err := h.claimService.GetUserClaims(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, claims)
}

func (h *ClaimHandler) AssignToRole(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req AssignClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	claimID, _ := uuid.Parse(req.ClaimID)

	if err := h.claimService.AssignToRole(c.Request.Context(), roleID, claimID); err != nil {
		if err == service.ErrClaimNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Claim.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Claim.Assigned})
}

func (h *ClaimHandler) RemoveFromRole(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	claimID, err := uuid.Parse(c.Param("claimId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.claimService.RemoveFromRole(c.Request.Context(), roleID, claimID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Claim.Unassigned})
}

func (h *ClaimHandler) GetRoleClaims(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	claims, err := h.claimService.GetRoleClaims(c.Request.Context(), roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, claims)
}

func (h *ClaimHandler) GetMyClaims(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var tenantID *uuid.UUID
	tenantIDStr := c.Query("tenantId")
	if tenantIDStr == "" {
		tenantIDStr = c.GetHeader("X-Tenant-ID")
	}

	if tenantIDStr != "" {
		parsed, err := uuid.Parse(tenantIDStr)
		if err != nil {
			tenant, err := h.tenantService.GetBySlug(c.Request.Context(), tenantIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
				return
			}
			tenantID = &tenant.ID
		} else {
			tenantID = &parsed
		}
	}

	claims, err := h.claimService.GetUserAllClaims(c.Request.Context(), userID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	permissions := make([]string, 0, len(claims))
	for _, claim := range claims {
		permissions = append(permissions, claim.Value)
	}

	c.JSON(http.StatusOK, gin.H{
		"claims":      claims,
		"permissions": permissions,
	})
}
