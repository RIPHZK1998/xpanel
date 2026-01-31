package models

import (
	"time"

	"gorm.io/gorm"
)

// PlanDuration represents the duration type of a subscription plan.
type PlanDuration string

const (
	DurationWeekly    PlanDuration = "weekly"    // 7 days
	DurationMonthly   PlanDuration = "monthly"   // 30 days
	DurationQuarterly PlanDuration = "quarterly" // 90 days
	DurationAnnual    PlanDuration = "annual"    // 365 days
)

// PlanStatus represents the status of a subscription plan.
type PlanStatus string

const (
	PlanStatusActive   PlanStatus = "active"
	PlanStatusArchived PlanStatus = "archived"
)

// SubscriptionPlan represents a subscription plan template.
type SubscriptionPlan struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null;unique" json:"name"`
	DisplayName string         `gorm:"type:varchar(100);not null" json:"display_name"`
	Duration    PlanDuration   `gorm:"type:varchar(20);not null" json:"duration"`
	Price       float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	DataLimitGB int64          `gorm:"default:0" json:"data_limit_gb"` // 0 = unlimited
	MaxDevices  int            `gorm:"default:5" json:"max_devices"`
	Status      PlanStatus     `gorm:"type:varchar(20);default:'active'" json:"status"`
	Description string         `gorm:"type:text" json:"description"`
	Features    string         `gorm:"type:text" json:"features"` // JSON string
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Nodes             []Node             `gorm:"many2many:plan_nodes" json:"nodes,omitempty"`
	UserSubscriptions []UserSubscription `gorm:"foreignKey:PlanID" json:"-"`
}

// GetDurationDays returns the number of days for the plan duration.
func (p *SubscriptionPlan) GetDurationDays() int {
	switch p.Duration {
	case DurationWeekly:
		return 7
	case DurationMonthly:
		return 30
	case DurationQuarterly:
		return 90
	case DurationAnnual:
		return 365
	default:
		return 30
	}
}

// GetDataLimitBytes returns the data limit in bytes.
func (p *SubscriptionPlan) GetDataLimitBytes() int64 {
	if p.DataLimitGB == 0 {
		return 0 // Unlimited
	}
	return p.DataLimitGB * 1024 * 1024 * 1024
}

// IsActive checks if the plan is currently active.
func (p *SubscriptionPlan) IsActive() bool {
	return p.Status == PlanStatusActive
}

// SubscriptionPlanResponse is the API response structure for subscription plans.
type SubscriptionPlanResponse struct {
	ID             uint         `json:"id"`
	Name           string       `json:"name"`
	DisplayName    string       `json:"display_name"`
	Duration       PlanDuration `json:"duration"`
	DurationDays   int          `json:"duration_days"`
	Price          float64      `json:"price"`
	DataLimitGB    int64        `json:"data_limit_gb"`
	DataLimitBytes int64        `json:"data_limit_bytes"`
	MaxDevices     int          `json:"max_devices"`
	Status         PlanStatus   `json:"status"`
	Description    string       `json:"description"`
	Features       []string     `json:"features"`
	NodeCount      int          `json:"node_count,omitempty"`
	UserCount      int          `json:"user_count,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

// ToResponse converts SubscriptionPlan to a response structure.
func (p *SubscriptionPlan) ToResponse() SubscriptionPlanResponse {
	// Parse features from JSON string
	var features []string
	if p.Features != "" {
		// Simple comma-separated for now, can be enhanced to proper JSON later
		features = []string{p.Features}
	}

	return SubscriptionPlanResponse{
		ID:             p.ID,
		Name:           p.Name,
		DisplayName:    p.DisplayName,
		Duration:       p.Duration,
		DurationDays:   p.GetDurationDays(),
		Price:          p.Price,
		DataLimitGB:    p.DataLimitGB,
		DataLimitBytes: p.GetDataLimitBytes(),
		MaxDevices:     p.MaxDevices,
		Status:         p.Status,
		Description:    p.Description,
		Features:       features,
		NodeCount:      len(p.Nodes),
		UserCount:      len(p.UserSubscriptions),
		CreatedAt:      p.CreatedAt,
	}
}
