package repository

import (
	"time"

	"xpanel/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ActivityRepository handles database operations for user activities.
type ActivityRepository struct {
	db *gorm.DB
}

// NewActivityRepository creates a new activity repository.
func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// UpdateUserActivity updates or creates activity record for a user.
// Uses upsert to handle both new and existing activity records.
func (r *ActivityRepository) UpdateUserActivity(userID, nodeID uint, lastSeen time.Time, isOnline bool) error {
	activity := models.UserActivity{
		UserID:    userID,
		NodeID:    nodeID,
		LastSeen:  lastSeen,
		IsOnline:  isOnline,
		UpdatedAt: time.Now(),
	}

	// Upsert: insert or update on conflict
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"node_id", "last_seen", "is_online", "updated_at"}),
	}).Create(&activity).Error
}

// GetUserActivity retrieves activity for a specific user.
func (r *ActivityRepository) GetUserActivity(userID uint) (*models.UserActivity, error) {
	var activity models.UserActivity
	err := r.db.Preload("Node").Where("user_id = ?", userID).First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

// GetAllActivities retrieves all user activities.
func (r *ActivityRepository) GetAllActivities() ([]models.UserActivity, error) {
	var activities []models.UserActivity
	err := r.db.Preload("Node").Preload("User").Find(&activities).Error
	return activities, err
}

// GetOnlineUsers retrieves all currently online users (last seen within threshold).
func (r *ActivityRepository) GetOnlineUsers(threshold time.Duration) ([]models.UserActivity, error) {
	var activities []models.UserActivity
	cutoff := time.Now().Add(-threshold)

	err := r.db.Preload("Node").Preload("User").
		Where("last_seen > ?", cutoff).
		Find(&activities).Error

	return activities, err
}

// GetOnlineCount returns the count of online users.
func (r *ActivityRepository) GetOnlineCount(threshold time.Duration) (int64, error) {
	var count int64
	cutoff := time.Now().Add(-threshold)

	err := r.db.Model(&models.UserActivity{}).
		Where("last_seen > ?", cutoff).
		Count(&count).Error

	return count, err
}

// MarkAllOffline marks all users as offline (e.g., when a node goes offline).
func (r *ActivityRepository) MarkAllOffline() error {
	return r.db.Model(&models.UserActivity{}).
		Where("is_online = ?", true).
		Update("is_online", false).Error
}

// MarkNodeUsersOffline marks all users on a specific node as offline.
func (r *ActivityRepository) MarkNodeUsersOffline(nodeID uint) error {
	return r.db.Model(&models.UserActivity{}).
		Where("node_id = ? AND is_online = ?", nodeID, true).
		Update("is_online", false).Error
}

// CleanupStaleActivities removes activity records for users who haven't been seen in a long time.
func (r *ActivityRepository) CleanupStaleActivities(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return r.db.Where("last_seen < ?", cutoff).Delete(&models.UserActivity{}).Error
}

// GetActivitiesForUsers retrieves activities for specific user IDs.
func (r *ActivityRepository) GetActivitiesForUsers(userIDs []uint) (map[uint]*models.UserActivity, error) {
	var activities []models.UserActivity
	err := r.db.Preload("Node").Where("user_id IN ?", userIDs).Find(&activities).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]*models.UserActivity)
	for i := range activities {
		result[activities[i].UserID] = &activities[i]
	}
	return result, nil
}

// CountOnlineUsersByNode counts online users for a specific node.
func (r *ActivityRepository) CountOnlineUsersByNode(nodeID uint, threshold time.Duration) (int64, error) {
	var count int64
	cutoff := time.Now().Add(-threshold)

	err := r.db.Model(&models.UserActivity{}).
		Where("node_id = ? AND last_seen > ?", nodeID, cutoff).
		Count(&count).Error

	return count, err
}
