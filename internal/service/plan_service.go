package service

import (
	"errors"
	"time"

	"xpanel/internal/models"
	"xpanel/internal/repository"
)

// Plan service errors
var (
	ErrPlanInUse = errors.New("cannot delete plan that has active users")
)

// PlanService handles subscription plan business logic.
type PlanService struct {
	planRepo    *repository.PlanRepository
	userSubRepo *repository.UserSubscriptionRepository
}

// NewPlanService creates a new plan service.
func NewPlanService(
	planRepo *repository.PlanRepository,
	userSubRepo *repository.UserSubscriptionRepository,
) *PlanService {
	return &PlanService{
		planRepo:    planRepo,
		userSubRepo: userSubRepo,
	}
}

// CreatePlanRequest contains data for creating a new plan.
type CreatePlanRequest struct {
	Name        string              `json:"name" binding:"required"`
	DisplayName string              `json:"display_name" binding:"required"`
	Duration    models.PlanDuration `json:"duration" binding:"required,oneof=weekly monthly quarterly annual"`
	Price       float64             `json:"price" binding:"min=0"`
	DataLimitGB int64               `json:"data_limit_gb" binding:"min=0"`
	MaxDevices  int                 `json:"max_devices" binding:"min=1"`
	Description string              `json:"description"`
	Features    string              `json:"features"`
	NodeIDs     []uint              `json:"node_ids"`
}

// UpdatePlanRequest contains data for updating a plan.
type UpdatePlanRequest struct {
	DisplayName *string              `json:"display_name"`
	Duration    *models.PlanDuration `json:"duration"`
	Price       *float64             `json:"price"`
	DataLimitGB *int64               `json:"data_limit_gb"`
	MaxDevices  *int                 `json:"max_devices"`
	Description *string              `json:"description"`
	Features    *string              `json:"features"`
	Status      *models.PlanStatus   `json:"status"`
}

// CreatePlan creates a new subscription plan.
func (s *PlanService) CreatePlan(req *CreatePlanRequest) (*models.SubscriptionPlan, error) {
	plan := &models.SubscriptionPlan{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Duration:    req.Duration,
		Price:       req.Price,
		DataLimitGB: req.DataLimitGB,
		MaxDevices:  req.MaxDevices,
		Status:      models.PlanStatusActive,
		Description: req.Description,
		Features:    req.Features,
	}

	if err := s.planRepo.Create(plan); err != nil {
		return nil, err
	}

	// Assign nodes if provided
	if len(req.NodeIDs) > 0 {
		if err := s.planRepo.AssignNodes(plan.ID, req.NodeIDs); err != nil {
			return nil, err
		}
	}

	// Reload with relationships
	return s.planRepo.GetByID(plan.ID)
}

// GetPlan retrieves a plan by ID.
func (s *PlanService) GetPlan(id uint) (*models.SubscriptionPlan, error) {
	return s.planRepo.GetByID(id)
}

// GetAllPlans retrieves all subscription plans.
func (s *PlanService) GetAllPlans() ([]models.SubscriptionPlan, error) {
	return s.planRepo.GetAll()
}

// GetActivePlans retrieves all active plans.
func (s *PlanService) GetActivePlans() ([]models.SubscriptionPlan, error) {
	return s.planRepo.GetActive()
}

// UpdatePlan updates a subscription plan.
func (s *PlanService) UpdatePlan(id uint, req *UpdatePlanRequest) (*models.SubscriptionPlan, error) {
	plan, err := s.planRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.DisplayName != nil {
		plan.DisplayName = *req.DisplayName
	}
	if req.Duration != nil {
		plan.Duration = *req.Duration
	}
	if req.Price != nil {
		plan.Price = *req.Price
	}
	if req.DataLimitGB != nil {
		plan.DataLimitGB = *req.DataLimitGB
	}
	if req.MaxDevices != nil {
		plan.MaxDevices = *req.MaxDevices
	}
	if req.Description != nil {
		plan.Description = *req.Description
	}
	if req.Features != nil {
		plan.Features = *req.Features
	}
	if req.Status != nil {
		plan.Status = *req.Status
	}

	if err := s.planRepo.Update(plan); err != nil {
		return nil, err
	}

	return s.planRepo.GetByID(id)
}

// ArchivePlan archives a subscription plan.
func (s *PlanService) ArchivePlan(id uint) error {
	// Check if plan has active users
	count, err := s.planRepo.CountActiveUsers(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrPlanInUse
	}

	return s.planRepo.Archive(id)
}

// DeletePlan deletes a subscription plan.
func (s *PlanService) DeletePlan(id uint) error {
	// Check if plan has any users (active or not)
	users, err := s.planRepo.GetPlanUsers(id)
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return ErrPlanInUse
	}

	return s.planRepo.Delete(id)
}

// AssignNodesToPlan assigns nodes to a plan.
func (s *PlanService) AssignNodesToPlan(planID uint, nodeIDs []uint) error {
	return s.planRepo.AssignNodes(planID, nodeIDs)
}

// GetPlanNodes retrieves all nodes for a plan.
func (s *PlanService) GetPlanNodes(planID uint) ([]models.Node, error) {
	return s.planRepo.GetPlanNodes(planID)
}

// GetPlanUsers retrieves all users subscribed to a plan.
func (s *PlanService) GetPlanUsers(planID uint) ([]models.UserSubscription, error) {
	return s.planRepo.GetPlanUsers(planID)
}

// AssignPlanToUserRequest contains data for assigning a plan to a user.
type AssignPlanToUserRequest struct {
	UserID    uint       `json:"user_id" binding:"required"`
	PlanID    uint       `json:"plan_id" binding:"required"`
	AutoRenew bool       `json:"auto_renew"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// AssignPlanToUser assigns or changes a user's subscription plan.
func (s *PlanService) AssignPlanToUser(req *AssignPlanToUserRequest) (*models.UserSubscription, error) {
	// Get the plan to calculate expiration
	plan, err := s.planRepo.GetByID(req.PlanID)
	if err != nil {
		return nil, err
	}

	// Check if user already has a subscription
	existingSub, err := s.userSubRepo.GetByUserID(req.UserID)
	if err != nil && !errors.Is(err, repository.ErrUserSubscriptionNotFound) {
		return nil, err
	}

	// Calculate expiration date if not provided
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		expiresAt = req.ExpiresAt
	} else {
		durationDays := plan.GetDurationDays()
		if durationDays > 0 {
			expiry := time.Now().AddDate(0, 0, durationDays)
			expiresAt = &expiry
		}
	}

	if existingSub != nil {
		// Update existing subscription
		existingSub.PlanID = req.PlanID
		existingSub.Status = models.SubscriptionActive
		existingSub.StartDate = time.Now()
		existingSub.ExpiresAt = expiresAt
		existingSub.AutoRenew = req.AutoRenew
		existingSub.DataUsedBytes = 0 // Reset data usage on plan change

		if err := s.userSubRepo.Update(existingSub); err != nil {
			return nil, err
		}
		return s.userSubRepo.GetByUserID(req.UserID)
	}

	// Create new subscription
	newSub := &models.UserSubscription{
		UserID:        req.UserID,
		PlanID:        req.PlanID,
		Status:        models.SubscriptionActive,
		DataUsedBytes: 0,
		StartDate:     time.Now(),
		ExpiresAt:     expiresAt,
		AutoRenew:     req.AutoRenew,
	}

	if err := s.userSubRepo.Create(newSub); err != nil {
		return nil, err
	}

	return s.userSubRepo.GetByUserID(req.UserID)
}
