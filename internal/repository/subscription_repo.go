package repository

import (
	"errors"
	"time"

	"xpanel/internal/models"

	"gorm.io/gorm"
)

// Subscription errors
var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

// SubscriptionRepository handles subscription database operations.
type SubscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository creates a new subscription repository instance.
func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create inserts a new subscription into the database.
func (r *SubscriptionRepository) Create(sub *models.Subscription) error {
	return r.db.Create(sub).Error
}

// GetByID retrieves a subscription by its ID.
func (r *SubscriptionRepository) GetByID(id uint) (*models.Subscription, error) {
	var sub models.Subscription
	result := r.db.First(&sub, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, result.Error
	}
	return &sub, nil
}

// GetByUserID retrieves a subscription by user ID.
func (r *SubscriptionRepository) GetByUserID(userID uint) (*models.Subscription, error) {
	var sub models.Subscription
	result := r.db.Where("user_id = ?", userID).First(&sub)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, result.Error
	}
	return &sub, nil
}

// Update updates subscription fields.
func (r *SubscriptionRepository) Update(sub *models.Subscription) error {
	return r.db.Save(sub).Error
}

// UpdateStatus updates the subscription status.
func (r *SubscriptionRepository) UpdateStatus(id uint, status models.SubscriptionStatus) error {
	return r.db.Model(&models.Subscription{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateDataUsage updates the data usage counters.
func (r *SubscriptionRepository) UpdateDataUsage(userID uint, bytesUsed int64) error {
	return r.db.Model(&models.Subscription{}).
		Where("user_id = ?", userID).
		UpdateColumn("data_used_bytes", gorm.Expr("data_used_bytes + ?", bytesUsed)).
		Error
}

// GetExpiredSubscriptions retrieves all subscriptions that have expired but not yet marked.
func (r *SubscriptionRepository) GetExpiredSubscriptions() ([]models.Subscription, error) {
	var subs []models.Subscription
	result := r.db.Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?",
		models.SubscriptionActive, time.Now()).Find(&subs)
	if result.Error != nil {
		return nil, result.Error
	}
	return subs, nil
}

// MarkExpired marks multiple subscriptions as expired in batch.
func (r *SubscriptionRepository) MarkExpired(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&models.Subscription{}).
		Where("id IN ?", ids).
		Update("status", models.SubscriptionExpired).
		Error
}

// Renew updates subscription for renewal with new plan details.
func (r *SubscriptionRepository) Renew(userID uint, plan models.PlanType, expiresAt *time.Time, dataLimit int64) error {
	updates := map[string]interface{}{
		"plan":             plan,
		"status":           models.SubscriptionActive,
		"expires_at":       expiresAt,
		"data_limit_bytes": dataLimit,
		"data_used_bytes":  0,
		"start_date":       time.Now(),
	}
	return r.db.Model(&models.Subscription{}).Where("user_id = ?", userID).Updates(updates).Error
}

// GetActiveSubscriptions retrieves all active subscriptions.
func (r *SubscriptionRepository) GetActiveSubscriptions() ([]models.Subscription, error) {
	var subs []models.Subscription
	result := r.db.Preload("User").Where("status = ?", models.SubscriptionActive).Find(&subs)
	if result.Error != nil {
		return nil, result.Error
	}
	return subs, nil
}
