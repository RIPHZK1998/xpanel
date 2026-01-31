package handler

import (
	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminStatsHandler handles admin statistics requests.
type AdminStatsHandler struct {
	adminService *service.AdminService
}

// NewAdminStatsHandler creates a new admin stats handler.
func NewAdminStatsHandler(adminService *service.AdminService) *AdminStatsHandler {
	return &AdminStatsHandler{
		adminService: adminService,
	}
}

// GetOverview retrieves system overview statistics.
// GET /api/v1/admin/stats/overview
func (h *AdminStatsHandler) GetOverview(c *gin.Context) {
	stats, err := h.adminService.GetSystemStats()
	if err != nil {
		response.InternalServerError(c, "failed to retrieve statistics")
		return
	}

	response.OK(c, "statistics retrieved successfully", stats)
}
