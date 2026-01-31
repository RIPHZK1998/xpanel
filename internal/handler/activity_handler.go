package handler

import (
	"net/http"

	"xpanel/internal/models"
	"xpanel/internal/service"

	"github.com/gin-gonic/gin"
)

// ActivityHandler handles activity-related API endpoints.
type ActivityHandler struct {
	activityService *service.ActivityService
}

// NewActivityHandler creates a new activity handler.
func NewActivityHandler(activityService *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		activityService: activityService,
	}
}

// ReportActivity handles POST /api/v1/node-agent/activity
// Receives activity reports from node agents.
func (h *ActivityHandler) ReportActivity(c *gin.Context) {
	var report models.NodeActivityReport
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	// Process the activity report
	if err := h.activityService.ProcessActivityReport(report.NodeID, report.Users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to process activity report",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Activity report processed",
	})
}

// GetOnlineUsers handles GET /api/v1/admin/activity/online
// Returns the list of currently online users.
func (h *ActivityHandler) GetOnlineUsers(c *gin.Context) {
	activities, err := h.activityService.GetOnlineUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get online users",
		})
		return
	}

	// Convert to response format
	users := make([]models.UserActivityResponse, len(activities))
	for i, a := range activities {
		users[i] = a.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"users": users,
			"count": len(users),
		},
	})
}

// GetStats handles GET /api/v1/admin/activity/stats
// Returns activity statistics (online count, total count).
func (h *ActivityHandler) GetStats(c *gin.Context) {
	online, total, err := h.activityService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get activity stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"online": online,
			"total":  total,
		},
	})
}
