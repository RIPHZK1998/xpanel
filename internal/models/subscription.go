package models

import (
	"time"

	"gorm.io/gorm"
)

// PlanType represents subscription plan types.
type PlanType string

const (
	PlanFree    PlanType = "free"
	PlanMonthly PlanType = "monthly"
	PlanYearly  PlanType = "yearly"
)

// SubscriptionStatus represents the status of a subscription.
type SubscriptionStatus string

const (
	SubscriptionActive    SubscriptionStatus = "active"
	SubscriptionExpired   SubscriptionStatus = "expired"
	SubscriptionSuspended SubscriptionStatus = "suspended"
	SubscriptionCanceled  SubscriptionStatus = "canceled"
)

// Subscription represents a user's VPN subscription.
type Subscription struct {
	ID             uint               `gorm:"primaryKey" json:"id"`
	UserID         uint               `gorm:"unique;not null" json:"user_id"`
	Plan           PlanType           `gorm:"type:varchar(20);default:'free'" json:"plan"`
	Status         SubscriptionStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	DataLimitBytes int64              `gorm:"default:0" json:"data_limit_bytes"` // 0 = unlimited
	DataUsedBytes  int64              `gorm:"default:0" json:"data_used_bytes"`
	StartDate      time.Time          `json:"start_date"`
	ExpiresAt      *time.Time         `json:"expires_at,omitempty"` // nil = never expires (free plan)
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	DeletedAt      gorm.DeletedAt     `gorm:"index" json:"-"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

// IsActive checks if the subscription is currently active and not expired.
func (s *Subscription) IsActive() bool {
	if s.Status != SubscriptionActive {
		return false
	}
	if s.ExpiresAt != nil && time.Now().After(*s.ExpiresAt) {
		return false
	}
	return true
}

// HasDataRemaining checks if the user has remaining data quota.
func (s *Subscription) HasDataRemaining() bool {
	if s.DataLimitBytes == 0 {
		return true // Unlimited
	}
	return s.DataUsedBytes < s.DataLimitBytes
}

// DaysRemaining returns the number of days until subscription expires.
func (s *Subscription) DaysRemaining() int {
	if s.ExpiresAt == nil {
		return -1 // Never expires
	}
	remaining := time.Until(*s.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return int(remaining.Hours() / 24)
}

// SubscriptionResponse is the subscription data structure for API responses.
type SubscriptionResponse struct {
	ID             uint               `json:"id"`
	UserEmail      string             `json:"user_email"`
	Plan           PlanType           `json:"plan"`
	Status         SubscriptionStatus `json:"status"`
	DataLimitBytes int64              `json:"data_limit_bytes"`
	DataUsedBytes  int64              `json:"data_used_bytes"`
	StartDate      time.Time          `json:"start_date"`
	ExpiresAt      *time.Time         `json:"expires_at,omitempty"`
	DaysRemaining  int                `json:"days_remaining"`
	IsActive       bool               `json:"is_active"`
}

// ToResponse converts Subscription to a response structure.
func (s *Subscription) ToResponse() SubscriptionResponse {
	resp := SubscriptionResponse{
		ID:             s.ID,
		Plan:           s.Plan,
		Status:         s.Status,
		DataLimitBytes: s.DataLimitBytes,
		DataUsedBytes:  s.DataUsedBytes,
		StartDate:      s.StartDate,
		ExpiresAt:      s.ExpiresAt,
		DaysRemaining:  s.DaysRemaining(),
		IsActive:       s.IsActive(),
	}

	if s.User != nil {
		resp.UserEmail = s.User.Email
	}

	return resp
}

// GetPlanDetails returns data limit and duration for each plan type.
func GetPlanDetails(plan PlanType) (dataLimitGB int64, durationDays int) {
	switch plan {
	case PlanFree:
		return 5, 0 // 5GB, never expires but limited data
	case PlanMonthly:
		return 100, 30 // 100GB, 30 days
	case PlanYearly:
		return 0, 365 // Unlimited, 365 days
	default:
		return 5, 0
	}
}
