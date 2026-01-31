package models

import (
	"time"

	"gorm.io/gorm"
)

// UserSubscription represents a user's active subscription instance.
// This links a user to a subscription plan template.
type UserSubscription struct {
	ID            uint               `gorm:"primaryKey" json:"id"`
	UserID        uint               `gorm:"unique;not null" json:"user_id"`
	PlanID        uint               `gorm:"not null" json:"plan_id"`
	Status        SubscriptionStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	DataUsedBytes int64              `gorm:"default:0" json:"data_used_bytes"`
	StartDate     time.Time          `json:"start_date"`
	ExpiresAt     *time.Time         `json:"expires_at,omitempty"`
	AutoRenew     bool               `gorm:"default:false" json:"auto_renew"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	DeletedAt     gorm.DeletedAt     `gorm:"index" json:"-"`

	// Relationships
	User *User             `gorm:"foreignKey:UserID" json:"-"`
	Plan *SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

// IsActive checks if the subscription is currently active and not expired.
func (s *UserSubscription) IsActive() bool {
	if s.Status != SubscriptionActive {
		return false
	}
	if s.ExpiresAt != nil && time.Now().After(*s.ExpiresAt) {
		return false
	}
	return true
}

// HasDataRemaining checks if the user has remaining data quota.
func (s *UserSubscription) HasDataRemaining() bool {
	if s.Plan == nil {
		return true
	}
	dataLimit := s.Plan.GetDataLimitBytes()
	if dataLimit == 0 {
		return true // Unlimited
	}
	return s.DataUsedBytes < dataLimit
}

// DaysRemaining returns the number of days until subscription expires.
func (s *UserSubscription) DaysRemaining() int {
	if s.ExpiresAt == nil {
		return -1 // Never expires
	}
	remaining := time.Until(*s.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return int(remaining.Hours() / 24)
}

// GetDataLimitBytes returns the data limit in bytes from the plan.
func (s *UserSubscription) GetDataLimitBytes() int64 {
	if s.Plan == nil {
		return 0
	}
	return s.Plan.GetDataLimitBytes()
}

// UserSubscriptionResponse is the API response structure for user subscriptions.
type UserSubscriptionResponse struct {
	ID             uint                      `json:"id"`
	UserID         uint                      `json:"user_id"`
	UserEmail      string                    `json:"user_email,omitempty"`
	Plan           *SubscriptionPlanResponse `json:"plan,omitempty"`
	Status         SubscriptionStatus        `json:"status"`
	DataUsedBytes  int64                     `json:"data_used_bytes"`
	DataLimitBytes int64                     `json:"data_limit_bytes"`
	StartDate      time.Time                 `json:"start_date"`
	ExpiresAt      *time.Time                `json:"expires_at,omitempty"`
	DaysRemaining  int                       `json:"days_remaining"`
	AutoRenew      bool                      `json:"auto_renew"`
	IsActive       bool                      `json:"is_active"`
	CreatedAt      time.Time                 `json:"created_at"`
}

// ToResponse converts UserSubscription to a response structure.
func (s *UserSubscription) ToResponse() UserSubscriptionResponse {
	resp := UserSubscriptionResponse{
		ID:             s.ID,
		UserID:         s.UserID,
		Status:         s.Status,
		DataUsedBytes:  s.DataUsedBytes,
		DataLimitBytes: s.GetDataLimitBytes(),
		StartDate:      s.StartDate,
		ExpiresAt:      s.ExpiresAt,
		DaysRemaining:  s.DaysRemaining(),
		AutoRenew:      s.AutoRenew,
		IsActive:       s.IsActive(),
		CreatedAt:      s.CreatedAt,
	}

	if s.User != nil {
		resp.UserEmail = s.User.Email
	}

	if s.Plan != nil {
		planResp := s.Plan.ToResponse()
		resp.Plan = &planResp
	}

	return resp
}
