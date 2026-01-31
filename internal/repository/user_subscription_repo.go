package repository

import (
	"errors"

	"xpanel/internal/models"

	"gorm.io/gorm"
)

// UserSubscription errors
var (
	ErrUserSubscriptionNotFound = errors.New("user subscription not found")
)

// UserSubscriptionRepository handles user subscription database operations.
type UserSubscriptionRepository struct {
	db *gorm.DB
}

// NewUserSubscriptionRepository creates a new user subscription repository instance.
func NewUserSubscriptionRepository(db *gorm.DB) *UserSubscriptionRepository {
	return &UserSubscriptionRepository{db: db}
}

// Create inserts a new user subscription into the database.
func (r *UserSubscriptionRepository) Create(sub *models.UserSubscription) error {
	return r.db.Create(sub).Error
}

// GetByID retrieves a user subscription by its ID.
func (r *UserSubscriptionRepository) GetByID(id uint) (*models.UserSubscription, error) {
	var sub models.UserSubscription
	result := r.db.Preload("User").Preload("Plan").Preload("Plan.Nodes").First(&sub, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserSubscriptionNotFound
		}
		return nil, result.Error
	}
	return &sub, nil
}

// GetByUserID retrieves a user's subscription.
func (r *UserSubscriptionRepository) GetByUserID(userID uint) (*models.UserSubscription, error) {
	var sub models.UserSubscription
	result := r.db.Preload("User").Preload("Plan").Preload("Plan.Nodes").
		Where("user_id = ?", userID).First(&sub)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserSubscriptionNotFound
		}
		return nil, result.Error
	}
	return &sub, nil
}

// Update updates user subscription fields.
func (r *UserSubscriptionRepository) Update(sub *models.UserSubscription) error {
	return r.db.Save(sub).Error
}

// UpdateStatus updates the subscription status.
func (r *UserSubscriptionRepository) UpdateStatus(id uint, status models.SubscriptionStatus) error {
	return r.db.Model(&models.UserSubscription{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// UpdateDataUsage updates the data usage counters.
func (r *UserSubscriptionRepository) UpdateDataUsage(userID uint, bytesUsed int64) error {
	return r.db.Model(&models.UserSubscription{}).
		Where("user_id = ?", userID).
		UpdateColumn("data_used_bytes", gorm.Expr("data_used_bytes + ?", bytesUsed)).
		Error
}

// ResetDataUsage resets the data usage to zero.
func (r *UserSubscriptionRepository) ResetDataUsage(userID uint) error {
	return r.db.Model(&models.UserSubscription{}).
		Where("user_id = ?", userID).
		Update("data_used_bytes", 0).Error
}

// Delete soft-deletes a user subscription.
func (r *UserSubscriptionRepository) Delete(id uint) error {
	return r.db.Delete(&models.UserSubscription{}, id).Error
}

// GetAll retrieves all user subscriptions with plan and user data.
func (r *UserSubscriptionRepository) GetAll() ([]models.UserSubscription, error) {
	var subs []models.UserSubscription
	result := r.db.Preload("User").Preload("Plan").Find(&subs)
	if result.Error != nil {
		return nil, result.Error
	}
	return subs, nil
}

// GetActiveSubscriptions retrieves all active subscriptions with users and plans.
func (r *UserSubscriptionRepository) GetActiveSubscriptions() ([]models.UserSubscription, error) {
	var subs []models.UserSubscription
	result := r.db.Preload("User").Preload("Plan").
		Where("status = ?", models.SubscriptionActive).Find(&subs)
	if result.Error != nil {
		return nil, result.Error
	}
	return subs, nil
}

// ChangePlan changes a user's subscription to a different plan.
func (r *UserSubscriptionRepository) ChangePlan(userID, newPlanID uint) error {
	return r.db.Model(&models.UserSubscription{}).
		Where("user_id = ?", userID).
		Update("plan_id", newPlanID).Error
}
