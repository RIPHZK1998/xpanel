package handler

import (
	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// NodeHandler handles node-related HTTP requests.
type NodeHandler struct {
	nodeService *service.NodeService
}

// NewNodeHandler creates a new node handler.
func NewNodeHandler(nodeService *service.NodeService) *NodeHandler {
	return &NodeHandler{nodeService: nodeService}
}

// GetNodes retrieves all available VPN nodes.
// GET /api/v1/nodes
func (h *NodeHandler) GetNodes(c *gin.Context) {
	nodes, err := h.nodeService.GetOnlineNodes()
	if err != nil {
		response.InternalServerError(c, "failed to retrieve nodes")
		return
	}

	nodeResponses := make([]interface{}, len(nodes))
	for i, node := range nodes {
		nodeResponses[i] = node.ToResponse()
	}

	response.OK(c, "nodes retrieved successfully", gin.H{
		"nodes": nodeResponses,
	})
}
