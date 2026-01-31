package middleware

import (
	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware creates an admin authorization middleware.
// This middleware must be used AFTER AuthMiddleware.
func AdminMiddleware(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthMiddleware)
		userID, ok := GetUserID(c)
		if !ok {
			response.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}

		// Get user from database
		user, err := userService.GetProfile(userID)
		if err != nil {
			response.Unauthorized(c, "user not found")
			c.Abort()
			return
		}

		// Check if user is admin
		if !user.IsAdmin() {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}
