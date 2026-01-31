package repository

import (
	"errors"

	"xpanel/internal/models"

	"gorm.io/gorm"
)

// Plan errors
var (
	ErrPlanNotFound = errors.New("subscription plan not found")
	ErrPlanExists   = errors.New("plan with this name already exists")
)

// PlanRepository handles subscription plan database operations.
type PlanRepository struct {
	db *gorm.DB
}

// NewPlanRepository creates a new plan repository instance.
func NewPlanRepository(db *gorm.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

// Create inserts a new subscription plan into the database.
func (r *PlanRepository) Create(plan *models.SubscriptionPlan) error {
	result := r.db.Create(plan)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrPlanExists
		}
		return result.Error
	}
	return nil
}

// GetByID retrieves a plan by its ID with nodes preloaded.
func (r *PlanRepository) GetByID(id uint) (*models.SubscriptionPlan, error) {
	var plan models.SubscriptionPlan
	result := r.db.Preload("Nodes").Preload("UserSubscriptions").First(&plan, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrPlanNotFound
		}
		return nil, result.Error
	}
	return &plan, nil
}

// GetByName retrieves a plan by its name.
func (r *PlanRepository) GetByName(name string) (*models.SubscriptionPlan, error) {
	var plan models.SubscriptionPlan
	result := r.db.Preload("Nodes").Where("name = ?", name).First(&plan)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrPlanNotFound
		}
		return nil, result.Error
	}
	return &plan, nil
}

// GetAll retrieves all subscription plans.
func (r *PlanRepository) GetAll() ([]models.SubscriptionPlan, error) {
	var plans []models.SubscriptionPlan
	result := r.db.Preload("Nodes").Preload("UserSubscriptions").Find(&plans)
	if result.Error != nil {
		return nil, result.Error
	}
	return plans, nil
}

// GetActive retrieves all active (non-archived) subscription plans.
func (r *PlanRepository) GetActive() ([]models.SubscriptionPlan, error) {
	var plans []models.SubscriptionPlan
	result := r.db.Preload("Nodes").Where("status = ?", models.PlanStatusActive).Find(&plans)
	if result.Error != nil {
		return nil, result.Error
	}
	return plans, nil
}

// Update updates a subscription plan's fields.
func (r *PlanRepository) Update(plan *models.SubscriptionPlan) error {
	return r.db.Save(plan).Error
}

// Delete soft-deletes a subscription plan.
func (r *PlanRepository) Delete(id uint) error {
	return r.db.Delete(&models.SubscriptionPlan{}, id).Error
}

// Archive sets a plan's status to archived.
func (r *PlanRepository) Archive(id uint) error {
	return r.db.Model(&models.SubscriptionPlan{}).
		Where("id = ?", id).
		Update("status", models.PlanStatusArchived).Error
}

// AssignNodes assigns nodes to a plan (replaces existing assignments).
func (r *PlanRepository) AssignNodes(planID uint, nodeIDs []uint) error {
	var plan models.SubscriptionPlan
	if err := r.db.First(&plan, planID).Error; err != nil {
		return err
	}

	// Load nodes
	var nodes []models.Node
	if len(nodeIDs) > 0 {
		if err := r.db.Find(&nodes, nodeIDs).Error; err != nil {
			return err
		}
	}

	// Replace associations
	return r.db.Model(&plan).Association("Nodes").Replace(nodes)
}

// GetPlanNodes retrieves all nodes assigned to a plan.
func (r *PlanRepository) GetPlanNodes(planID uint) ([]models.Node, error) {
	var plan models.SubscriptionPlan
	if err := r.db.Preload("Nodes").First(&plan, planID).Error; err != nil {
		return nil, err
	}
	return plan.Nodes, nil
}

// GetPlanUsers retrieves all users subscribed to a plan.
func (r *PlanRepository) GetPlanUsers(planID uint) ([]models.UserSubscription, error) {
	var subscriptions []models.UserSubscription
	result := r.db.Preload("User").Where("plan_id = ?", planID).Find(&subscriptions)
	if result.Error != nil {
		return nil, result.Error
	}
	return subscriptions, nil
}

// CountActiveUsers counts how many users are currently on a plan.
func (r *PlanRepository) CountActiveUsers(planID uint) (int64, error) {
	var count int64
	result := r.db.Model(&models.UserSubscription{}).
		Where("plan_id = ? AND status = ?", planID, models.SubscriptionActive).
		Count(&count)
	return count, result.Error
}
