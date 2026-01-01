package middleware

import (
	"strings"

	"github.com/creafly/identity/internal/domain/repository"
	"github.com/creafly/identity/internal/domain/service"
	"github.com/creafly/identity/internal/i18n"
	sharedmw "github.com/creafly/middleware"
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

		tokenString := parts[1]

		if tokenService.IsTokenRevoked(tokenString) {
			c.JSON(401, gin.H{"error": "Token has been revoked", "code": "TOKEN_REVOKED"})
			c.Abort()
			return
		}

		claims, err := tokenService.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if tokenService.IsUserTokensRevoked(claims.UserID) {
			c.JSON(401, gin.H{"error": "All user tokens have been revoked", "code": "USER_TOKENS_REVOKED"})
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

func GetLocale(c *gin.Context) i18n.Locale {
	return i18n.ParseLocale(sharedmw.GetLocale(c))
}
