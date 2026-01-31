package models

import (
	"time"

	"gorm.io/gorm"
)

// Device represents a user's connected device/session.
type Device struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	DeviceName   string         `gorm:"type:varchar(100)" json:"device_name"`
	DeviceType   string         `gorm:"type:varchar(50)" json:"device_type"` // ios, android, windows, macos, linux
	IPAddress    string         `gorm:"type:varchar(45)" json:"ip_address"`  // Supports IPv6
	UserAgent    string         `gorm:"type:varchar(500)" json:"user_agent,omitempty"`
	LastActiveAt time.Time      `json:"last_active_at"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

// DeviceResponse is the device data structure for API responses.
type DeviceResponse struct {
	ID           uint      `json:"id"`
	DeviceName   string    `json:"device_name"`
	DeviceType   string    `json:"device_type"`
	IPAddress    string    `json:"ip_address"`
	LastActiveAt time.Time `json:"last_active_at"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// ToResponse converts Device to a response structure.
func (d *Device) ToResponse() DeviceResponse {
	return DeviceResponse{
		ID:           d.ID,
		DeviceName:   d.DeviceName,
		DeviceType:   d.DeviceType,
		IPAddress:    d.IPAddress,
		LastActiveAt: d.LastActiveAt,
		IsActive:     d.IsActive,
		CreatedAt:    d.CreatedAt,
	}
}
