package handler

import (
	"xpanel/internal/middleware"
	"xpanel/internal/service"
	"xpanel/internal/xray"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// ConfigHandler handles VPN configuration requests.
type ConfigHandler struct {
	userService *service.UserService
	nodeService *service.NodeService
	xrayManager *xray.Manager
}

// NewConfigHandler creates a new config handler.
func NewConfigHandler(
	userService *service.UserService,
	nodeService *service.NodeService,
	xrayManager *xray.Manager,
) *ConfigHandler {
	return &ConfigHandler{
		userService: userService,
		nodeService: nodeService,
		xrayManager: xrayManager,
	}
}

// GetConfig generates VPN configuration for the user.
// GET /api/v1/user/config
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}

	// Get user
	user, err := h.userService.GetProfile(userID)
	if err != nil {
		response.InternalServerError(c, "failed to retrieve user")
		return
	}

	// Get available nodes
	nodes, err := h.nodeService.GetAvailableNodes()
	if err != nil {
		response.InternalServerError(c, "failed to retrieve nodes")
		return
	}

	if len(nodes) == 0 {
		response.NotFound(c, "no available nodes")
		return
	}

	// Generate configs for all available nodes
	configs := make([]interface{}, 0, len(nodes))
	for _, node := range nodes {
		config, err := h.xrayManager.GenerateClientConfig(user, &node)
		if err != nil {
			continue
		}
		configs = append(configs, config)
	}

	if len(configs) == 0 {
		response.InternalServerError(c, "failed to generate configurations")
		return
	}

	response.OK(c, "configurations generated successfully", gin.H{
		"configs": configs,
	})
}
