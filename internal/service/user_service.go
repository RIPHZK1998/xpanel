package service

import (
	"xpanel/internal/models"
	"xpanel/internal/repository"
)

// UserService handles user-related business logic.
type UserService struct {
	userRepo            *repository.UserRepository
	deviceRepo          *repository.DeviceRepository
	subscriptionService *SubscriptionService
}

// NewUserService creates a new user service.
func NewUserService(
	userRepo *repository.UserRepository,
	deviceRepo *repository.DeviceRepository,
	subscriptionService *SubscriptionService,
) *UserService {
	return &UserService{
		userRepo:            userRepo,
		deviceRepo:          deviceRepo,
		subscriptionService: subscriptionService,
	}
}

// GetProfile retrieves a user's profile by ID.
func (s *UserService) GetProfile(userID uint) (*models.User, error) {
	return s.userRepo.GetByIDWithSubscription(userID)
}

// GetUserByUUID retrieves a user by their xray UUID.
func (s *UserService) GetUserByUUID(uuid string) (*models.User, error) {
	return s.userRepo.GetByUUID(uuid)
}

// UpdateStatus updates a user's account status.
func (s *UserService) UpdateStatus(userID uint, status models.UserStatus) error {
	return s.userRepo.UpdateStatus(userID, status)
}

// SuspendUser suspends a user account and their subscription.
func (s *UserService) SuspendUser(userID uint) error {
	// Deactivate all devices
	if err := s.deviceRepo.DeactivateAllForUser(userID); err != nil {
		return err
	}

	// Update user status
	if err := s.userRepo.UpdateStatus(userID, models.UserStatusSuspended); err != nil {
		return err
	}

	// Also suspend the user's subscription
	if err := s.subscriptionService.Suspend(userID); err != nil {
		// Log error but don't fail if subscription doesn't exist
		// This handles cases where user might not have a subscription yet
		return nil
	}

	return nil
}

// ActivateUser activates a suspended user account and their subscription.
func (s *UserService) ActivateUser(userID uint) error {
	// Update user status
	if err := s.userRepo.UpdateStatus(userID, models.UserStatusActive); err != nil {
		return err
	}

	// Also activate the user's subscription
	if err := s.subscriptionService.Activate(userID); err != nil {
		// Log error but don't fail if subscription doesn't exist
		return nil
	}

	return nil
}

// GetDevices retrieves all devices for a user.
func (s *UserService) GetDevices(userID uint) ([]models.Device, error) {
	return s.deviceRepo.GetByUserID(userID)
}

// GetActiveDevices retrieves active devices for a user.
func (s *UserService) GetActiveDevices(userID uint) ([]models.Device, error) {
	return s.deviceRepo.GetActiveByUserID(userID)
}

// RegisterDevice registers or updates a device for a user.
func (s *UserService) RegisterDevice(userID uint, ipAddress, deviceType, deviceName, userAgent string) (*models.Device, error) {
	return s.deviceRepo.FindOrCreate(userID, ipAddress, deviceType, deviceName, userAgent)
}

// DeactivateDevice deactivates a specific device.
func (s *UserService) DeactivateDevice(userID, deviceID uint) error {
	device, err := s.deviceRepo.GetByID(deviceID)
	if err != nil {
		return err
	}
	if device.UserID != userID {
		return repository.ErrDeviceNotFound
	}
	return s.deviceRepo.Deactivate(deviceID)
}

// DeactivateAllDevices deactivates all devices for a user.
func (s *UserService) DeactivateAllDevices(userID uint) error {
	return s.deviceRepo.DeactivateAllForUser(userID)
}

// GetActiveUsers retrieves all active users with their subscriptions.
func (s *UserService) GetActiveUsers() ([]models.User, error) {
	return s.userRepo.GetActiveUsersWithSubscription()
}
