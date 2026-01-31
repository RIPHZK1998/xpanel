package models

import (
	"time"

	"gorm.io/gorm"
)

// TrafficLog represents traffic usage data per user per node.
type TrafficLog struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"index;not null" json:"user_id"`
	NodeID        uint           `gorm:"index;not null" json:"node_id"`
	UploadBytes   int64          `gorm:"default:0" json:"upload_bytes"`
	DownloadBytes int64          `gorm:"default:0" json:"download_bytes"`
	RecordedAt    time.Time      `gorm:"index" json:"recorded_at"`
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"-"`
	Node *Node `gorm:"foreignKey:NodeID" json:"-"`
}

// TotalBytes returns total traffic in bytes.
func (t *TrafficLog) TotalBytes() int64 {
	return t.UploadBytes + t.DownloadBytes
}

// TrafficSummary represents aggregated traffic data.
type TrafficSummary struct {
	UserID        uint   `json:"user_id"`
	NodeID        uint   `json:"node_id,omitempty"`
	NodeName      string `json:"node_name,omitempty"`
	UploadBytes   int64  `json:"upload_bytes"`
	DownloadBytes int64  `json:"download_bytes"`
	TotalBytes    int64  `json:"total_bytes"`
}

// UserTrafficStats represents a user's overall traffic statistics.
type UserTrafficStats struct {
	TotalUpload   int64            `json:"total_upload"`
	TotalDownload int64            `json:"total_download"`
	TotalUsage    int64            `json:"total_usage"`
	ByNode        []TrafficSummary `json:"by_node,omitempty"`
}
