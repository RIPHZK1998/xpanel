package xray

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// Default download URLs for GeoIP/GeoSite data files
	GeoIPURL   = "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geoip.dat"
	GeoSiteURL = "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geosite.dat"

	// Update interval - check for updates every 24 hours
	GeoDataUpdateInterval = 24 * time.Hour
)

// GeoDataManager handles GeoIP/GeoSite data file management.
type GeoDataManager struct {
	geoIPPath   string
	geoSitePath string
	logger      *logrus.Logger
}

// NewGeoDataManager creates a new GeoData manager.
func NewGeoDataManager(geoIPPath, geoSitePath string, logger *logrus.Logger) *GeoDataManager {
	return &GeoDataManager{
		geoIPPath:   geoIPPath,
		geoSitePath: geoSitePath,
		logger:      logger,
	}
}

// EnsureDataFiles checks if GeoIP/GeoSite files exist and downloads them if missing.
func (m *GeoDataManager) EnsureDataFiles() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(m.geoIPPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create geo data directory: %w", err)
	}

	// Check and download GeoIP
	if err := m.ensureFile(m.geoIPPath, GeoIPURL, "geoip.dat"); err != nil {
		return err
	}

	// Check and download GeoSite
	if err := m.ensureFile(m.geoSitePath, GeoSiteURL, "geosite.dat"); err != nil {
		return err
	}

	return nil
}

// ensureFile checks if a file exists and downloads it if missing or outdated.
func (m *GeoDataManager) ensureFile(filePath, downloadURL, name string) error {
	fileInfo, err := os.Stat(filePath)

	// File exists - check if it needs update
	if err == nil {
		age := time.Since(fileInfo.ModTime())
		if age < GeoDataUpdateInterval {
			m.logger.Infof("Using existing %s (age: %v)", name, age.Round(time.Hour))
			return nil
		}
		m.logger.Infof("Updating %s (age: %v exceeds %v)", name, age.Round(time.Hour), GeoDataUpdateInterval)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat %s: %w", name, err)
	} else {
		m.logger.Infof("Downloading %s (file not found)", name)
	}

	// Download the file
	return m.downloadFile(filePath, downloadURL, name)
}

// downloadFile downloads a file from URL to the specified path.
func (m *GeoDataManager) downloadFile(filePath, downloadURL, name string) error {
	m.logger.Infof("Downloading %s from %s...", name, downloadURL)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute, // Large files may take time
	}

	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: HTTP %d", name, resp.StatusCode)
	}

	// Create temporary file first (atomic write)
	tmpFile := filePath + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file for %s: %w", name, err)
	}

	// Copy data
	written, err := io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to write %s: %w", name, err)
	}

	// Rename temp file to final path (atomic)
	if err := os.Rename(tmpFile, filePath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to finalize %s: %w", name, err)
	}

	m.logger.Infof("Downloaded %s successfully (%d bytes)", name, written)
	return nil
}

// GetAssetDir returns the directory containing the geo data files.
func (m *GeoDataManager) GetAssetDir() string {
	return filepath.Dir(m.geoIPPath)
}

// StartAutoUpdate starts a background goroutine to periodically update geo data files.
func (m *GeoDataManager) StartAutoUpdate(stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(GeoDataUpdateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				m.logger.Info("Stopping geo data auto-update")
				return
			case <-ticker.C:
				m.logger.Info("Checking for geo data updates...")
				if err := m.EnsureDataFiles(); err != nil {
					m.logger.Errorf("Failed to update geo data: %v", err)
				}
			}
		}
	}()
}
