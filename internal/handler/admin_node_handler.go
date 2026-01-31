package handler

import (
	"strconv"

	"xpanel/internal/models"
	"xpanel/internal/service"
	"xpanel/internal/xray"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminNodeHandler handles admin node management requests.
type AdminNodeHandler struct {
	nodeService *service.NodeService
	userService *service.UserService
	xrayManager *xray.Manager
}

// NewAdminNodeHandler creates a new admin node handler.
func NewAdminNodeHandler(
	nodeService *service.NodeService,
	userService *service.UserService,
	xrayManager *xray.Manager,
) *AdminNodeHandler {
	return &AdminNodeHandler{
		nodeService: nodeService,
		userService: userService,
		xrayManager: xrayManager,
	}
}

// ListNodes retrieves all nodes (including offline).
// GET /api/v1/admin/nodes
func (h *AdminNodeHandler) ListNodes(c *gin.Context) {
	nodes, err := h.nodeService.GetAllNodes()
	if err != nil {
		response.InternalServerError(c, "failed to retrieve nodes")
		return
	}

	response.OK(c, "nodes retrieved successfully", gin.H{
		"nodes": nodes,
	})
}

// CreateNode creates a new VPN node.
// POST /api/v1/admin/nodes
func (h *AdminNodeHandler) CreateNode(c *gin.Context) {
	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// Validate Reality configuration if enabled
	if node.RealityEnabled {
		if node.RealityDest == "" {
			response.BadRequest(c, "Target Destination is required when Reality is enabled")
			return
		}
		if node.RealityServerNames == "" {
			response.BadRequest(c, "Server Names (SNI) is required when Reality is enabled")
			return
		}
		if node.RealityShortIds == "" {
			response.BadRequest(c, "Short IDs is required when Reality is enabled")
			return
		}
	}

	// Set default inbound_tag if not provided
	if node.InboundTag == "" {
		node.InboundTag = "proxy"
	}

	if err := h.nodeService.CreateNode(&node); err != nil {
		response.InternalServerError(c, "failed to create node")
		return
	}

	// Register with xray manager
	h.xrayManager.RegisterNode(&node)

	response.Created(c, "node created successfully", gin.H{
		"node": node,
	})
}

// GetNode retrieves a specific node by ID.
// GET /api/v1/admin/nodes/:id
func (h *AdminNodeHandler) GetNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid node ID")
		return
	}

	node, err := h.nodeService.GetNode(uint(id))
	if err != nil {
		response.NotFound(c, "node not found")
		return
	}

	response.OK(c, "node retrieved successfully", gin.H{
		"node": node,
	})
}

// UpdateNode updates an existing node.
// PUT /api/v1/admin/nodes/:id
func (h *AdminNodeHandler) UpdateNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid node ID")
		return
	}

	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// Validate Reality configuration if enabled
	if node.RealityEnabled {
		if node.RealityDest == "" {
			response.BadRequest(c, "Target Destination is required when Reality is enabled")
			return
		}
		if node.RealityServerNames == "" {
			response.BadRequest(c, "Server Names (SNI) is required when Reality is enabled")
			return
		}
		if node.RealityShortIds == "" {
			response.BadRequest(c, "Short IDs is required when Reality is enabled")
			return
		}
	}

	node.ID = uint(id)
	if err := h.nodeService.UpdateNode(&node); err != nil {
		response.InternalServerError(c, "failed to update node")
		return
	}

	// Re-register with xray manager
	h.xrayManager.RegisterNode(&node)

	response.OK(c, "node updated successfully", gin.H{
		"node": node,
	})
}

// DeleteNode deletes a node.
// DELETE /api/v1/admin/nodes/:id
func (h *AdminNodeHandler) DeleteNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid node ID")
		return
	}

	if err := h.nodeService.DeleteNode(uint(id)); err != nil {
		response.InternalServerError(c, "failed to delete node")
		return
	}

	// Unregister from xray manager
	h.xrayManager.UnregisterNode(uint(id))

	response.OK(c, "node deleted successfully", nil)
}

// SyncNode syncs all active users to a specific node.
// POST /api/v1/admin/nodes/:id/sync
func (h *AdminNodeHandler) SyncNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid node ID")
		return
	}

	node, err := h.nodeService.GetNode(uint(id))
	if err != nil {
		response.NotFound(c, "node not found")
		return
	}

	// Get all active users
	users, err := h.userService.GetActiveUsers()
	if err != nil {
		response.InternalServerError(c, "failed to retrieve users")
		return
	}

	// Provision users to node
	successCount := 0
	for _, user := range users {
		if err := h.xrayManager.ProvisionUser(node.ID, &user, node); err == nil {
			successCount++
		}
	}

	response.OK(c, "node sync completed", gin.H{
		"total_users":  len(users),
		"synced_users": successCount,
		"failed_users": len(users) - successCount,
	})
}

// GetNodeStats retrieves statistics for a specific node.
// GET /api/v1/admin/nodes/:id/stats
func (h *AdminNodeHandler) GetNodeStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid node ID")
		return
	}

	node, err := h.nodeService.GetNode(uint(id))
	if err != nil {
		response.NotFound(c, "node not found")
		return
	}

	response.OK(c, "node stats retrieved successfully", gin.H{
		"node": node,
	})
}
