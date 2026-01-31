package handler

import (
	"strconv"

	"xpanel/internal/models"
	"xpanel/internal/service"
	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminUserHandler handles admin user management requests.
type AdminUserHandler struct {
	adminService *service.AdminService
	userService  *service.UserService
}

// NewAdminUserHandler creates a new admin user handler.
func NewAdminUserHandler(adminService *service.AdminService, userService *service.UserService) *AdminUserHandler {
	return &AdminUserHandler{
		adminService: adminService,
		userService:  userService,
	}
}

// ListUsers retrieves all users with pagination.
// GET /api/v1/admin/users
func (h *AdminUserHandler) ListUsers(c *gin.Context) {
	var req service.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query parameters: "+err.Error())
		return
	}

	result, err := h.adminService.ListUsers(&req)
	if err != nil {
		response.InternalServerError(c, "failed to retrieve users")
		return
	}

	response.OK(c, "users retrieved successfully", result)
}

// GetUser retrieves a specific user by ID.
// GET /api/v1/admin/users/:id
func (h *AdminUserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	user, err := h.userService.GetProfile(uint(id))
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	response.OK(c, "user retrieved successfully", gin.H{
		"user":         user.ToResponse(),
		"subscription": user.Subscription,
	})
}

// UpdateUser updates a user's information.
// PUT /api/v1/admin/users/:id
func (h *AdminUserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	user, err := h.adminService.UpdateUser(uint(id), &req)
	if err != nil {
		response.InternalServerError(c, "failed to update user")
		return
	}

	response.OK(c, "user updated successfully", gin.H{
		"user": user.ToResponse(),
	})
}

// DeleteUser deletes a user.
// DELETE /api/v1/admin/users/:id
func (h *AdminUserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	if err := h.userService.SuspendUser(uint(id)); err != nil {
		response.InternalServerError(c, "failed to delete user")
		return
	}

	response.OK(c, "user deleted successfully", nil)
}

// SuspendUser suspends a user account.
// POST /api/v1/admin/users/:id/suspend
func (h *AdminUserHandler) SuspendUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	if err := h.userService.SuspendUser(uint(id)); err != nil {
		response.InternalServerError(c, "failed to suspend user")
		return
	}

	response.OK(c, "user suspended successfully", nil)
}

// ActivateUser activates a suspended user account.
// POST /api/v1/admin/users/:id/activate
func (h *AdminUserHandler) ActivateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	if err := h.userService.ActivateUser(uint(id)); err != nil {
		response.InternalServerError(c, "failed to activate user")
		return
	}

	response.OK(c, "user activated successfully", nil)
}

// CreateUser creates a new user.
// POST /api/v1/admin/users
func (h *AdminUserHandler) CreateUser(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if req.Role == "" {
		req.Role = models.UserRoleUser
	}
	if req.Status == "" {
		req.Status = models.UserStatusActive
	}

	user, err := h.adminService.CreateUser(&req)
	if err != nil {
		response.InternalServerError(c, "failed to create user: "+err.Error())
		return
	}

	response.OK(c, "user created successfully", gin.H{
		"user": user.ToResponse(),
	})
}

// GetUserLinks retrieves connection links for all nodes accessible to a user.
// GET /api/v1/admin/users/:id/links
func (h *AdminUserHandler) GetUserLinks(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	linksResp, err := h.adminService.GetUserConnectionLinks(uint(id))
	if err != nil {
		response.InternalServerError(c, "failed to retrieve user links: "+err.Error())
		return
	}

	response.OK(c, "user links retrieved successfully", linksResp)
}
