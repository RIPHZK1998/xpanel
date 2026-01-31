package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"xpanel/internal/models"
	"xpanel/internal/repository"
)

// SystemConfigService handles system configuration with encryption and caching
type SystemConfigService struct {
	repo          *repository.SystemConfigRepository
	encryptionKey []byte
	cache         map[string]*cachedConfig
	cacheMutex    sync.RWMutex
	cacheTTL      time.Duration
}

type cachedConfig struct {
	value     string
	expiresAt time.Time
}

// NewSystemConfigService creates a new system config service
func NewSystemConfigService(repo *repository.SystemConfigRepository, jwtSecret string) *SystemConfigService {
	// Derive encryption key from JWT secret using SHA-256
	hash := sha256.Sum256([]byte(jwtSecret))

	return &SystemConfigService{
		repo:          repo,
		encryptionKey: hash[:],
		cache:         make(map[string]*cachedConfig),
		cacheTTL:      5 * time.Minute, // 5 minute cache
	}
}

// GetConfig retrieves and decrypts a configuration value
func (s *SystemConfigService) GetConfig(key string) (string, error) {
	// Check cache first
	s.cacheMutex.RLock()
	if cached, ok := s.cache[key]; ok && time.Now().Before(cached.expiresAt) {
		s.cacheMutex.RUnlock()
		return cached.value, nil
	}
	s.cacheMutex.RUnlock()

	// Fetch from database
	config, err := s.repo.Get(key)
	if err != nil {
		return "", err
	}

	value := config.Value
	if config.Encrypted {
		decrypted, err := s.decrypt(value)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt config: %w", err)
		}
		value = decrypted
	}

	// Update cache
	s.cacheMutex.Lock()
	s.cache[key] = &cachedConfig{
		value:     value,
		expiresAt: time.Now().Add(s.cacheTTL),
	}
	s.cacheMutex.Unlock()

	return value, nil
}

// SetConfig stores and optionally encrypts a configuration value
func (s *SystemConfigService) SetConfig(key, value string, encrypted bool, description string, updatedBy *uint) error {
	storedValue := value
	if encrypted {
		encryptedValue, err := s.encrypt(value)
		if err != nil {
			return fmt.Errorf("failed to encrypt config: %w", err)
		}
		storedValue = encryptedValue
	}

	config := &models.SystemConfig{
		Key:         key,
		Value:       storedValue,
		Encrypted:   encrypted,
		Description: description,
		UpdatedBy:   updatedBy,
	}

	if err := s.repo.Set(config); err != nil {
		return err
	}

	// Invalidate cache
	s.cacheMutex.Lock()
	delete(s.cache, key)
	s.cacheMutex.Unlock()

	return nil
}

// GetAllConfigs retrieves all configuration entries
func (s *SystemConfigService) GetAllConfigs(maskEncrypted bool) ([]models.SystemConfigResponse, error) {
	configs, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := make([]models.SystemConfigResponse, len(configs))
	for i, config := range configs {
		responses[i] = config.ToResponse(maskEncrypted)
	}

	return responses, nil
}

// ReloadCache clears the cache, forcing fresh reads from database
func (s *SystemConfigService) ReloadCache() {
	s.cacheMutex.Lock()
	s.cache = make(map[string]*cachedConfig)
	s.cacheMutex.Unlock()
}

// GetNodeApiKey is a convenience method to get the node API key
func (s *SystemConfigService) GetNodeApiKey() (string, error) {
	return s.GetConfig("node_api_key")
}

// encrypt encrypts a string using AES-256-GCM
func (s *SystemConfigService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts a string using AES-256-GCM
func (s *SystemConfigService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
