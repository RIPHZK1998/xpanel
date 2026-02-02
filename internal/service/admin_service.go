package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
	"xpanel/internal/models"
	"xpanel/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// AdminService handles admin-related business logic.
type AdminService struct {
	userRepo     *repository.UserRepository
	subRepo      *repository.SubscriptionRepository
	userSubRepo  *repository.UserSubscriptionRepository
	nodeRepo     *repository.NodeRepository
	trafficRepo  *repository.TrafficRepository
	activityRepo *repository.ActivityRepository
}

// NewAdminService creates a new admin service.
func NewAdminService(
	userRepo *repository.UserRepository,
	subRepo *repository.SubscriptionRepository,
	userSubRepo *repository.UserSubscriptionRepository,
	nodeRepo *repository.NodeRepository,
	trafficRepo *repository.TrafficRepository,
	activityRepo *repository.ActivityRepository,
) *AdminService {
	return &AdminService{
		userRepo:     userRepo,
		subRepo:      subRepo,
		userSubRepo:  userSubRepo,
		nodeRepo:     nodeRepo,
		trafficRepo:  trafficRepo,
		activityRepo: activityRepo,
	}
}

// ListUsersRequest contains pagination and filter parameters.
type ListUsersRequest struct {
	Page     int               `form:"page" binding:"min=1"`
	PageSize int               `form:"page_size" binding:"min=1,max=100"`
	Status   models.UserStatus `form:"status"`
	Role     models.UserRole   `form:"role"`
	Search   string            `form:"search"` // Email search
}

// ListUsersResponse contains paginated user list.
type ListUsersResponse struct {
	Users      []models.UserResponse `json:"users"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// ListUsers retrieves a paginated list of users with filters.
func (s *AdminService) ListUsers(req *ListUsersRequest) (*ListUsersResponse, error) {
	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// This is a simplified implementation
	// In production, add proper pagination to repository layer
	users, err := s.userRepo.GetAllUsersWithSubscription()
	if err != nil {
		return nil, err
	}

	// Collect user IDs for activity lookup
	userIDs := make([]uint, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	// Fetch activity data for all users
	activities, _ := s.activityRepo.GetActivitiesForUsers(userIDs)

	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		resp := user.ToResponse()

		// Add activity data if available
		if activity, ok := activities[user.ID]; ok {
			// User is considered online if seen within last 2 minutes
			if time.Since(activity.LastSeen) < 2*time.Minute {
				resp.OnlineDevices = 1 // At least 1 device online
			}
			resp.LastSeen = &activity.LastSeen
		}

		userResponses[i] = resp
	}

	return &ListUsersResponse{
		Users:      userResponses,
		Total:      int64(len(users)),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (len(users) + req.PageSize - 1) / req.PageSize,
	}, nil
}

// CreateUserRequest contains data for creating a new user.
type CreateUserRequest struct {
	Email    string            `json:"email"`
	Password string            `json:"password"`
	Role     models.UserRole   `json:"role"`
	Status   models.UserStatus `json:"status"`
}

// CreateUser creates a new user with subscription.
// For admin users, no subscription is created.
func (s *AdminService) CreateUser(req *CreateUserRequest) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		Status:       req.Status,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Only create subscription for regular users, not admins
	if req.Role != models.UserRoleAdmin {
		// Create free subscription by default
		// Assuming ID 1 is the default free plan.
		subscription := &models.UserSubscription{
			UserID:    user.ID,
			PlanID:    1,
			Status:    models.SubscriptionActive,
			StartDate: time.Now(),
		}

		if err := s.userSubRepo.Create(subscription); err != nil {
			_ = s.userRepo.Delete(user.ID)
			return nil, err
		}

		// Load full subscription details (with Plan) for response
		fullSub, err := s.userSubRepo.GetByUserID(user.ID)
		if err == nil {
			user.Subscription = fullSub
		} else {
			// Fallback if fetch fails
			user.Subscription = subscription
		}
	}

	return user, nil
}

// UpdateUserRequest contains user update data.
type UpdateUserRequest struct {
	Status *models.UserStatus `json:"status"`
	Role   *models.UserRole   `json:"role"`
}

// UpdateUser updates a user's information.
func (s *AdminService) UpdateUser(userID uint, req *UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.Role != nil {
		user.Role = *req.Role
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

// SystemStats represents overall system statistics.
type SystemStats struct {
	TotalUsers          int64 `json:"total_users"`
	ActiveUsers         int64 `json:"active_users"`
	TotalNodes          int64 `json:"total_nodes"`
	OnlineNodes         int64 `json:"online_nodes"`
	TotalTrafficBytes   int64 `json:"total_traffic_bytes"`
	ActiveSubscriptions int64 `json:"active_subscriptions"`
}

// GetSystemStats retrieves overall system statistics.
func (s *AdminService) GetSystemStats() (*SystemStats, error) {
	stats := &SystemStats{}

	// Get user counts
	allUsers, err := s.userRepo.GetActiveUsers()
	if err != nil {
		return nil, err
	}
	stats.TotalUsers = int64(len(allUsers))
	stats.ActiveUsers = int64(len(allUsers)) // Simplified

	// Get node counts
	allNodes, err := s.nodeRepo.GetAll()
	if err != nil {
		return nil, err
	}
	stats.TotalNodes = int64(len(allNodes))

	onlineNodes, err := s.nodeRepo.GetOnlineNodes()
	if err != nil {
		return nil, err
	}
	stats.OnlineNodes = int64(len(onlineNodes))

	// Get subscription counts
	activeSubs, err := s.subRepo.GetActiveSubscriptions()
	if err != nil {
		return nil, err
	}
	stats.ActiveSubscriptions = int64(len(activeSubs))

	return stats, nil
}

// ResetDataUsageRequest contains data reset parameters.
type ResetDataUsageRequest struct {
	ResetToZero bool `json:"reset_to_zero"`
}

// UpdateUserSubscription updates a user's subscription.
func (s *AdminService) UpdateUserSubscription(subscription *models.UserSubscription) error {
	return s.userSubRepo.Update(subscription)
}

// Administrator Management

// ListAdministrators retrieves all admin users.
func (s *AdminService) ListAdministrators() ([]models.UserResponse, error) {
	users, err := s.userRepo.GetAllUsersWithSubscription()
	if err != nil {
		return nil, err
	}

	// Filter only admin users
	var admins []models.UserResponse
	for _, user := range users {
		if user.Role == models.UserRoleAdmin {
			admins = append(admins, user.ToResponse())
		}
	}

	return admins, nil
}

// CreateAdministratorRequest contains data for creating a new administrator.
type CreateAdministratorRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// CreateAdministrator creates a new admin user without a subscription.
func (s *AdminService) CreateAdministrator(req *CreateAdministratorRequest) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	admin := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         models.UserRoleAdmin,
		Status:       models.UserStatusActive,
	}

	if err := s.userRepo.Create(admin); err != nil {
		return nil, err
	}

	return admin, nil
}

// ChangeAdminPasswordRequest contains password change data.
type ChangeAdminPasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ChangeAdminPassword updates an admin's password after verifying current password.
func (s *AdminService) ChangeAdminPassword(adminID uint, req *ChangeAdminPasswordRequest) error {
	admin, err := s.userRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin.PasswordHash = string(hashedPassword)
	return s.userRepo.Update(admin)
}

// DeleteAdministrator soft-deletes an admin user.
func (s *AdminService) DeleteAdministrator(adminID uint) error {
	return s.userRepo.Delete(adminID)
}

// UserLinksResponse contains all connection links for a user.
type UserLinksResponse struct {
	UserUUID        string         `json:"user_uuid"`
	Links           []NodeLinkInfo `json:"links"`
	SubscriptionURL string         `json:"subscription_url,omitempty"`
}

// NodeLinkInfo represents connection information for a single node.
type NodeLinkInfo struct {
	NodeID   uint   `json:"node_id"`
	NodeName string `json:"node_name"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Link     string `json:"link"`
	QRData   string `json:"qr_data"` // Same as link, for QR code generation
}

// GetUserConnectionLinks generates connection links for all nodes accessible to a user.
func (s *AdminService) GetUserConnectionLinks(userID uint) (*UserLinksResponse, error) {
	// Get user with subscription
	user, err := s.userRepo.GetByIDWithSubscription(userID)
	if err != nil {
		return nil, err
	}

	response := &UserLinksResponse{
		UserUUID: user.UUID,
		Links:    []NodeLinkInfo{},
	}

	// If user has no subscription, return empty links
	if user.Subscription == nil {
		return response, nil
	}

	// Get subscription with plan details
	subscription, err := s.userSubRepo.GetByUserID(user.ID)
	if err != nil {
		return response, nil
	}

	// If subscription has no plan, return empty links
	if subscription.Plan == nil {
		return response, nil
	}

	// Get all nodes associated with the plan
	plan := subscription.Plan
	if len(plan.Nodes) == 0 {
		return response, nil
	}

	// Generate links for each node
	for _, node := range plan.Nodes {
		// Import the proxy package functionality inline to avoid circular dependency
		// We'll generate the link directly here

		// For Reality, use RealityServerNames as SNI; otherwise use regular SNI
		sni := node.SNI
		if node.RealityEnabled && node.RealityServerNames != "" {
			// Use first server name from Reality settings
			sni = extractFirstServerName(node.RealityServerNames)
		}

		link := generateNodeLink(
			user.UUID,
			node.Address,
			node.Port,
			string(node.Protocol),
			node.Name,
			sni,
			node.RealityEnabled,
			node.RealityPublicKey,
			extractFirstShortID(node.RealityShortIds),
		)

		response.Links = append(response.Links, NodeLinkInfo{
			NodeID:   node.ID,
			NodeName: node.Name,
			Protocol: string(node.Protocol),
			Address:  node.Address,
			Port:     node.Port,
			Link:     link,
			QRData:   link,
		})
	}

	return response, nil
}

// Helper function to generate node link (embedded version of proxy.GenerateNodeLink)
func generateNodeLink(uuid, address string, port int, protocol, nodeName, sni string, realityEnabled bool, realityPublicKey, realityShortID string) string {
	// Import the package to use its functions
	// This is a temporary wrapper - in production you'd import xpanel/pkg/proxy
	// For now, we'll use a simplified version
	remark := nodeName

	switch protocol {
	case "vless":
		return generateVLESSLink(uuid, address, port, sni, remark, realityEnabled, realityPublicKey, realityShortID)
	case "trojan":
		return generateTrojanLink(uuid, address, port, sni, remark)
	case "vmess":
		return generateVMessLink(uuid, address, port, sni, remark)
	default:
		return generateVLESSLink(uuid, address, port, sni, remark, realityEnabled, realityPublicKey, realityShortID)
	}
}

// Simplified link generators (these would normally come from pkg/proxy)
func generateVLESSLink(uuid, address string, port int, sni, remark string, realityEnabled bool, realityPublicKey, realityShortID string) string {
	params := url.Values{}
	if realityEnabled {
		params.Set("type", "tcp")
		params.Set("encryption", "none")
		params.Set("security", "reality")
		params.Set("flow", "xtls-rprx-vision")
		params.Set("fp", "chrome")
		if sni != "" {
			params.Set("sni", sni)
		}
		if realityPublicKey != "" {
			params.Set("pbk", realityPublicKey)
		}
		if realityShortID != "" {
			params.Set("sid", realityShortID)
		}
	} else {
		params.Set("type", "tcp")
		params.Set("encryption", "none")
		params.Set("security", "tls")
		params.Set("fp", "chrome")
		if sni != "" {
			params.Set("sni", sni)
		}
	}

	link := fmt.Sprintf("vless://%s@%s:%d?%s", uuid, address, port, params.Encode())
	if remark != "" {
		link += "#" + url.QueryEscape(remark)
	}
	return link
}

func generateTrojanLink(uuid, address string, port int, sni, remark string) string {
	params := url.Values{}
	params.Set("security", "tls")
	params.Set("type", "tcp")
	params.Set("headerType", "none")
	params.Set("sni", sni)

	link := fmt.Sprintf("trojan://%s@%s:%d?%s", uuid, address, port, params.Encode())
	if remark != "" {
		link += "#" + url.QueryEscape(remark)
	}
	return link
}

func generateVMessLink(uuid, address string, port int, sni, remark string) string {
	config := map[string]interface{}{
		"v":    "2",
		"ps":   remark,
		"add":  address,
		"port": port,
		"id":   uuid,
		"aid":  0,
		"net":  "tcp",
		"type": "none",
		"host": sni,
		"path": "",
		"tls":  "tls",
		"sni":  sni,
	}

	jsonData, _ := json.Marshal(config)
	encoded := base64.StdEncoding.EncodeToString(jsonData)
	return "vmess://" + encoded
}

func extractFirstShortID(shortIds string) string {
	if shortIds == "" {
		return ""
	}
	parts := strings.Split(shortIds, ",")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return ""
}

func extractFirstServerName(serverNames string) string {
	if serverNames == "" {
		return ""
	}
	parts := strings.Split(serverNames, ",")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return ""
}
