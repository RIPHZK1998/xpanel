// Package repository provides database access layer implementations.
package repository

import (
	"errors"

	"xpanel/internal/models"

	"gorm.io/gorm"
)

// Common errors
var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUserExists    = errors.New("user already exists")
	ErrInvalidUserID = errors.New("invalid user id")
)

// UserRepository handles user database operations.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(user *models.User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrUserExists
		}
		return result.Error
	}
	return nil
}

// GetByID retrieves a user by their ID.
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetByEmail retrieves a user by their email address.
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetByUUID retrieves a user by their xray UUID.
func (r *UserRepository) GetByUUID(uuid string) (*models.User, error) {
	var user models.User
	result := r.db.Where("uuid = ?", uuid).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetByIDWithSubscription retrieves a user with their subscription loaded.
func (r *UserRepository) GetByIDWithSubscription(id uint) (*models.User, error) {
	var user models.User
	result := r.db.Preload("Subscription").Preload("Subscription.Plan").Preload("Subscription.Plan.Nodes").First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// Update updates user fields.
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// UpdateStatus updates the user's status.
func (r *UserRepository) UpdateStatus(id uint, status models.UserStatus) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("status", status).Error
}

// Delete soft-deletes a user.
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// ExistsByEmail checks if a user with the given email exists.
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	result := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// GetActiveUsers retrieves all active users.
func (r *UserRepository) GetActiveUsers() ([]models.User, error) {
	var users []models.User
	result := r.db.Where("status = ?", models.UserStatusActive).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// GetActiveUsersWithSubscription retrieves all active users with their subscriptions.
func (r *UserRepository) GetActiveUsersWithSubscription() ([]models.User, error) {
	var users []models.User
	result := r.db.Preload("Subscription").Preload("Subscription.Plan").Where("status = ?", models.UserStatusActive).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// GetAllUsersWithSubscription retrieves all users (regardless of status) with their subscriptions.
func (r *UserRepository) GetAllUsersWithSubscription() ([]models.User, error) {
	var users []models.User
	result := r.db.Preload("Subscription").Preload("Subscription.Plan").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// GetActiveUsersForNode retrieves active users whose plans include the specified node.
// This uses a subquery to filter users by plan-node relationship.
func (r *UserRepository) GetActiveUsersForNode(nodeID uint) ([]models.User, error) {
	var users []models.User

	// Get users where:
	// 1. User status is active
	// 2. User has a subscription
	// 3. The subscription's plan has this node in the plan_nodes join table
	result := r.db.
		Preload("Subscription").
		Preload("Subscription.Plan").
		Joins("JOIN user_subscriptions ON user_subscriptions.user_id = users.id").
		Joins("JOIN plan_nodes ON plan_nodes.subscription_plan_id = user_subscriptions.plan_id").
		Where("users.status = ?", models.UserStatusActive).
		Where("plan_nodes.node_id = ?", nodeID).
		Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// CountActive returns the count of active users.
func (r *UserRepository) CountActive() (int64, error) {
	var count int64
	result := r.db.Model(&models.User{}).Where("status = ?", models.UserStatusActive).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}
