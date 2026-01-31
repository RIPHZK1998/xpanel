package middleware

import (
	"net/http"
	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// NodeAuth creates middleware to authenticate node agent requests
func NodeAuth(configService *service.SystemConfigService) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		if apiKey == "" {
			response.Error(c, http.StatusUnauthorized, "missing API key")
			c.Abort()
			return
		}

		// Get the configured node API key (from cache or database)
		expectedKey, err := configService.GetNodeApiKey()
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to verify API key")
			c.Abort()
			return
		}

		// If no key is configured yet, allow (for initial setup)
		if expectedKey == "" {
			c.Next()
			return
		}

		if apiKey != expectedKey {
			response.Error(c, http.StatusUnauthorized, "invalid API key")
			c.Abort()
			return
		}

		c.Next()
	}
}
