package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/creafly/identity/internal/domain/service"
	"github.com/creafly/identity/internal/i18n"
	"github.com/creafly/identity/internal/logger"
)

func resolveTenantID(c *gin.Context, tenantService service.TenantService, identifier string) *uuid.UUID {
	if identifier == "" {
		logger.Debug().Msg("[Claims] resolveTenantID: identifier is empty")
		return nil
	}

	parsed, err := uuid.Parse(identifier)
	if err == nil {
		logger.Debug().Str("tenantID", parsed.String()).Msg("[Claims] resolveTenantID: parsed as UUID")
		return &parsed
	}

	if tenantService != nil {
		tenant, err := tenantService.GetBySlug(c.Request.Context(), identifier)
		if err == nil && tenant != nil {
			logger.Debug().Str("slug", identifier).Str("tenantID", tenant.ID.String()).Msg("[Claims] resolveTenantID: resolved slug to UUID")
			return &tenant.ID
		}
		logger.Debug().Str("slug", identifier).Err(err).Msg("[Claims] resolveTenantID: failed to resolve slug")
	} else {
		logger.Debug().Str("identifier", identifier).Msg("[Claims] resolveTenantID: tenantService is nil, cannot resolve slug")
	}

	return nil
}

func RequireClaims(claimService service.ClaimService, requiredClaims ...string) gin.HandlerFunc {
	return RequireClaimsWithTenant(claimService, nil, requiredClaims...)
}

func RequireClaimsWithTenant(claimService service.ClaimService, tenantService service.TenantService, requiredClaims ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := GetLocale(c)
		messages := i18n.GetMessages(locale)

		userIDVal, exists := c.Get("userID")
		if !exists {
			logger.Debug().Msg("[Claims] RequireClaimsWithTenant: userID not found in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
			c.Abort()
			return
		}

		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			logger.Debug().Msg("[Claims] RequireClaimsWithTenant: userID is not UUID")
			c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
			c.Abort()
			return
		}

		tenantIDStr := c.GetHeader("X-Tenant-ID")
		if tenantIDStr == "" {
			tenantIDStr = c.Param("id")
		}

		logger.Debug().Str("userID", userID.String()).Str("tenantIDStr", tenantIDStr).Strs("requiredClaims", requiredClaims).Msg("[Claims] RequireClaimsWithTenant: checking claims")

		tenantID := resolveTenantID(c, tenantService, tenantIDStr)

		if tenantID == nil {
			logger.Debug().Msg("[Claims] RequireClaimsWithTenant: tenantID is nil after resolution")
		} else {
			logger.Debug().Str("tenantID", tenantID.String()).Msg("[Claims] RequireClaimsWithTenant: tenantID resolved")
		}

		userClaims, err := claimService.GetUserAllClaims(c.Request.Context(), userID, tenantID)
		if err != nil {
			logger.Error().Err(err).Msg("[Claims] RequireClaimsWithTenant: failed to get user claims")
			c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
			c.Abort()
			return
		}

		userClaimSet := make(map[string]bool)
		claimValues := make([]string, 0, len(userClaims))
		for _, claim := range userClaims {
			userClaimSet[claim.Value] = true
			claimValues = append(claimValues, claim.Value)
		}

		logger.Debug().Strs("userClaims", claimValues).Msg("[Claims] RequireClaimsWithTenant: user claims fetched")

		for _, requiredClaim := range requiredClaims {
			if !userClaimSet[requiredClaim] {
				logger.Debug().Str("missingClaim", requiredClaim).Msg("[Claims] RequireClaimsWithTenant: missing required claim")
				c.JSON(http.StatusForbidden, gin.H{"error": messages.Errors.Forbidden})
				c.Abort()
				return
			}
		}

		c.Set("userClaims", userClaimSet)
		c.Set("tenantID", tenantID)

		c.Next()
	}
}

func RequireAnyClaim(claimService service.ClaimService, requiredClaims ...string) gin.HandlerFunc {
	return RequireAnyClaimWithTenant(claimService, nil, requiredClaims...)
}

func RequireAnyClaimWithTenant(claimService service.ClaimService, tenantService service.TenantService, requiredClaims ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := GetLocale(c)
		messages := i18n.GetMessages(locale)

		userIDVal, exists := c.Get("userID")
		if !exists {
			logger.Debug().Msg("[Claims] RequireAnyClaimWithTenant: userID not found in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
			c.Abort()
			return
		}

		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			logger.Debug().Msg("[Claims] RequireAnyClaimWithTenant: userID is not UUID")
			c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
			c.Abort()
			return
		}

		tenantIDStr := c.GetHeader("X-Tenant-ID")
		if tenantIDStr == "" {
			tenantIDStr = c.Param("id")
		}

		logger.Debug().Str("userID", userID.String()).Str("tenantIDStr", tenantIDStr).Strs("requiredClaims", requiredClaims).Msg("[Claims] RequireAnyClaimWithTenant: checking claims")

		tenantID := resolveTenantID(c, tenantService, tenantIDStr)

		if tenantID == nil {
			logger.Debug().Msg("[Claims] RequireAnyClaimWithTenant: tenantID is nil after resolution")
		} else {
			logger.Debug().Str("tenantID", tenantID.String()).Msg("[Claims] RequireAnyClaimWithTenant: tenantID resolved")
		}

		userClaims, err := claimService.GetUserAllClaims(c.Request.Context(), userID, tenantID)
		if err != nil {
			logger.Error().Err(err).Msg("[Claims] RequireAnyClaimWithTenant: failed to get user claims")
			c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
			c.Abort()
			return
		}

		userClaimSet := make(map[string]bool)
		claimValues := make([]string, 0, len(userClaims))
		for _, claim := range userClaims {
			userClaimSet[claim.Value] = true
			claimValues = append(claimValues, claim.Value)
		}

		logger.Debug().Strs("userClaims", claimValues).Msg("[Claims] RequireAnyClaimWithTenant: user claims fetched")

		hasAnyClaim := false
		for _, requiredClaim := range requiredClaims {
			if userClaimSet[requiredClaim] {
				hasAnyClaim = true
				break
			}
		}

		if !hasAnyClaim {
			logger.Debug().Strs("requiredClaims", requiredClaims).Msg("[Claims] RequireAnyClaimWithTenant: user has none of the required claims")
			c.JSON(http.StatusForbidden, gin.H{"error": messages.Errors.Forbidden})
			c.Abort()
			return
		}

		c.Set("userClaims", userClaimSet)
		c.Set("tenantID", tenantID)

		c.Next()
	}
}

func HasClaim(c *gin.Context, claim string) bool {
	claimsVal, exists := c.Get("userClaims")
	if !exists {
		return false
	}

	claims, ok := claimsVal.(map[string]bool)
	if !ok {
		return false
	}

	return claims[claim]
}
