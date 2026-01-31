// Package handler contains HTTP request handlers.
package handler

import (
	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles user registration.
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		if err == service.ErrEmailTaken {
			response.Conflict(c, "email already registered")
			return
		}
		response.InternalServerError(c, "failed to register user")
		return
	}

	response.Created(c, "user registered successfully", gin.H{
		"user": user.ToResponse(),
	})
}

// Login handles user authentication.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	tokens, user, err := h.authService.Login(&req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			response.Unauthorized(c, "invalid email or password")
			return
		}
		response.InternalServerError(c, "login failed")
		return
	}

	response.OK(c, "login successful", gin.H{
		"user":   user.ToResponse(),
		"tokens": tokens,
	})
}

// Refresh handles token refresh.
// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	tokens, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if err == service.ErrTokenBlacklisted || err == service.ErrInvalidToken {
			response.Unauthorized(c, "invalid or expired refresh token")
			return
		}
		response.InternalServerError(c, "token refresh failed")
		return
	}

	response.OK(c, "token refreshed successfully", gin.H{
		"tokens": tokens,
	})
}

// Logout handles user logout.
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	// Get access token from header
	accessToken := ""
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if len(authHeader) > 7 {
			accessToken = authHeader[7:] // Remove "Bearer " prefix
		}
	}

	// Get refresh token from body (optional)
	_ = c.ShouldBindJSON(&req)

	if err := h.authService.Logout(accessToken, req.RefreshToken); err != nil {
		response.InternalServerError(c, "logout failed")
		return
	}

	response.OK(c, "logout successful", nil)
}
