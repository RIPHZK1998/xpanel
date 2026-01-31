package repository

import (
	"errors"
	"time"

	"xpanel/internal/models"

	"gorm.io/gorm"
)

// Device errors
var (
	ErrDeviceNotFound = errors.New("device not found")
)

// DeviceRepository handles device database operations.
type DeviceRepository struct {
	db *gorm.DB
}

// NewDeviceRepository creates a new device repository instance.
func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// Create inserts a new device record.
func (r *DeviceRepository) Create(device *models.Device) error {
	return r.db.Create(device).Error
}

// GetByID retrieves a device by its ID.
func (r *DeviceRepository) GetByID(id uint) (*models.Device, error) {
	var device models.Device
	result := r.db.First(&device, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrDeviceNotFound
		}
		return nil, result.Error
	}
	return &device, nil
}

// GetByUserID retrieves all devices for a user.
func (r *DeviceRepository) GetByUserID(userID uint) ([]models.Device, error) {
	var devices []models.Device
	result := r.db.Where("user_id = ?", userID).Order("last_active_at DESC").Find(&devices)
	if result.Error != nil {
		return nil, result.Error
	}
	return devices, nil
}

// GetActiveByUserID retrieves active devices for a user.
func (r *DeviceRepository) GetActiveByUserID(userID uint) ([]models.Device, error) {
	var devices []models.Device
	result := r.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_active_at DESC").Find(&devices)
	if result.Error != nil {
		return nil, result.Error
	}
	return devices, nil
}

// Update updates device fields.
func (r *DeviceRepository) Update(device *models.Device) error {
	return r.db.Save(device).Error
}

// UpdateLastActive updates the last active timestamp for a device.
func (r *DeviceRepository) UpdateLastActive(id uint) error {
	return r.db.Model(&models.Device{}).Where("id = ?", id).
		Update("last_active_at", time.Now()).Error
}

// Deactivate marks a device as inactive.
func (r *DeviceRepository) Deactivate(id uint) error {
	return r.db.Model(&models.Device{}).Where("id = ?", id).Update("is_active", false).Error
}

// DeactivateAllForUser deactivates all devices for a user.
func (r *DeviceRepository) DeactivateAllForUser(userID uint) error {
	return r.db.Model(&models.Device{}).Where("user_id = ?", userID).Update("is_active", false).Error
}

// Delete removes a device record.
func (r *DeviceRepository) Delete(id uint) error {
	return r.db.Delete(&models.Device{}, id).Error
}

// CountActiveByUserID counts active devices for a user.
func (r *DeviceRepository) CountActiveByUserID(userID uint) (int64, error) {
	var count int64
	result := r.db.Model(&models.Device{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&count)
	return count, result.Error
}

// FindOrCreate finds an existing device by IP or creates a new one.
func (r *DeviceRepository) FindOrCreate(userID uint, ipAddress, deviceType, deviceName, userAgent string) (*models.Device, error) {
	var device models.Device
	result := r.db.Where("user_id = ? AND ip_address = ?", userID, ipAddress).First(&device)

	if result.Error == nil {
		// Update existing device
		device.LastActiveAt = time.Now()
		device.IsActive = true
		device.UserAgent = userAgent
		if deviceName != "" {
			device.DeviceName = deviceName
		}
		if deviceType != "" {
			device.DeviceType = deviceType
		}
		if err := r.db.Save(&device).Error; err != nil {
			return nil, err
		}
		return &device, nil
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create new device
		device = models.Device{
			UserID:       userID,
			DeviceName:   deviceName,
			DeviceType:   deviceType,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
			LastActiveAt: time.Now(),
			IsActive:     true,
		}
		if err := r.db.Create(&device).Error; err != nil {
			return nil, err
		}
		return &device, nil
	}

	return nil, result.Error
}
