package handler

import (
	"strconv"
	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminAdministratorHandler handles administrator management requests.
type AdminAdministratorHandler struct {
	adminService *service.AdminService
}

// NewAdminAdministratorHandler creates a new administrator handler.
func NewAdminAdministratorHandler(adminService *service.AdminService) *AdminAdministratorHandler {
	return &AdminAdministratorHandler{
		adminService: adminService,
	}
}

// ListAdministrators retrieves all admin users.
// GET /api/v1/admin/administrators
func (h *AdminAdministratorHandler) ListAdministrators(c *gin.Context) {
	admins, err := h.adminService.ListAdministrators()
	if err != nil {
		response.InternalServerError(c, "failed to retrieve administrators")
		return
	}

	response.OK(c, "administrators retrieved successfully", gin.H{
		"administrators": admins,
	})
}

// CreateAdministrator creates a new admin user.
// POST /api/v1/admin/administrators
func (h *AdminAdministratorHandler) CreateAdministrator(c *gin.Context) {
	var req service.CreateAdministratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	admin, err := h.adminService.CreateAdministrator(&req)
	if err != nil {
		response.InternalServerError(c, "failed to create administrator")
		return
	}

	response.Created(c, "administrator created successfully", admin.ToResponse())
}

// ChangePassword changes an admin's password.
// PUT /api/v1/admin/administrators/:id/password
func (h *AdminAdministratorHandler) ChangePassword(c *gin.Context) {
	idStr := c.Param("id")
	adminID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid administrator ID")
		return
	}

	var req service.ChangeAdminPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if err := h.adminService.ChangeAdminPassword(uint(adminID), &req); err != nil {
		response.Unauthorized(c, "invalid current password or update failed")
		return
	}

	response.OK(c, "password changed successfully", nil)
}

// DeleteAdministrator deletes an admin user.
// DELETE /api/v1/admin/administrators/:id
func (h *AdminAdministratorHandler) DeleteAdministrator(c *gin.Context) {
	idStr := c.Param("id")
	adminID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid administrator ID")
		return
	}

	if err := h.adminService.DeleteAdministrator(uint(adminID)); err != nil {
		response.InternalServerError(c, "failed to delete administrator")
		return
	}

	response.OK(c, "administrator deleted successfully", nil)
}
