package handler

import (
	"xpanel/internal/models"
	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// NodeAgentHandler handles requests from node agents.
type NodeAgentHandler struct {
	nodeAgentService *service.NodeAgentService
}

// NewNodeAgentHandler creates a new node agent handler.
func NewNodeAgentHandler(nodeAgentService *service.NodeAgentService) *NodeAgentHandler {
	return &NodeAgentHandler{
		nodeAgentService: nodeAgentService,
	}
}

// Heartbeat receives heartbeat from a node agent.
// POST /api/v1/node-agent/heartbeat
func (h *NodeAgentHandler) Heartbeat(c *gin.Context) {
	var heartbeat models.NodeHeartbeat
	if err := c.ShouldBindJSON(&heartbeat); err != nil {
		response.BadRequest(c, "invalid heartbeat data: "+err.Error())
		return
	}

	if err := h.nodeAgentService.ProcessHeartbeat(&heartbeat); err != nil {
		response.InternalServerError(c, "failed to process heartbeat")
		return
	}

	response.OK(c, "heartbeat received", nil)
}

// SyncUsers returns the list of users that should be provisioned on the node.
// GET /api/v1/node-agent/:node_id/sync
func (h *NodeAgentHandler) SyncUsers(c *gin.Context) {
	var params struct {
		NodeID uint `uri:"node_id" binding:"required"`
	}

	if err := c.ShouldBindUri(&params); err != nil {
		response.BadRequest(c, "invalid node ID")
		return
	}

	syncData, err := h.nodeAgentService.GetUserSyncData(params.NodeID)
	if err != nil {
		response.NotFound(c, "node not found")
		return
	}

	response.OK(c, "user sync data retrieved", syncData)
}

// ReportTraffic receives traffic statistics from a node agent.
// POST /api/v1/node-agent/traffic
func (h *NodeAgentHandler) ReportTraffic(c *gin.Context) {
	var report models.NodeTrafficReport
	if err := c.ShouldBindJSON(&report); err != nil {
		response.BadRequest(c, "invalid traffic report: "+err.Error())
		return
	}

	if err := h.nodeAgentService.ProcessTrafficReport(&report); err != nil {
		response.InternalServerError(c, "failed to process traffic report")
		return
	}

	response.OK(c, "traffic report received", nil)
}
