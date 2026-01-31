package handler

import (
	"strconv"
	"time"

	"xpanel/internal/models"
	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminUserSubscriptionHandler handles user subscription management.
type AdminUserSubscriptionHandler struct {
	adminService *service.AdminService
	userService  *service.UserService
}

// NewAdminUserSubscriptionHandler creates a new handler.
func NewAdminUserSubscriptionHandler(adminService *service.AdminService, userService *service.UserService) *AdminUserSubscriptionHandler {
	return &AdminUserSubscriptionHandler{
		adminService: adminService,
		userService:  userService,
	}
}

// UpdateUserSubscriptionRequest represents subscription update data.
type UpdateUserSubscriptionRequest struct {
	Plan          string     `json:"plan" binding:"required,oneof=free monthly yearly"`
	DataLimitGB   int64      `json:"data_limit_gb"`
	ExpiresAt     *time.Time `json:"expires_at"`
	ResetDataUsed bool       `json:"reset_data_used"`
}

// UpdateUserSubscription updates a user's subscription.
// PUT /api/v1/admin/users/:id/subscription
func (h *AdminUserSubscriptionHandler) UpdateUserSubscription(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	var req UpdateUserSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// Get user with subscription
	user, err := h.userService.GetProfile(uint(userID))
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	if user.Subscription == nil {
		response.BadRequest(c, "user has no subscription")
		return
	}

	// Update plan logic needs to map string plan nameto PlanID
	// For now, this old API might need to be deprecated or mapped
	// Let's assume the request sends plan ID as string or ignore plan change here
	// and use the separate AssignPlan API.
	// But to keep old frontend working, we might need a mapping.
	// However, we are building NEW frontend.
	// The new frontend calls `AssignPlanToUser` which is in `PlanHandler`.
	// This handler seems to be legacy or for manual edits?
	// If the user uses the new UI, they use `PlanService`.
	// Let's simplify this to just updating expiration and data usage for now,
	// or fail if Plan string is provided but invalid.

	// Actually, let's just support resetting data used and updating expiration.
	// Plan changes should go through AssignPlan.

	if req.ResetDataUsed {
		user.Subscription.DataUsedBytes = 0
	}

	if req.ExpiresAt != nil {
		user.Subscription.ExpiresAt = req.ExpiresAt
	}

	// Ensure subscription is active if not expired
	if user.Subscription.ExpiresAt == nil || time.Now().Before(*user.Subscription.ExpiresAt) {
		user.Subscription.Status = models.SubscriptionActive
	}

	// Update in database
	if err := h.adminService.UpdateUserSubscription(user.Subscription); err != nil {
		response.InternalServerError(c, "failed to update subscription")
		return
	}

	response.OK(c, "subscription updated successfully", gin.H{
		"subscription": user.Subscription.ToResponse(),
	})
}
