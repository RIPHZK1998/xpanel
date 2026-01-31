package handler

import (
	"fmt"
	"strconv"

	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// NodeConfigHandler handles node configuration requests from agents.
type NodeConfigHandler struct {
	nodeService *service.NodeService
}

// NewNodeConfigHandler creates a new node config handler.
func NewNodeConfigHandler(nodeService *service.NodeService) *NodeConfigHandler {
	return &NodeConfigHandler{
		nodeService: nodeService,
	}
}

// GetNodeConfig returns the configuration for a specific node.
// GET /api/v1/node-agent/:node_id/config
func (h *NodeConfigHandler) GetNodeConfig(c *gin.Context) {
	nodeID, err := strconv.ParseUint(c.Param("node_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid node ID")
		return
	}

	node, err := h.nodeService.GetNode(uint(nodeID))
	if err != nil {
		response.NotFound(c, "node not found")
		return
	}

	// Return node configuration
	config := gin.H{
		"id":           node.ID,
		"name":         node.Name,
		"address":      node.Address,
		"port":         node.Port,
		"protocol":     node.Protocol,
		"tls_enabled":  node.TLSEnabled,
		"sni":          node.SNI,
		"inbound_tag":  node.InboundTag,
		"api_endpoint": node.APIEndpoint,
		"api_port":     node.APIPort,

		// Reality settings
		"reality_enabled":      node.RealityEnabled,
		"reality_dest":         node.RealityDest,
		"reality_server_names": node.RealityServerNames,
		"reality_private_key":  node.RealityPrivateKey,
		"reality_public_key":   node.RealityPublicKey,
		"reality_short_ids":    node.RealityShortIds,
	}

	// Add logging
	if h.nodeService != nil {
		fmt.Printf("[Config] Agent fetched config for node %d (%s): protocol=%s, port=%d, reality=%v\n",
			node.ID, node.Name, node.Protocol, node.Port, node.RealityEnabled)
	}

	response.OK(c, "node configuration retrieved", gin.H{
		"node": config,
	})
}
