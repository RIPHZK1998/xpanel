package service

import (
	"errors"
	"time"

	"xpanel/internal/models"
	"xpanel/internal/repository"
)

// Subscription errors
var (
	ErrNoActiveSubscription  = errors.New("no active subscription found")
	ErrSubscriptionSuspended = errors.New("subscription is suspended")
	ErrDataLimitExceeded     = errors.New("data limit exceeded")
	ErrInvalidPlan           = errors.New("invalid subscription plan")
)

// SubscriptionService handles subscription-related business logic.
type SubscriptionService struct {
	subRepo  *repository.SubscriptionRepository
	userRepo *repository.UserRepository
	planRepo *repository.PlanRepository
}

// NewSubscriptionService creates a new subscription service.
func NewSubscriptionService(
	subRepo *repository.SubscriptionRepository,
	userRepo *repository.UserRepository,
	planRepo *repository.PlanRepository,
) *SubscriptionService {
	return &SubscriptionService{
		subRepo:  subRepo,
		userRepo: userRepo,
		planRepo: planRepo,
	}
}

// GetSubscription retrieves a user's subscription.
func (s *SubscriptionService) GetSubscription(userID uint) (*models.UserSubscription, error) {
	return s.subRepo.GetByUserID(userID)
}

// RenewRequest contains subscription renewal data.
type RenewRequest struct {
	PlanID    uint `json:"plan_id" binding:"required"`
	AutoRenew bool `json:"auto_renew"`
}

// Renew renews or upgrades a user's subscription.
func (s *SubscriptionService) Renew(userID uint, planID uint, autoRenew bool) (*models.UserSubscription, error) {
	// Get plan details
	plan, err := s.planRepo.GetByID(planID)
	if err != nil {
		return nil, ErrInvalidPlan
	}

	// Get current subscription
	sub, err := s.subRepo.GetByUserID(userID)
	if err != nil && !errors.Is(err, repository.ErrSubscriptionNotFound) {
		return nil, err
	}

	var expiresAt *time.Time
	durationDays := plan.GetDurationDays()
	if durationDays > 0 {
		expiry := time.Now().AddDate(0, 0, durationDays)
		expiresAt = &expiry
	}

	if sub == nil {
		// Create new subscription
		sub = &models.UserSubscription{
			UserID:    userID,
			PlanID:    plan.ID,
			Status:    models.SubscriptionActive,
			StartDate: time.Now(),
			ExpiresAt: expiresAt,
			AutoRenew: autoRenew,
		}
		if err := s.subRepo.Create(sub); err != nil {
			return nil, err
		}
	} else {
		// Update existing subscription
		if err := s.subRepo.Renew(userID, plan.ID, expiresAt, autoRenew); err != nil {
			return nil, err
		}
		// Refresh subscription data
		sub, err = s.subRepo.GetByUserID(userID)
		if err != nil {
			return nil, err
		}
	}

	// Load plan data into sub for return
	sub.Plan = plan
	return sub, nil
}

// Suspend suspends a user's subscription.
func (s *SubscriptionService) Suspend(userID uint) error {
	sub, err := s.subRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	return s.subRepo.UpdateStatus(sub.ID, models.SubscriptionSuspended)
}

// Activate activates a suspended subscription.
func (s *SubscriptionService) Activate(userID uint) error {
	sub, err := s.subRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	return s.subRepo.UpdateStatus(sub.ID, models.SubscriptionActive)
}

// CheckAndExpire checks for expired subscriptions and marks them.
func (s *SubscriptionService) CheckAndExpire() (int, error) {
	expired, err := s.subRepo.GetExpiredSubscriptions()
	if err != nil {
		return 0, err
	}

	if len(expired) == 0 {
		return 0, nil
	}

	ids := make([]uint, len(expired))
	for i, sub := range expired {
		ids[i] = sub.ID
	}

	if err := s.subRepo.MarkExpired(ids); err != nil {
		return 0, err
	}

	return len(expired), nil
}

// UpdateDataUsage updates the data usage for a user's subscription.
func (s *SubscriptionService) UpdateDataUsage(userID uint, bytesUsed int64) error {
	return s.subRepo.UpdateDataUsage(userID, bytesUsed)
}

// CheckDataLimit checks if a user has exceeded their data limit.
func (s *SubscriptionService) CheckDataLimit(userID uint) (bool, error) {
	sub, err := s.subRepo.GetByUserID(userID)
	if err != nil {
		return false, err
	}
	return sub.HasDataRemaining(), nil
}

// IsActive checks if a user's subscription is active.
func (s *SubscriptionService) IsActive(userID uint) (bool, error) {
	sub, err := s.subRepo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNotFound) {
			return false, nil
		}
		return false, err
	}
	return sub.IsActive(), nil
}

// GetActiveSubscriptions retrieves all active subscriptions with users.
func (s *SubscriptionService) GetActiveSubscriptions() ([]models.UserSubscription, error) {
	return s.subRepo.GetActiveSubscriptions()
}
