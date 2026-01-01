package middleware

import (
	"strings"

	"github.com/creafly/identity/internal/domain/repository"
	"github.com/creafly/identity/internal/domain/service"
	"github.com/creafly/identity/internal/i18n"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthMiddleware(tokenService service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := tokenService.ValidateAccessToken(parts[1])
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

func BlockedUserMiddleware(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := GetLocale(c)
		messages := i18n.GetMessages(locale)

		userIDStr, exists := c.Get("userID")
		if !exists {
			c.JSON(401, gin.H{"error": messages.Errors.Unauthorized})
			c.Abort()
			return
		}

		userID, ok := userIDStr.(uuid.UUID)
		if !ok {
			c.JSON(401, gin.H{"error": messages.Errors.Unauthorized})
			c.Abort()
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(401, gin.H{"error": messages.Errors.UserNotFound})
			c.Abort()
			return
		}

		if user.IsBlocked {
			c.JSON(403, gin.H{
				"error":       messages.Errors.UserBlocked,
				"code":        "USER_BLOCKED",
				"isBlocked":   true,
				"blockReason": user.BlockReason,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
