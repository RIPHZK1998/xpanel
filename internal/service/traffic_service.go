package service

import (
	"time"

	"xpanel/internal/models"
	"xpanel/internal/repository"
)

// TrafficService handles traffic-related business logic.
type TrafficService struct {
	trafficRepo *repository.TrafficRepository
	subRepo     *repository.SubscriptionRepository
}

// NewTrafficService creates a new traffic service.
func NewTrafficService(
	trafficRepo *repository.TrafficRepository,
	subRepo *repository.SubscriptionRepository,
) *TrafficService {
	return &TrafficService{
		trafficRepo: trafficRepo,
		subRepo:     subRepo,
	}
}

// RecordTraffic records traffic usage for a user on a node.
func (s *TrafficService) RecordTraffic(userID, nodeID uint, uploadBytes, downloadBytes int64) error {
	log := &models.TrafficLog{
		UserID:        userID,
		NodeID:        nodeID,
		UploadBytes:   uploadBytes,
		DownloadBytes: downloadBytes,
		RecordedAt:    time.Now(),
	}

	if err := s.trafficRepo.Create(log); err != nil {
		return err
	}

	// Update subscription data usage
	totalBytes := uploadBytes + downloadBytes
	return s.subRepo.UpdateDataUsage(userID, totalBytes)
}

// RecordTrafficBatch records multiple traffic logs at once.
func (s *TrafficService) RecordTrafficBatch(logs []models.TrafficLog) error {
	return s.trafficRepo.CreateBatch(logs)
}

// GetUserStats retrieves total traffic statistics for a user.
func (s *TrafficService) GetUserStats(userID uint) (*models.UserTrafficStats, error) {
	stats, err := s.trafficRepo.GetUserTotalTraffic(userID)
	if err != nil {
		return nil, err
	}

	// Get traffic by node
	byNode, err := s.trafficRepo.GetUserTrafficByNode(userID)
	if err != nil {
		return nil, err
	}
	stats.ByNode = byNode

	return stats, nil
}

// GetUserTrafficHistory retrieves traffic logs for a user within a time range.
func (s *TrafficService) GetUserTrafficHistory(userID uint, from, to time.Time) ([]models.TrafficLog, error) {
	return s.trafficRepo.GetUserTraffic(userID, from, to)
}

// GetNodeStats retrieves traffic statistics for a node.
func (s *TrafficService) GetNodeStats(nodeID uint) (*models.TrafficSummary, error) {
	return s.trafficRepo.GetNodeTraffic(nodeID)
}

// CleanupOldLogs removes traffic logs older than the specified days.
func (s *TrafficService) CleanupOldLogs(retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return s.trafficRepo.DeleteOldLogs(cutoff)
}
