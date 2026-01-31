// Package models contains shared data structures.
package models

import "time"

// UserConfig represents a user configuration from the panel.
type UserConfig struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	UUID      string `json:"uuid"`
	Status    string `json:"status"`
	DataLimit int64  `json:"data_limit_bytes"`
	DataUsed  int64  `json:"data_used_bytes"`
}

// UserSyncResponse represents the response from panel user sync endpoint.
type UserSyncResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Users []UserConfig `json:"users"`
	} `json:"data"`
}

// HeartbeatRequest represents a heartbeat sent to the panel.
type HeartbeatRequest struct {
	NodeID           uint      `json:"node_id"`
	Status           string    `json:"status"`
	CurrentUsers     int       `json:"current_users"`
	CPUUsage         float64   `json:"cpu_usage"`
	MemoryUsage      float64   `json:"memory_usage"`
	Uptime           int64     `json:"uptime"`
	Timestamp        time.Time `json:"timestamp"`
	RealityPublicKey string    `json:"reality_public_key,omitempty"` // Node's auto-generated Reality public key
}

// TrafficReportRequest represents traffic data sent to the panel.
type TrafficReportRequest struct {
	NodeID    uint                `json:"node_id"`
	Traffic   []UserTrafficReport `json:"traffic"`
	Timestamp time.Time           `json:"timestamp"`
}

// UserTrafficReport represents traffic for a single user.
type UserTrafficReport struct {
	UserEmail     string `json:"user_email"`
	UploadBytes   int64  `json:"upload_bytes"`
	DownloadBytes int64  `json:"download_bytes"`
}

// XrayUser represents a user in xray-core.
type XrayUser struct {
	Email    string
	UUID     string
	Level    int
	Protocol string // vless, vmess, etc.
	Flow     string // xtls-rprx-vision for Reality
}

// XrayStats represents traffic statistics from xray-core.
type XrayStats struct {
	UploadBytes   int64
	DownloadBytes int64
}

// UserActivityReport represents activity status for a single user.
type UserActivityReport struct {
	Email       string    `json:"email"`
	LastSeen    time.Time `json:"last_seen"`
	IsOnline    bool      `json:"is_online"`
	DeviceCount int       `json:"device_count"`         // Number of unique IP addresses (devices)
	DeviceIPs   []string  `json:"device_ips,omitempty"` // List of IP addresses
}

// ActivityReportRequest represents activity data sent to the panel.
type ActivityReportRequest struct {
	NodeID    uint                 `json:"node_id"`
	Users     []UserActivityReport `json:"users"`
	Timestamp time.Time            `json:"timestamp"`
}
