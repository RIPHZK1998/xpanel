package handler

import (
	"strconv"

	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminSubscriptionHandler handles admin subscription management requests.
type AdminSubscriptionHandler struct {
	subscriptionService *service.SubscriptionService
}

// NewAdminSubscriptionHandler creates a new admin subscription handler.
func NewAdminSubscriptionHandler(subscriptionService *service.SubscriptionService) *AdminSubscriptionHandler {
	return &AdminSubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// ListSubscriptions retrieves all subscriptions.
// GET /api/v1/admin/subscriptions
func (h *AdminSubscriptionHandler) ListSubscriptions(c *gin.Context) {
	subs, err := h.subscriptionService.GetActiveSubscriptions()
	if err != nil {
		response.InternalServerError(c, "failed to retrieve subscriptions")
		return
	}

	subResponses := make([]interface{}, len(subs))
	for i, sub := range subs {
		subResponses[i] = sub.ToResponse()
	}

	response.OK(c, "subscriptions retrieved successfully", gin.H{
		"subscriptions": subResponses,
		"total":         len(subs),
	})
}

// ExtendSubscription extends a subscription's expiration date.
// POST /api/v1/admin/subscriptions/:id/extend
func (h *AdminSubscriptionHandler) ExtendSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid subscription ID")
		return
	}

	var req service.ExtendSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// Get subscription
	sub, err := h.subscriptionService.GetSubscription(uint(id))
	if err != nil {
		response.NotFound(c, "subscription not found")
		return
	}

	// Extend expiration via service
	sub, err := h.subscriptionService.Extend(sub.UserID, req.Days)
	if err != nil {
		response.InternalServerError(c, "failed to extend subscription")
		return
	}

	response.OK(c, "subscription extended successfully", gin.H{
		"subscription": sub.ToResponse(),
	})
}

// ResetDataUsage resets a subscription's data usage.
// POST /api/v1/admin/subscriptions/:id/reset-data
func (h *AdminSubscriptionHandler) ResetDataUsage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid subscription ID")
		return
	}

	// Get subscription
	sub, err := h.subscriptionService.GetSubscription(uint(id))
	if err != nil {
		response.NotFound(c, "subscription not found")
		return
	}

	// Reset data usage
	if err := h.subscriptionService.UpdateDataUsage(sub.UserID, -sub.DataUsedBytes); err != nil {
		response.InternalServerError(c, "failed to reset data usage")
		return
	}

	response.OK(c, "data usage reset successfully", nil)
}
