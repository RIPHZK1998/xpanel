// Package models contains all GORM database models.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserStatus represents the status of a user account.
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// UserRole represents the role of a user.
type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

// User represents a VPN service user.
type User struct {
	ID           uint           `gorm:"primaryKey;index:idx_user_lookup,priority:1" json:"id"`
	Email        string         `gorm:"type:varchar(255);unique;not null;index" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	UUID         string         `gorm:"type:varchar(36);unique;not null;index" json:"uuid"` // For xray-core
	Role         UserRole       `gorm:"type:varchar(20);default:'user';index" json:"role"`
	Status       UserStatus     `gorm:"type:varchar(20);default:'active';index" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index:idx_user_lookup,priority:2" json:"-"`

	// Relationships
	Subscription *UserSubscription `gorm:"foreignKey:UserID" json:"subscription,omitempty"`
	Devices      []Device          `gorm:"foreignKey:UserID" json:"devices,omitempty"`
	TrafficLogs  []TrafficLog      `gorm:"foreignKey:UserID" json:"-"`
}

// BeforeCreate generates a UUID for new users.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		u.UUID = uuid.New().String()
	}
	return nil
}

// IsAdmin checks if the user has admin role.
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// UserResponse is the safe user data structure for API responses.
type UserResponse struct {
	ID           uint                      `json:"id"`
	Email        string                    `json:"email"`
	UUID         string                    `json:"uuid"`
	Role         UserRole                  `json:"role"`
	Status       UserStatus                `json:"status"`
	CreatedAt    time.Time                 `json:"created_at"`
	Subscription *UserSubscriptionResponse `json:"subscription,omitempty"`
}

// ToResponse converts User to a safe response structure.
func (u *User) ToResponse() UserResponse {
	resp := UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		UUID:      u.UUID,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}

	if u.Subscription != nil {
		subResp := u.Subscription.ToResponse()
		resp.Subscription = &subResp
	}

	return resp
}
