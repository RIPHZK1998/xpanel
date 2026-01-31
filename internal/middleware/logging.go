package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger creates a logging middleware.
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Get request details
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Log fields
		fields := logrus.Fields{
			"status":     statusCode,
			"method":     method,
			"path":       path,
			"ip":         clientIP,
			"latency_ms": latency.Milliseconds(),
			"user_agent": userAgent,
		}

		// Add user info if available
		if userID, exists := c.Get("user_id"); exists {
			fields["user_id"] = userID
		}
		if email, exists := c.Get("email"); exists {
			fields["email"] = email
		}

		// Add error if exists
		if len(c.Errors) > 0 {
			fields["error"] = c.Errors.String()
		}

		// Log based on status code
		if statusCode >= 500 {
			logger.WithFields(fields).Error("Server error")
		} else if statusCode >= 400 {
			logger.WithFields(fields).Warn("Client error")
		} else {
			logger.WithFields(fields).Info("Request processed")
		}
	}
}
