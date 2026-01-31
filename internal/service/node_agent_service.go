package service

import (
	"time"

	"xpanel/internal/models"
	"xpanel/internal/repository"
)

// NodeAgentService handles node agent communication.
type NodeAgentService struct {
	nodeRepo    *repository.NodeRepository
	userRepo    *repository.UserRepository
	userSubRepo *repository.UserSubscriptionRepository
	trafficRepo *repository.TrafficRepository
}

// NewNodeAgentService creates a new node agent service.
func NewNodeAgentService(
	nodeRepo *repository.NodeRepository,
	userRepo *repository.UserRepository,
	userSubRepo *repository.UserSubscriptionRepository,
	trafficRepo *repository.TrafficRepository,
) *NodeAgentService {
	return &NodeAgentService{
		nodeRepo:    nodeRepo,
		userRepo:    userRepo,
		userSubRepo: userSubRepo,
		trafficRepo: trafficRepo,
	}
}

// ProcessHeartbeat processes a heartbeat from a node agent.
func (s *NodeAgentService) ProcessHeartbeat(heartbeat *models.NodeHeartbeat) error {
	// Verify node exists
	_, err := s.nodeRepo.GetByID(heartbeat.NodeID)
	if err != nil {
		return err
	}

	// Map heartbeat status to node status
	var status models.NodeStatus
	switch heartbeat.Status {
	case "online":
		status = models.NodeStatusOnline
	case "offline":
		status = models.NodeStatusOffline
	case "maintenance":
		status = models.NodeStatusMaintenance
	default:
		status = models.NodeStatusOffline
	}

	now := time.Now()

	// Optimized: Only update the fields that changed (status, current_users, last_check_at, reality_public_key)
	// This is much faster than updating all 30+ fields
	return s.nodeRepo.UpdateHeartbeat(heartbeat.NodeID, status, heartbeat.CurrentUsers, &now, heartbeat.RealityPublicKey)
}

// GetUserSyncData retrieves users for a specific node to sync.
// Only returns users whose subscription plans include this node.
func (s *NodeAgentService) GetUserSyncData(nodeID uint) (*models.NodeUserSync, error) {
	// Verify node exists
	_, err := s.nodeRepo.GetByID(nodeID)
	if err != nil {
		return nil, err
	}

	// Get only users whose plans include this specific node
	users, err := s.userRepo.GetActiveUsersForNode(nodeID)
	if err != nil {
		return nil, err
	}

	// Build user configs
	userConfigs := make([]models.UserNodeConfig, 0, len(users))
	for _, user := range users {
		// Skip users without subscriptions
		if user.Subscription == nil {
			continue
		}

		// Only include users with active status (not suspended)
		if user.Status != models.UserStatusActive {
			continue
		}

		// Only include users with active subscriptions (not expired)
		if !user.Subscription.IsActive() {
			continue
		}

		// Only include users with remaining data quota
		if !user.Subscription.HasDataRemaining() {
			continue
		}

		userConfigs = append(userConfigs, models.UserNodeConfig{
			UserID:    user.ID,
			Email:     user.Email,
			UUID:      user.UUID,
			Status:    string(user.Status),
			DataLimit: user.Subscription.GetDataLimitBytes(),
			DataUsed:  user.Subscription.DataUsedBytes,
		})

	}

	return &models.NodeUserSync{
		Users: userConfigs,
	}, nil
}

// ProcessTrafficReport processes traffic data reported by a node.
func (s *NodeAgentService) ProcessTrafficReport(report *models.NodeTrafficReport) error {
	// Verify node exists
	_, err := s.nodeRepo.GetByID(report.NodeID)
	if err != nil {
		return err
	}

	// Process each user's traffic
	for _, userTraffic := range report.Traffic {
		// Find user by email
		user, err := s.userRepo.GetByEmail(userTraffic.UserEmail)
		if err != nil {
			continue // Skip if user not found
		}

		// Create traffic log
		trafficLog := &models.TrafficLog{
			UserID:        user.ID,
			NodeID:        report.NodeID,
			UploadBytes:   userTraffic.UploadBytes,
			DownloadBytes: userTraffic.DownloadBytes,
			RecordedAt:    report.Timestamp,
		}

		if err := s.trafficRepo.Create(trafficLog); err != nil {
			continue // Skip on error
		}

		// Update subscription data usage
		totalBytes := userTraffic.UploadBytes + userTraffic.DownloadBytes
		_ = s.userSubRepo.UpdateDataUsage(user.ID, totalBytes)
	}

	return nil
}
