package models

import "time"

// NodeHeartbeat represents a heartbeat report from a node agent.
type NodeHeartbeat struct {
	NodeID           uint      `json:"node_id" binding:"required"`
	Status           string    `json:"status" binding:"required"` // online, offline, maintenance
	CurrentUsers     int       `json:"current_users"`
	CPUUsage         float64   `json:"cpu_usage"`
	MemoryUsage      float64   `json:"memory_usage"`
	Uptime           int64     `json:"uptime"` // seconds
	Timestamp        time.Time `json:"timestamp"`
	RealityPublicKey string    `json:"reality_public_key,omitempty"` // Node's auto-generated Reality public key
}

// NodeUserSync represents user data to sync to a node.
type NodeUserSync struct {
	Users []UserNodeConfig `json:"users"`
}

// UserNodeConfig represents user configuration for a node.
type UserNodeConfig struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	UUID      string `json:"uuid"`
	Status    string `json:"status"` // active, suspended
	DataLimit int64  `json:"data_limit_bytes"`
	DataUsed  int64  `json:"data_used_bytes"`
}

// NodeTrafficReport represents traffic data reported by a node.
type NodeTrafficReport struct {
	NodeID    uint                `json:"node_id" binding:"required"`
	Traffic   []UserTrafficReport `json:"traffic" binding:"required"`
	Timestamp time.Time           `json:"timestamp"`
}

// UserTrafficReport represents traffic for a single user.
type UserTrafficReport struct {
	UserEmail     string `json:"user_email" binding:"required"`
	UploadBytes   int64  `json:"upload_bytes"`
	DownloadBytes int64  `json:"download_bytes"`
}
