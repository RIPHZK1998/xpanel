// Package config handles configuration loading for the node agent.
package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete agent configuration.
type Config struct {
	Node      NodeConfig     `yaml:"node"`
	Panel     PanelConfig    `yaml:"panel"`
	Files     FilesConfig    `yaml:"files"`
	Xray      XrayConfig     `yaml:"xray"`
	Intervals IntervalConfig `yaml:"intervals"`
	Logging   LoggingConfig  `yaml:"logging"`
}

// NodeConfig contains node identification.
type NodeConfig struct {
	ID   uint   `yaml:"id"`
	Name string `yaml:"name"`
}

// PanelConfig contains management panel connection settings.
type PanelConfig struct {
	URL     string        `yaml:"url"`
	APIKey  string        `yaml:"api_key"`
	Timeout time.Duration `yaml:"timeout"`
}

// FilesConfig contains local file paths.
type FilesConfig struct {
	TLSCert        string `yaml:"tls_cert"`
	TLSKey         string `yaml:"tls_key"`
	GeoIP          string `yaml:"geoip"`
	GeoSite        string `yaml:"geosite"`
	XrayConfig     string `yaml:"xray_config"`
	RealityKeyPath string `yaml:"reality_key"` // Path to store auto-generated Reality private key
}

// XrayConfig contains xray-core API settings (local only).
// Protocol, port, TLS settings are fetched from panel.
type XrayConfig struct {
	APIAddress string `yaml:"api_address"`
	APIPort    int    `yaml:"api_port"`

	// These will be overridden by panel config
	Protocol      string `yaml:"-"` // Fetched from panel
	ListenAddress string `yaml:"-"` // Always 0.0.0.0
	ListenPort    int    `yaml:"-"` // Fetched from panel
	TLSEnabled    bool   `yaml:"-"` // Fetched from panel
	CertFile      string `yaml:"-"` // From files.tls_cert
	KeyFile       string `yaml:"-"` // From files.tls_key
	SNI           string `yaml:"-"` // Fetched from panel
	InboundTag    string `yaml:"-"` // Fetched from panel
	GeoIPPath     string `yaml:"-"` // From files.geoip
	GeoSitePath   string `yaml:"-"` // From files.geosite
	ConfigPath    string `yaml:"-"` // From files.xray_config

	// Reality settings (fetched from panel)
	RealityEnabled     bool   `yaml:"-"`
	RealityDest        string `yaml:"-"`
	RealityServerNames string `yaml:"-"`
	RealityPrivateKey  string `yaml:"-"`
	RealityPublicKey   string `yaml:"-"`
	RealityShortIds    string `yaml:"-"`
	RealityKeyPath     string `yaml:"-"` // Path to store auto-generated Reality private key
}

// IntervalConfig contains timing intervals for various operations.
type IntervalConfig struct {
	Heartbeat      time.Duration `yaml:"heartbeat"`
	UserSync       time.Duration `yaml:"user_sync"`
	TrafficReport  time.Duration `yaml:"traffic_report"`
	ActivityReport time.Duration `yaml:"activity_report"`
}

// LoggingConfig contains logging settings.
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load reads and parses the configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Panel.Timeout == 0 {
		cfg.Panel.Timeout = 10 * time.Second
	}
	if cfg.Intervals.Heartbeat == 0 {
		cfg.Intervals.Heartbeat = 30 * time.Second
	}
	if cfg.Intervals.UserSync == 0 {
		cfg.Intervals.UserSync = 5 * time.Minute
	}
	if cfg.Intervals.TrafficReport == 0 {
		cfg.Intervals.TrafficReport = 1 * time.Minute
	}
	if cfg.Intervals.ActivityReport == 0 {
		cfg.Intervals.ActivityReport = 30 * time.Second
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}

	// Map file paths to xray config
	cfg.Xray.CertFile = cfg.Files.TLSCert
	cfg.Xray.KeyFile = cfg.Files.TLSKey
	cfg.Xray.GeoIPPath = cfg.Files.GeoIP
	cfg.Xray.GeoSitePath = cfg.Files.GeoSite
	cfg.Xray.ConfigPath = cfg.Files.XrayConfig
	cfg.Xray.ListenAddress = "0.0.0.0"

	// Set Reality key path (default if not specified)
	if cfg.Files.RealityKeyPath != "" {
		cfg.Xray.RealityKeyPath = cfg.Files.RealityKeyPath
	} else {
		cfg.Xray.RealityKeyPath = "reality_private.key"
	}

	return &cfg, nil
}
