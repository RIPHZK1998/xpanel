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
}

// NewSubscriptionService creates a new subscription service.
func NewSubscriptionService(
	subRepo *repository.SubscriptionRepository,
	userRepo *repository.UserRepository,
) *SubscriptionService {
	return &SubscriptionService{
		subRepo:  subRepo,
		userRepo: userRepo,
	}
}

// GetSubscription retrieves a user's subscription.
func (s *SubscriptionService) GetSubscription(userID uint) (*models.Subscription, error) {
	return s.subRepo.GetByUserID(userID)
}

// RenewRequest contains subscription renewal data.
type RenewRequest struct {
	Plan models.PlanType `json:"plan" binding:"required,oneof=free monthly yearly"`
}

// Renew renews or upgrades a user's subscription.
func (s *SubscriptionService) Renew(userID uint, plan models.PlanType) (*models.Subscription, error) {
	// Validate plan
	if plan != models.PlanFree && plan != models.PlanMonthly && plan != models.PlanYearly {
		return nil, ErrInvalidPlan
	}

	// Get current subscription
	sub, err := s.subRepo.GetByUserID(userID)
	if err != nil && !errors.Is(err, repository.ErrSubscriptionNotFound) {
		return nil, err
	}

	// Get plan details
	dataLimitGB, durationDays := models.GetPlanDetails(plan)
	dataLimitBytes := dataLimitGB * 1024 * 1024 * 1024

	var expiresAt *time.Time
	if durationDays > 0 {
		expiry := time.Now().AddDate(0, 0, durationDays)
		expiresAt = &expiry
	}

	if sub == nil {
		// Create new subscription
		sub = &models.Subscription{
			UserID:         userID,
			Plan:           plan,
			Status:         models.SubscriptionActive,
			DataLimitBytes: dataLimitBytes,
			StartDate:      time.Now(),
			ExpiresAt:      expiresAt,
		}
		if err := s.subRepo.Create(sub); err != nil {
			return nil, err
		}
	} else {
		// Update existing subscription
		if err := s.subRepo.Renew(userID, plan, expiresAt, dataLimitBytes); err != nil {
			return nil, err
		}
		// Refresh subscription data
		sub, err = s.subRepo.GetByUserID(userID)
		if err != nil {
			return nil, err
		}
	}

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
func (s *SubscriptionService) GetActiveSubscriptions() ([]models.Subscription, error) {
	return s.subRepo.GetActiveSubscriptions()
}
