package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hexaend/identity/internal/i18n"
)

const LocaleContextKey = "locale"

func LocaleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := c.GetHeader("Accept-Language")
		if locale == "" {
			locale = c.Query("locale")
		}

		parsedLocale := i18n.ParseLocale(locale)
		c.Set(LocaleContextKey, parsedLocale)

		c.Next()
	}
}

func GetLocale(c *gin.Context) i18n.Locale {
	if locale, exists := c.Get(LocaleContextKey); exists {
		return locale.(i18n.Locale)
	}
	return i18n.DefaultLocale
}
