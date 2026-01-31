package models

import (
	"time"
)

// UserActivity tracks user connection activity and online status.
type UserActivity struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex" json:"user_id"`
	NodeID    uint      `gorm:"index" json:"node_id"` // Last connected node
	LastSeen  time.Time `gorm:"index" json:"last_seen"`
	IsOnline  bool      `gorm:"index" json:"is_online"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Node *Node `gorm:"foreignKey:NodeID" json:"node,omitempty"`
}

// TableName specifies the table name for GORM.
func (UserActivity) TableName() string {
	return "user_activities"
}

// IsCurrentlyOnline checks if the user should be considered online.
// A user is considered online if their last activity was within 2 minutes.
func (ua *UserActivity) IsCurrentlyOnline() bool {
	return time.Since(ua.LastSeen) < 2*time.Minute
}

// UserActivityResponse is the safe activity data structure for API responses.
type UserActivityResponse struct {
	UserID      uint      `json:"user_id"`
	NodeID      uint      `json:"node_id"`
	NodeName    string    `json:"node_name,omitempty"`
	LastSeen    time.Time `json:"last_seen"`
	IsOnline    bool      `json:"is_online"`
	DeviceCount int       `json:"device_count"`
	DeviceIPs   []string  `json:"device_ips,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts UserActivity to a safe response structure.
func (ua *UserActivity) ToResponse() UserActivityResponse {
	resp := UserActivityResponse{
		UserID:    ua.UserID,
		NodeID:    ua.NodeID,
		LastSeen:  ua.LastSeen,
		IsOnline:  ua.IsCurrentlyOnline(), // Calculate real-time online status
		UpdatedAt: ua.UpdatedAt,
	}

	if ua.Node != nil {
		resp.NodeName = ua.Node.Name
	}

	return resp
}

// NodeActivityReport represents activity data received from a node agent.
type NodeActivityReport struct {
	NodeID    uint                     `json:"node_id"`
	Users     []NodeUserActivityReport `json:"users"`
	Timestamp time.Time                `json:"timestamp"`
}

// NodeUserActivityReport represents a single user's activity from a node.
type NodeUserActivityReport struct {
	Email       string    `json:"email"`
	LastSeen    time.Time `json:"last_seen"`
	IsOnline    bool      `json:"is_online"`
	DeviceCount int       `json:"device_count"`         // Number of unique devices (IPs)
	DeviceIPs   []string  `json:"device_ips,omitempty"` // List of connected IP addresses
}
