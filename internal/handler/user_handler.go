package handler

import (
	"xpanel/internal/middleware"
	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile retrieves the authenticated user's profile.
// GET /api/v1/user/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		response.InternalServerError(c, "failed to retrieve profile")
		return
	}

	response.OK(c, "profile retrieved successfully", gin.H{
		"user":         user.ToResponse(),
		"subscription": user.Subscription.ToResponse(),
	})
}

// GetDevices retrieves the user's connected devices.
// GET /api/v1/user/devices
func (h *UserHandler) GetDevices(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}

	devices, err := h.userService.GetDevices(userID)
	if err != nil {
		response.InternalServerError(c, "failed to retrieve devices")
		return
	}

	deviceResponses := make([]interface{}, len(devices))
	for i, device := range devices {
		deviceResponses[i] = device.ToResponse()
	}

	response.OK(c, "devices retrieved successfully", gin.H{
		"devices": deviceResponses,
	})
}

// DeactivateDevice deactivates a specific device.
// DELETE /api/v1/user/devices/:id
func (h *UserHandler) DeactivateDevice(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}

	var deviceID uint
	if err := c.ShouldBindUri(&struct {
		ID uint `uri:"id" binding:"required"`
	}{ID: deviceID}); err != nil {
		response.BadRequest(c, "invalid device ID")
		return
	}

	if err := h.userService.DeactivateDevice(userID, deviceID); err != nil {
		response.InternalServerError(c, "failed to deactivate device")
		return
	}

	response.OK(c, "device deactivated successfully", nil)
}
