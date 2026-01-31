package repository

import (
	"xpanel/internal/models"

	"gorm.io/gorm"
)

// SystemConfigRepository handles database operations for system configuration
type SystemConfigRepository struct {
	db *gorm.DB
}

// NewSystemConfigRepository creates a new system config repository
func NewSystemConfigRepository(db *gorm.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

// Get retrieves a configuration value by key
func (r *SystemConfigRepository) Get(key string) (*models.SystemConfig, error) {
	var config models.SystemConfig
	if err := r.db.Where("key = ?", key).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// GetAll retrieves all configuration entries
func (r *SystemConfigRepository) GetAll() ([]models.SystemConfig, error) {
	var configs []models.SystemConfig
	if err := r.db.Order("key ASC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// Set creates or updates a configuration entry
func (r *SystemConfigRepository) Set(config *models.SystemConfig) error {
	return r.db.Save(config).Error
}

// Delete removes a configuration entry
func (r *SystemConfigRepository) Delete(key string) error {
	return r.db.Where("key = ?", key).Delete(&models.SystemConfig{}).Error
}

// Exists checks if a configuration key exists
func (r *SystemConfigRepository) Exists(key string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.SystemConfig{}).Where("key = ?", key).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
