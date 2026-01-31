package service

import (
	"fmt"
	"time"

	"xpanel/internal/models"
	"xpanel/internal/repository"
)

const (
	// OnlineThreshold defines how long since last activity a user is considered online.
	OnlineThreshold = 2 * time.Minute
)

// ActivityService handles user activity business logic.
type ActivityService struct {
	activityRepo *repository.ActivityRepository
	userRepo     *repository.UserRepository
}

// NewActivityService creates a new activity service.
func NewActivityService(activityRepo *repository.ActivityRepository, userRepo *repository.UserRepository) *ActivityService {
	return &ActivityService{
		activityRepo: activityRepo,
		userRepo:     userRepo,
	}
}

// ProcessActivityReport processes an activity report from a node agent.
func (s *ActivityService) ProcessActivityReport(nodeID uint, activities []models.NodeUserActivityReport) error {
	for _, activity := range activities {
		// Look up user by email
		user, err := s.userRepo.GetByEmail(activity.Email)
		if err != nil {
			// User not found, skip
			continue
		}

		// Update activity record
		err = s.activityRepo.UpdateUserActivity(
			user.ID,
			nodeID,
			activity.LastSeen,
			activity.IsOnline,
		)
		if err != nil {
			return fmt.Errorf("failed to update activity for user %s: %w", activity.Email, err)
		}
	}

	return nil
}

// GetOnlineUsers returns all currently online users.
func (s *ActivityService) GetOnlineUsers() ([]models.UserActivity, error) {
	return s.activityRepo.GetOnlineUsers(OnlineThreshold)
}

// GetOnlineCount returns the count of online users.
func (s *ActivityService) GetOnlineCount() (int64, error) {
	return s.activityRepo.GetOnlineCount(OnlineThreshold)
}

// GetUserActivity retrieves activity for a specific user.
func (s *ActivityService) GetUserActivity(userID uint) (*models.UserActivity, error) {
	return s.activityRepo.GetUserActivity(userID)
}

// GetActivitiesForUsers retrieves activities for multiple users.
func (s *ActivityService) GetActivitiesForUsers(userIDs []uint) (map[uint]*models.UserActivity, error) {
	return s.activityRepo.GetActivitiesForUsers(userIDs)
}

// MarkNodeOffline marks all users on a node as offline (e.g., when node stops sending heartbeats).
func (s *ActivityService) MarkNodeOffline(nodeID uint) error {
	return s.activityRepo.MarkNodeUsersOffline(nodeID)
}

// GetStats returns activity statistics.
func (s *ActivityService) GetStats() (online int64, total int64, err error) {
	online, err = s.activityRepo.GetOnlineCount(OnlineThreshold)
	if err != nil {
		return 0, 0, err
	}

	total, err = s.userRepo.CountActive()
	if err != nil {
		return 0, 0, err
	}

	return online, total, nil
}
