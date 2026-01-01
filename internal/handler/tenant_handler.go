package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/domain/service"
	"github.com/creafly/identity/internal/i18n"
	"github.com/creafly/identity/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TenantHandler struct {
	tenantService     service.TenantService
	tenantRoleService service.TenantRoleService
	invitationService service.InvitationService
	userService       service.UserService
}

func NewTenantHandler(
	tenantService service.TenantService,
	tenantRoleService service.TenantRoleService,
	invitationService service.InvitationService,
	userService service.UserService,
) *TenantHandler {
	return &TenantHandler{
		tenantService:     tenantService,
		tenantRoleService: tenantRoleService,
		invitationService: invitationService,
		userService:       userService,
	}
}

type CreateTenantRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"displayName"`
	Slug        string `json:"slug"`
}

type UpdateTenantRequest struct {
	Name        *string `json:"name"`
	DisplayName *string `json:"displayName"`
	Slug        *string `json:"slug"`
	IsActive    *bool   `json:"isActive"`
}

type InviteMemberRequest struct {
	UserID   string `json:"userId"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type AddMemberCallbackRequest struct {
	TenantID string `json:"tenantId" binding:"required,uuid"`
	UserID   string `json:"userId" binding:"required,uuid"`
}

type BlockTenantRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *TenantHandler) Create(c *gin.Context) {
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

	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	tenant, err := h.tenantService.Create(c.Request.Context(), service.CreateTenantInput{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Slug:        req.Slug,
	})
	if err != nil {
		if err == service.ErrTenantAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": messages.Tenant.AlreadyExists})
			return
		}
		if err == service.ErrInvalidSlug {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Tenant.InvalidSlug})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	_ = h.tenantService.AddMember(c.Request.Context(), tenant.ID, userID)

	if err := h.tenantRoleService.CreateDefaultRoles(c.Request.Context(), tenant.ID); err != nil {
		log.Printf("[TenantHandler] Warning: Failed to create default roles for tenant %s: %v", tenant.ID, err)
	}

	_ = h.tenantRoleService.AssignOwnerRole(c.Request.Context(), userID, tenant.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": messages.Tenant.Created,
		"tenant":  tenant,
	})
}

func (h *TenantHandler) GetByID(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	tenant, err := h.tenantService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": messages.Tenant.NotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant": tenant,
	})
}

func (h *TenantHandler) List(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	tenants, err := h.tenantService.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, tenants)
}

func (h *TenantHandler) Update(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	tenant, err := h.tenantService.Update(c.Request.Context(), id, service.UpdateTenantInput{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Slug:        req.Slug,
		IsActive:    req.IsActive,
	})
	if err != nil {
		if err == service.ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Tenant.NotFound})
			return
		}
		if err == service.ErrTenantAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": messages.Tenant.AlreadyExists})
			return
		}
		if err == service.ErrInvalidSlug {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Tenant.InvalidSlug})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": messages.Tenant.Updated,
		"tenant":  tenant,
	})
}

func (h *TenantHandler) Delete(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.tenantService.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Tenant.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Tenant.Deleted})
}

func (h *TenantHandler) InviteMember(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	inviterIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}
	inviterID, ok := inviterIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var invitee *entity.User
	ctx := c.Request.Context()

	if req.Email != "" {
		invitee, err = h.userService.GetByEmail(ctx, req.Email)
	} else if req.Username != "" {
		invitee, err = h.userService.GetByUsername(ctx, req.Username)
	} else if req.UserID != "" {
		inviteeID, parseErr := uuid.Parse(req.UserID)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
			return
		}
		invitee, err = h.userService.GetByID(ctx, inviteeID)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err != nil || invitee == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": messages.Errors.UserNotFound})
		return
	}

	tenant, err := h.tenantService.GetByID(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": messages.Tenant.NotFound})
		return
	}

	members, err := h.tenantService.GetMembers(ctx, tenantID)
	if err == nil {
		for _, member := range members {
			if member.ID == invitee.ID {
				c.JSON(http.StatusConflict, gin.H{"error": messages.Tenant.AlreadyMember})
				return
			}
		}
	}

	inviter, err := h.userService.GetByID(ctx, inviterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	err = h.invitationService.RequestInvitation(ctx, service.RequestInvitationInput{
		TenantID:    tenantID,
		TenantName:  tenant.Name,
		InviterID:   inviterID,
		InviterName: inviter.FirstName + " " + inviter.LastName,
		InviteeID:   invitee.ID,
		Email:       invitee.Email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Tenant.InvitationFailed})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": messages.Tenant.InvitationSent,
	})
}

func (h *TenantHandler) AddMemberCallback(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	serviceName := c.GetHeader("X-Service-Name")
	if serviceName != "notifications" {
		c.JSON(http.StatusForbidden, gin.H{"error": messages.Errors.Forbidden})
		return
	}

	var req AddMemberCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	tenantID, _ := uuid.Parse(req.TenantID)
	userID, _ := uuid.Parse(req.UserID)

	if err := h.tenantService.AddMember(c.Request.Context(), tenantID, userID); err != nil {
		if err == service.ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Tenant.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	memberRole, err := h.tenantRoleService.GetByName(c.Request.Context(), tenantID, "member")
	if err == nil {
		_ = h.tenantRoleService.AssignToUser(c.Request.Context(), userID, tenantID, memberRole.ID)
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Tenant.CallbackSuccess})
}

func (h *TenantHandler) RemoveMember(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.tenantService.RemoveMember(c.Request.Context(), tenantID, userID); err != nil {
		if err == service.ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Tenant.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Tenant.MemberRemoved})
}

func (h *TenantHandler) GetMembers(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	members, err := h.tenantService.GetMembers(c.Request.Context(), tenantID)
	if err != nil {
		if err == service.ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Tenant.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, members)
}

func (h *TenantHandler) GetUserRoles(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	roles, err := h.tenantRoleService.GetUserRoles(c.Request.Context(), userID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}

func (h *TenantHandler) AssignRolesToTenantUser(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req struct {
		RoleIDs []string `json:"roleIds" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var roleUUIDs []uuid.UUID
	for _, roleIDStr := range req.RoleIDs {
		roleID, err := uuid.Parse(roleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
			return
		}
		roleUUIDs = append(roleUUIDs, roleID)
	}

	for _, roleID := range roleUUIDs {
		if err := h.tenantRoleService.AssignToUser(c.Request.Context(), userID, tenantID, roleID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Roles assigned successfully"})
}

func (h *TenantHandler) RemoveRolesFromTenantUser(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	currentUserID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req struct {
		RoleIDs []string `json:"roleIds" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var roleUUIDs []uuid.UUID
	for _, roleIDStr := range req.RoleIDs {
		roleID, err := uuid.Parse(roleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
			return
		}
		roleUUIDs = append(roleUUIDs, roleID)
	}

	if currentUserID == targetUserID {
		ownerRole, err := h.tenantRoleService.GetByName(c.Request.Context(), tenantID, "owner")
		if err == nil {
			for _, roleID := range roleUUIDs {
				if roleID == ownerRole.ID {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove owner role from yourself"})
					return
				}
			}
		}
	}

	for _, roleID := range roleUUIDs {
		if err := h.tenantRoleService.RemoveFromUser(c.Request.Context(), targetUserID, tenantID, roleID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Roles removed successfully"})
}

func (h *TenantHandler) AssignRoleToTenantUser(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

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

	if err := h.tenantRoleService.AssignToUser(c.Request.Context(), userID, tenantID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned successfully"})
}

func (h *TenantHandler) RemoveRoleFromTenantUser(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	currentUserID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if currentUserID == targetUserID {
		ownerRole, err := h.tenantRoleService.GetByName(c.Request.Context(), tenantID, "owner")
		if err == nil && roleID == ownerRole.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove owner role from yourself"})
			return
		}
	}

	if err := h.tenantRoleService.RemoveFromUser(c.Request.Context(), targetUserID, tenantID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed successfully"})
}

func (h *TenantHandler) BatchUpdateTenantUserRoles(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	currentUserID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req struct {
		AssignRoleIDs []string `json:"assignRoleIds,omitempty"`
		RemoveRoleIDs []string `json:"removeRoleIds,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var assignRoleUUIDs []uuid.UUID
	for _, roleIDStr := range req.AssignRoleIDs {
		roleID, err := uuid.Parse(roleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
			return
		}
		assignRoleUUIDs = append(assignRoleUUIDs, roleID)
	}

	var removeRoleUUIDs []uuid.UUID
	for _, roleIDStr := range req.RemoveRoleIDs {
		roleID, err := uuid.Parse(roleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
			return
		}
		removeRoleUUIDs = append(removeRoleUUIDs, roleID)
	}

	if currentUserID == targetUserID {
		ownerRole, err := h.tenantRoleService.GetByName(c.Request.Context(), tenantID, "owner")
		if err == nil {
			for _, roleID := range removeRoleUUIDs {
				if roleID == ownerRole.ID {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove owner role from yourself"})
					return
				}
			}
		}
	}

	for _, roleID := range assignRoleUUIDs {
		if err := h.tenantRoleService.AssignToUser(c.Request.Context(), targetUserID, tenantID, roleID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
			return
		}
	}

	for _, roleID := range removeRoleUUIDs {
		if err := h.tenantRoleService.RemoveFromUser(c.Request.Context(), targetUserID, tenantID, roleID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Roles updated successfully"})
}

func (h *TenantHandler) GetMyTenants(c *gin.Context) {
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

	tenants, err := h.tenantService.GetUserTenants(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": tenants,
	})
}

func (h *TenantHandler) ResolveSlug(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	tenant, err := h.tenantService.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": messages.Tenant.NotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          tenant.ID.String(),
		"slug":        tenant.Slug,
		"name":        tenant.Name,
		"displayName": tenant.DisplayName,
		"isActive":    tenant.IsActive,
	})
}

func (h *TenantHandler) ValidateTenantAccess(c *gin.Context) {
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

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	tenant, err := h.tenantService.GetByID(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":    false,
			"isMember": false,
			"reason":   "tenant_not_found",
		})
		return
	}

	if !tenant.IsActive {
		c.JSON(http.StatusOK, gin.H{
			"valid":    false,
			"isMember": false,
			"reason":   "tenant_inactive",
		})
		return
	}

	isMember, err := h.tenantService.IsMember(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    isMember,
		"isMember": isMember,
		"tenantId": tenantID.String(),
		"userId":   userID.String(),
	})
}

func (h *TenantHandler) Block(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req BlockTenantRequest
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

	if err := h.tenantService.BlockTenant(c.Request.Context(), id, req.Reason, blockedByID); err != nil {
		if err == service.ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Tenant.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Tenant.Blocked})
}

func (h *TenantHandler) Unblock(c *gin.Context) {
	locale := middleware.GetLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.tenantService.UnblockTenant(c.Request.Context(), id); err != nil {
		if err == service.ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Tenant.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Tenant.Unblocked})
}
