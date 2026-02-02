package handler

import (
	"xpanel/internal/middleware"
	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// SubscriptionHandler handles subscription-related HTTP requests.
type SubscriptionHandler struct {
	subscriptionService *service.SubscriptionService
	trafficService      *service.TrafficService
}

// NewSubscriptionHandler creates a new subscription handler.
func NewSubscriptionHandler(
	subscriptionService *service.SubscriptionService,
	trafficService *service.TrafficService,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
		trafficService:      trafficService,
	}
}

// GetSubscription retrieves the user's subscription details.
// GET /api/v1/user/subscription
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}

	sub, err := h.subscriptionService.GetSubscription(userID)
	if err != nil {
		response.InternalServerError(c, "failed to retrieve subscription")
		return
	}

	// Get traffic stats
	stats, err := h.trafficService.GetUserStats(userID)
	if err != nil {
		// Continue even if stats fail
		stats = nil
	}

	response.OK(c, "subscription retrieved successfully", gin.H{
		"subscription": sub.ToResponse(),
		"traffic":      stats,
	})
}

// Renew renews or upgrades the user's subscription.
// POST /api/v1/subscription/renew
func (h *SubscriptionHandler) Renew(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}

	var req service.RenewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	sub, err := h.subscriptionService.Renew(userID, req.PlanID, req.AutoRenew)
	if err != nil {
		if err == service.ErrInvalidPlan {
			response.BadRequest(c, "invalid subscription plan")
			return
		}
		response.InternalServerError(c, "failed to renew subscription")
		return
	}

	response.OK(c, "subscription renewed successfully", gin.H{
		"subscription": sub.ToResponse(),
	})
}
