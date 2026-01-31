package handler

import (
	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

type AdminSystemHandler struct {
	configService *service.SystemConfigService
}

func NewAdminSystemHandler(configService *service.SystemConfigService) *AdminSystemHandler {
	return &AdminSystemHandler{
		configService: configService,
	}
}

type UpdateSystemConfigRequest struct {
	Value string `json:"value" binding:"required"`
}

// GetConfig returns all system configuration
// GET /api/v1/admin/system/config
func (h *AdminSystemHandler) GetConfig(c *gin.Context) {
	configs, err := h.configService.GetAllConfigs(true) // Mask encrypted values
	if err != nil {
		response.InternalServerError(c, "failed to load config")
		return
	}

	response.OK(c, "system config retrieved", gin.H{
		"configs": configs,
	})
}

// UpdateConfig updates a specific configuration value
// PUT /api/v1/admin/system/config/:key
func (h *AdminSystemHandler) UpdateConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "config key is required")
		return
	}

	var req UpdateSystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// Get current user ID for audit trail
	userID, exists := c.Get("user_id")
	var updatedBy *uint
	if exists {
		if uid, ok := userID.(uint); ok {
			updatedBy = &uid
		}
	}

	// Determine if this key should be encrypted
	encrypted := key == "node_api_key" // Add more sensitive keys as needed

	// Set the configuration
	if err := h.configService.SetConfig(key, req.Value, encrypted, "", updatedBy); err != nil {
		response.InternalServerError(c, "failed to update config")
		return
	}

	response.OK(c, "configuration updated successfully", nil)
}

// ReloadCache forces a cache reload
// POST /api/v1/admin/system/config/reload
func (h *AdminSystemHandler) ReloadCache(c *gin.Context) {
	h.configService.ReloadCache()
	response.OK(c, "configuration cache reloaded", nil)
}

// RevealConfig returns the actual value of an encrypted config key
// GET /api/v1/admin/system/config/:key/reveal
func (h *AdminSystemHandler) RevealConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "config key is required")
		return
	}

	value, err := h.configService.GetConfig(key)
	if err != nil {
		response.NotFound(c, "configuration not found")
		return
	}

	response.OK(c, "configuration value revealed", gin.H{
		"key":   key,
		"value": value,
	})
}
