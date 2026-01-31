package repository

import (
	"time"

	"xpanel/internal/models"

	"gorm.io/gorm"
)

// TrafficRepository handles traffic log database operations.
type TrafficRepository struct {
	db *gorm.DB
}

// NewTrafficRepository creates a new traffic repository instance.
func NewTrafficRepository(db *gorm.DB) *TrafficRepository {
	return &TrafficRepository{db: db}
}

// Create inserts a new traffic log record.
func (r *TrafficRepository) Create(log *models.TrafficLog) error {
	return r.db.Create(log).Error
}

// CreateBatch inserts multiple traffic log records.
func (r *TrafficRepository) CreateBatch(logs []models.TrafficLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.CreateInBatches(logs, 100).Error
}

// GetUserTraffic retrieves traffic logs for a user within a time range.
func (r *TrafficRepository) GetUserTraffic(userID uint, from, to time.Time) ([]models.TrafficLog, error) {
	var logs []models.TrafficLog
	result := r.db.Where("user_id = ? AND recorded_at BETWEEN ? AND ?", userID, from, to).
		Order("recorded_at DESC").Find(&logs)
	if result.Error != nil {
		return nil, result.Error
	}
	return logs, nil
}

// GetUserTotalTraffic calculates total traffic for a user.
func (r *TrafficRepository) GetUserTotalTraffic(userID uint) (*models.UserTrafficStats, error) {
	var result struct {
		TotalUpload   int64
		TotalDownload int64
	}

	err := r.db.Model(&models.TrafficLog{}).
		Select("COALESCE(SUM(upload_bytes), 0) as total_upload, COALESCE(SUM(download_bytes), 0) as total_download").
		Where("user_id = ?", userID).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &models.UserTrafficStats{
		TotalUpload:   result.TotalUpload,
		TotalDownload: result.TotalDownload,
		TotalUsage:    result.TotalUpload + result.TotalDownload,
	}, nil
}

// GetUserTrafficByNode calculates traffic grouped by node for a user.
func (r *TrafficRepository) GetUserTrafficByNode(userID uint) ([]models.TrafficSummary, error) {
	var summaries []models.TrafficSummary

	err := r.db.Model(&models.TrafficLog{}).
		Select("user_id, node_id, nodes.name as node_name, "+
			"COALESCE(SUM(upload_bytes), 0) as upload_bytes, "+
			"COALESCE(SUM(download_bytes), 0) as download_bytes, "+
			"COALESCE(SUM(upload_bytes + download_bytes), 0) as total_bytes").
		Joins("LEFT JOIN nodes ON nodes.id = traffic_logs.node_id").
		Where("user_id = ?", userID).
		Group("user_id, node_id, nodes.name").
		Scan(&summaries).Error
	if err != nil {
		return nil, err
	}

	return summaries, nil
}

// GetNodeTraffic calculates total traffic for a node.
func (r *TrafficRepository) GetNodeTraffic(nodeID uint) (*models.TrafficSummary, error) {
	var summary models.TrafficSummary

	err := r.db.Model(&models.TrafficLog{}).
		Select("node_id, "+
			"COALESCE(SUM(upload_bytes), 0) as upload_bytes, "+
			"COALESCE(SUM(download_bytes), 0) as download_bytes, "+
			"COALESCE(SUM(upload_bytes + download_bytes), 0) as total_bytes").
		Where("node_id = ?", nodeID).
		Group("node_id").
		Scan(&summary).Error
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

// GetTrafficSince retrieves all traffic logs since a specific time.
func (r *TrafficRepository) GetTrafficSince(since time.Time) ([]models.TrafficLog, error) {
	var logs []models.TrafficLog
	result := r.db.Where("recorded_at > ?", since).Find(&logs)
	if result.Error != nil {
		return nil, result.Error
	}
	return logs, nil
}

// DeleteOldLogs removes traffic logs older than the specified duration.
func (r *TrafficRepository) DeleteOldLogs(olderThan time.Time) (int64, error) {
	result := r.db.Where("recorded_at < ?", olderThan).Delete(&models.TrafficLog{})
	return result.RowsAffected, result.Error
}
