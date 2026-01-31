package xray

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"xpanel-agent/config"

	"github.com/sirupsen/logrus"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"

	// Import these to register protocol handlers
	_ "github.com/xtls/xray-core/app/dispatcher"
	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	_ "github.com/xtls/xray-core/proxy/blackhole"
	_ "github.com/xtls/xray-core/proxy/dokodemo"
	_ "github.com/xtls/xray-core/proxy/freedom"
	_ "github.com/xtls/xray-core/proxy/vless/inbound"
	_ "github.com/xtls/xray-core/proxy/vless/outbound"
	_ "github.com/xtls/xray-core/proxy/vmess/inbound"
	_ "github.com/xtls/xray-core/proxy/vmess/outbound"
	_ "github.com/xtls/xray-core/transport/internet/reality"
	_ "github.com/xtls/xray-core/transport/internet/tcp"
)

// Manager handles xray-core lifecycle using embedded library.
type Manager struct {
	cfg      *config.XrayConfig
	logger   *logrus.Logger
	instance *core.Instance
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewManager creates a new xray manager.
func NewManager(cfg *config.XrayConfig, logger *logrus.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		cfg:    cfg,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the embedded xray-core instance.
func (m *Manager) Start() error {
	m.logger.Info("Starting embedded xray-core...")

	// Set asset directory environment variable for GeoIP/GeoSite
	if m.cfg.GeoIPPath != "" {
		assetDir := filepath.Dir(m.cfg.GeoIPPath)
		os.Setenv("XRAY_LOCATION_ASSET", assetDir)
		m.logger.Infof("Set XRAY_LOCATION_ASSET=%s", assetDir)
	}

	// Generate xray config JSON
	configJSON, err := m.generateConfigJSON()
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// Always save config to temp for debugging
	tempConfigPath := "/tmp/xray-config.json"
	if err := os.WriteFile(tempConfigPath, configJSON, 0644); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	m.logger.Infof("Config saved to: %s", tempConfigPath)

	// Also save to configured path if set
	if m.cfg.ConfigPath != "" {
		os.MkdirAll("/etc/xpanel-agent", 0755)
		if err := os.WriteFile(m.cfg.ConfigPath, configJSON, 0644); err != nil {
			m.logger.Warnf("Failed to save config file: %v", err)
		}
	}

	// Parse JSON config using xray's conf package
	pbConfig, err := serial.LoadJSONConfig(bytes.NewReader(configJSON))
	if err != nil {
		m.logger.Errorf("Failed to load config from JSON")
		m.logger.Errorf("Check /tmp/xray-config.json for the generated config")
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create xray instance
	instance, err := core.New(pbConfig)
	if err != nil {
		return fmt.Errorf("failed to create xray instance: %w", err)
	}

	m.instance = instance

	// Start xray
	if err := m.instance.Start(); err != nil {
		return fmt.Errorf("failed to start xray: %w", err)
	}

	m.logger.Info("Embedded xray-core started successfully")
	return nil
}

// Stop stops the embedded xray-core instance.
func (m *Manager) Stop() error {
	if m.instance == nil {
		return nil
	}

	m.logger.Info("Stopping embedded xray-core...")
	m.cancel()

	if err := m.instance.Close(); err != nil {
		return fmt.Errorf("failed to stop xray: %w", err)
	}

	m.logger.Info("Embedded xray-core stopped")
	return nil
}

// IsRunning checks if xray-core is running.
func (m *Manager) IsRunning() bool {
	return m.instance != nil
}

// generateConfigJSON generates the xray configuration as JSON bytes.
func (m *Manager) generateConfigJSON() ([]byte, error) {
	config := m.buildConfig()
	return json.MarshalIndent(config, "", "  ")
}

// buildConfig builds the xray configuration structure.
func (m *Manager) buildConfig() map[string]interface{} {
	config := map[string]interface{}{
		"log": map[string]interface{}{
			"loglevel": "warning",
		},
		"api": map[string]interface{}{
			"tag":      "api",
			"services": []string{"HandlerService", "StatsService"},
		},
		"stats": map[string]interface{}{},
		"policy": map[string]interface{}{
			"levels": map[string]interface{}{
				"0": map[string]interface{}{
					"statsUserUplink":   true,
					"statsUserDownlink": true,
				},
			},
			"system": map[string]interface{}{
				"statsInboundUplink":   true,
				"statsInboundDownlink": true,
			},
		},
		"inbounds":  m.buildInbounds(),
		"outbounds": m.buildOutbounds(),
		"routing": map[string]interface{}{
			"domainStrategy": "AsIs",
			"rules": []interface{}{
				map[string]interface{}{
					"type":        "field",
					"inboundTag":  []string{"api"},
					"outboundTag": "api",
				},
			},
		},
	}

	return config
}

// buildInbounds builds all inbound configurations.
func (m *Manager) buildInbounds() []interface{} {
	inbounds := []interface{}{
		// API inbound for stats
		map[string]interface{}{
			"tag":      "api",
			"listen":   m.cfg.APIAddress,
			"port":     m.cfg.APIPort,
			"protocol": "dokodemo-door",
			"settings": map[string]interface{}{
				"address": m.cfg.APIAddress,
			},
		},
	}

	// Add main proxy inbound
	proxyInbound := m.buildProxyInbound()
	if proxyInbound != nil {
		inbounds = append(inbounds, proxyInbound)
	}

	return inbounds
}

// buildProxyInbound builds the main proxy inbound.
func (m *Manager) buildProxyInbound() map[string]interface{} {
	inbound := map[string]interface{}{
		"tag":      m.cfg.InboundTag,
		"port":     m.cfg.ListenPort,
		"protocol": m.cfg.Protocol,
		"listen":   m.cfg.ListenAddress,
	}

	// Protocol-specific settings
	switch m.cfg.Protocol {
	case "vless":
		inbound["settings"] = map[string]interface{}{
			"clients":    []interface{}{},
			"decryption": "none",
		}
	case "vmess":
		inbound["settings"] = map[string]interface{}{
			"clients": []interface{}{},
		}
	default:
		m.logger.Warnf("Unsupported protocol: %s, falling back to vless", m.cfg.Protocol)
		inbound["settings"] = map[string]interface{}{
			"clients":    []interface{}{},
			"decryption": "none",
		}
	}

	// Add stream settings if needed
	if m.cfg.RealityEnabled {
		m.logger.Info("Configuring Reality protocol")
		inbound["streamSettings"] = m.buildRealityStream()
	} else if m.cfg.TLSEnabled {
		m.logger.Warn("TLS requested but disabled for now - use Reality instead")
		m.logger.Warn("Running without encryption (NOT secure!)")
	}

	return inbound
}

// buildRealityStream builds Reality stream settings.
func (m *Manager) buildRealityStream() map[string]interface{} {
	serverNames := strings.Split(m.cfg.RealityServerNames, ",")
	for i := range serverNames {
		serverNames[i] = strings.TrimSpace(serverNames[i])
	}

	shortIds := strings.Split(m.cfg.RealityShortIds, ",")
	for i := range shortIds {
		shortIds[i] = strings.TrimSpace(shortIds[i])
	}

	return map[string]interface{}{
		"network":  "tcp",
		"security": "reality",
		"realitySettings": map[string]interface{}{
			"show":        false,
			"dest":        m.cfg.RealityDest,
			"xver":        0,
			"serverNames": serverNames,
			"privateKey":  m.cfg.RealityPrivateKey,
			"shortIds":    shortIds,
		},
	}
}

// buildOutbounds builds all outbound configurations.
func (m *Manager) buildOutbounds() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"protocol": "freedom",
			"tag":      "direct",
		},
	}
}
