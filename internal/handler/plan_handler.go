package handler

import (
	"strconv"

	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// PlanHandler handles subscription plan-related HTTP requests.
type PlanHandler struct {
	planService *service.PlanService
}

// NewPlanHandler creates a new plan handler instance.
func NewPlanHandler(planService *service.PlanService) *PlanHandler {
	return &PlanHandler{
		planService: planService,
	}
}

// ListPlans retrieves all subscription plans.
// GET /api/v1/admin/plans
func (h *PlanHandler) ListPlans(c *gin.Context) {
	plans, err := h.planService.GetAllPlans()
	if err != nil {
		response.InternalServerError(c, "failed to retrieve plans")
		return
	}

	// Convert to response format
	planResponses := make([]interface{}, len(plans))
	for i, plan := range plans {
		planResponses[i] = plan.ToResponse()
	}

	response.OK(c, "plans retrieved successfully", gin.H{
		"plans": planResponses,
		"total": len(planResponses),
	})
}

// GetPlan retrieves a specific subscription plan.
// GET /api/v1/admin/plans/:id
func (h *PlanHandler) GetPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid plan ID")
		return
	}

	plan, err := h.planService.GetPlan(uint(id))
	if err != nil {
		response.NotFound(c, "plan not found")
		return
	}

	response.OK(c, "plan retrieved successfully", gin.H{
		"plan": plan.ToResponse(),
	})
}

// CreatePlan creates a new subscription plan.
// POST /api/v1/admin/plans
func (h *PlanHandler) CreatePlan(c *gin.Context) {
	var req service.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	plan, err := h.planService.CreatePlan(&req)
	if err != nil {
		response.InternalServerError(c, "failed to create plan: "+err.Error())
		return
	}

	response.Created(c, "plan created successfully", gin.H{
		"plan": plan.ToResponse(),
	})
}

// UpdatePlan updates an existing subscription plan.
// PUT /api/v1/admin/plans/:id
func (h *PlanHandler) UpdatePlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid plan ID")
		return
	}

	var req service.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	plan, err := h.planService.UpdatePlan(uint(id), &req)
	if err != nil {
		response.InternalServerError(c, "failed to update plan: "+err.Error())
		return
	}

	response.OK(c, "plan updated successfully", gin.H{
		"plan": plan.ToResponse(),
	})
}

// DeletePlan deletes a subscription plan.
// DELETE /api/v1/admin/plans/:id
func (h *PlanHandler) DeletePlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid plan ID")
		return
	}

	// Try to archive first (safer than delete)
	if err := h.planService.ArchivePlan(uint(id)); err != nil {
		response.InternalServerError(c, "failed to archive plan: "+err.Error())
		return
	}

	response.OK(c, "plan archived successfully", nil)
}

// AssignNodes assigns nodes to a subscription plan.
// PUT /api/v1/admin/plans/:id/nodes
func (h *PlanHandler) AssignNodes(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid plan ID")
		return
	}

	var req struct {
		NodeIDs []uint `json:"node_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if err := h.planService.AssignNodesToPlan(uint(id), req.NodeIDs); err != nil {
		response.InternalServerError(c, "failed to assign nodes: "+err.Error())
		return
	}

	// Get updated plan
	plan, err := h.planService.GetPlan(uint(id))
	if err != nil {
		response.InternalServerError(c, "failed to retrieve updated plan")
		return
	}

	response.OK(c, "nodes assigned successfully", gin.H{
		"plan": plan.ToResponse(),
	})
}

// GetPlanNodes retrieves all nodes assigned to a plan.
// GET /api/v1/admin/plans/:id/nodes
func (h *PlanHandler) GetPlanNodes(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid plan ID")
		return
	}

	nodes, err := h.planService.GetPlanNodes(uint(id))
	if err != nil {
		response.InternalServerError(c, "failed to retrieve nodes")
		return
	}

	// Convert to response format
	nodeResponses := make([]interface{}, len(nodes))
	for i, node := range nodes {
		nodeResponses[i] = node.ToResponse()
	}

	response.OK(c, "nodes retrieved successfully", gin.H{
		"nodes": nodeResponses,
		"total": len(nodeResponses),
	})
}

// GetPlanUsers retrieves all users subscribed to a plan.
// GET /api/v1/admin/plans/:id/users
func (h *PlanHandler) GetPlanUsers(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid plan ID")
		return
	}

	subscriptions, err := h.planService.GetPlanUsers(uint(id))
	if err != nil {
		response.InternalServerError(c, "failed to retrieve users")
		return
	}

	// Convert to response format
	subResponses := make([]interface{}, len(subscriptions))
	for i, sub := range subscriptions {
		subResponses[i] = sub.ToResponse()
	}

	response.OK(c, "users retrieved successfully", gin.H{
		"subscriptions": subResponses,
		"total":         len(subResponses),
	})
}

// AssignPlanToUser assigns a subscription plan to a user.
// PUT /api/v1/admin/users/:id/plan
func (h *PlanHandler) AssignPlanToUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	var req struct {
		PlanID    uint `json:"plan_id" binding:"required"`
		AutoRenew bool `json:"auto_renew"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	assignReq := &service.AssignPlanToUserRequest{
		UserID:    uint(userID),
		PlanID:    req.PlanID,
		AutoRenew: req.AutoRenew,
	}

	subscription, err := h.planService.AssignPlanToUser(assignReq)
	if err != nil {
		response.InternalServerError(c, "failed to assign plan: "+err.Error())
		return
	}

	response.OK(c, "plan assigned successfully", gin.H{
		"subscription": subscription.ToResponse(),
	})
}
