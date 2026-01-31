// Package agent contains the main agent logic.
package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"xpanel-agent/config"
	"xpanel-agent/internal/models"
	"xpanel-agent/internal/panel"
	"xpanel-agent/internal/xray"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sirupsen/logrus"
)

// Agent is the main node agent orchestrator.
type Agent struct {
	cfg            *config.Config
	logger         *logrus.Logger
	panelClient    *panel.Client
	xrayAPI        *xray.APIClient
	xrayManager    *xray.Manager
	geoDataManager *xray.GeoDataManager
	currentUsers   map[string]*models.UserConfig // email -> user
	activityCache  map[string]*UserActivityState // email -> activity state
	usersMutex     sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	realityKeys    *RealityKeys // Auto-generated Reality keypair
}

// NewAgent creates a new agent instance.
func NewAgent(cfg *config.Config, logger *logrus.Logger) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		cfg:           cfg,
		logger:        logger,
		panelClient:   panel.NewClient(cfg.Panel.URL, cfg.Panel.APIKey, cfg.Panel.Timeout),
		xrayAPI:       nil, // Will be initialized in Start() after fetching panel config
		xrayManager:   xray.NewManager(&cfg.Xray, logger),
		currentUsers:  make(map[string]*models.UserConfig),
		activityCache: make(map[string]*UserActivityState),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the agent and all background tasks.
func (a *Agent) Start() error {
	a.logger.Info("Starting xpanel-agent...")

	// Initialize Reality keys (generate if not exists)
	if err := a.initRealityKeys(); err != nil {
		a.logger.Warnf("Failed to initialize Reality keys: %v", err)
	}

	// Fetch node configuration from panel
	a.logger.Info("Fetching node configuration from panel...")
	nodeConfig, err := a.panelClient.FetchNodeConfig(a.cfg.Node.ID)
	if err != nil {
		a.logger.Warnf("Failed to fetch node config from panel: %v", err)
		a.logger.Info("Using local configuration from config.yaml")

		// Set default inbound tag if not from panel
		if a.cfg.Xray.InboundTag == "" {
			a.cfg.Xray.InboundTag = "inbound-vless"
		}
	} else {
		// Update xray config with panel settings
		a.logger.Infof("Using panel configuration: protocol=%s, port=%d, tls=%v, reality=%v",
			nodeConfig.Protocol, nodeConfig.Port, nodeConfig.TLSEnabled, nodeConfig.RealityEnabled)

		a.cfg.Xray.Protocol = nodeConfig.Protocol
		a.cfg.Xray.ListenPort = nodeConfig.Port
		a.cfg.Xray.TLSEnabled = nodeConfig.TLSEnabled
		a.cfg.Xray.SNI = nodeConfig.SNI
		a.cfg.Xray.InboundTag = nodeConfig.InboundTag

		// Ensure InboundTag is not empty
		if a.cfg.Xray.InboundTag == "" {
			a.cfg.Xray.InboundTag = "proxy"
			a.logger.Warn("Panel returned empty InboundTag, using default 'proxy'")
		}

		// Apply Reality settings if enabled
		if nodeConfig.RealityEnabled {
			// Validate required Reality fields
			if nodeConfig.RealityDest == "" {
				return fmt.Errorf("Reality protocol is enabled but Target Destination is not configured in the panel. Please configure it in the admin panel")
			}
			if nodeConfig.RealityServerNames == "" {
				return fmt.Errorf("Reality protocol is enabled but Server Names is not configured in the panel. Please configure it in the admin panel")
			}
			if nodeConfig.RealityShortIds == "" {
				return fmt.Errorf("Reality protocol is enabled but Short IDs is not configured in the panel. Please configure it in the admin panel")
			}

			a.logger.WithFields(logrus.Fields{
				"dest":         nodeConfig.RealityDest,
				"server_names": nodeConfig.RealityServerNames,
				"short_ids":    nodeConfig.RealityShortIds,
			}).Info("Reality protocol enabled with panel settings")

			a.cfg.Xray.RealityEnabled = true
			a.cfg.Xray.RealityDest = nodeConfig.RealityDest
			a.cfg.Xray.RealityServerNames = nodeConfig.RealityServerNames
			// Use locally-generated private key instead of panel setting
			if a.realityKeys != nil {
				a.cfg.Xray.RealityPrivateKey = a.realityKeys.PrivateKey
				a.cfg.Xray.RealityPublicKey = a.realityKeys.PublicKey
				a.logger.Info("Using locally-generated Reality keypair")
			} else {
				// Fallback to panel keys if local generation failed
				a.cfg.Xray.RealityPrivateKey = nodeConfig.RealityPrivateKey
				a.cfg.Xray.RealityPublicKey = nodeConfig.RealityPublicKey
				a.logger.Warn("Using panel-provided Reality keys (local generation failed)")
			}
			a.cfg.Xray.RealityShortIds = nodeConfig.RealityShortIds
		} else {
			a.logger.Info("Reality protocol disabled in panel configuration")
			a.cfg.Xray.RealityEnabled = false
		}

		// Update API settings if provided
		if nodeConfig.APIPort > 0 {
			a.cfg.Xray.APIPort = nodeConfig.APIPort
		}
	}

	// Initialize XrayAPI client AFTER we have the InboundTag from panel
	a.xrayAPI = xray.NewAPIClient(a.cfg.Xray.APIAddress, a.cfg.Xray.APIPort, a.cfg.Xray.InboundTag)
	a.logger.Infof("XrayAPI initialized with inbound tag: %s", a.cfg.Xray.InboundTag)

	// Initialize GeoData manager and ensure data files exist
	if a.cfg.Xray.GeoIPPath != "" && a.cfg.Xray.GeoSitePath != "" {
		a.geoDataManager = xray.NewGeoDataManager(a.cfg.Xray.GeoIPPath, a.cfg.Xray.GeoSitePath, a.logger)
		if err := a.geoDataManager.EnsureDataFiles(); err != nil {
			a.logger.Warnf("Failed to ensure GeoIP/GeoSite files: %v", err)
		} else {
			// Start auto-update in background
			a.geoDataManager.StartAutoUpdate(a.ctx.Done())
		}
	}

	// Start xray-core
	if err := a.xrayManager.Start(); err != nil {
		return err
	}

	// Wait for xray to be ready
	time.Sleep(2 * time.Second)

	// Initial user sync
	if err := a.syncUsers(); err != nil {
		a.logger.Warnf("Initial user sync failed: %v", err)
	}

	// Start background tasks
	go a.heartbeatLoop()
	go a.userSyncLoop()
	go a.trafficReportLoop()
	go a.activityLoop()

	a.logger.Info("xpanel-agent started successfully")
	return nil
}

// Stop stops the agent and all background tasks.
func (a *Agent) Stop() error {
	a.logger.Info("Stopping xpanel-agent...")

	// Cancel all background tasks
	a.cancel()

	// Stop xray-core
	if err := a.xrayManager.Stop(); err != nil {
		a.logger.Errorf("Failed to stop xray-core: %v", err)
	}

	a.logger.Info("xpanel-agent stopped")
	return nil
}

// heartbeatLoop sends periodic heartbeats to the panel.
func (a *Agent) heartbeatLoop() {
	ticker := time.NewTicker(a.cfg.Intervals.Heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if err := a.sendHeartbeat(); err != nil {
				a.logger.Errorf("Failed to send heartbeat: %v", err)
			}
		}
	}
}

// sendHeartbeat sends a heartbeat to the panel.
func (a *Agent) sendHeartbeat() error {
	// Get system metrics
	cpuPercent, _ := cpu.Percent(time.Second, false)
	memInfo, _ := mem.VirtualMemory()
	uptime, _ := host.Uptime()

	var cpuUsage float64
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	var memUsage float64
	if memInfo != nil {
		memUsage = memInfo.UsedPercent
	}

	// Get current user count
	a.usersMutex.RLock()
	userCount := len(a.currentUsers)
	a.usersMutex.RUnlock()

	heartbeat := &models.HeartbeatRequest{
		NodeID:       a.cfg.Node.ID,
		Status:       "online",
		CurrentUsers: userCount,
		CPUUsage:     cpuUsage,
		MemoryUsage:  memUsage,
		Uptime:       int64(uptime),
		Timestamp:    time.Now(),
	}

	// Include Reality public key if available
	if a.realityKeys != nil {
		heartbeat.RealityPublicKey = a.realityKeys.PublicKey
	}

	if err := a.panelClient.SendHeartbeat(heartbeat); err != nil {
		return err
	}

	a.logger.WithFields(logrus.Fields{
		"users":  userCount,
		"cpu":    cpuUsage,
		"memory": memUsage,
	}).Debug("Heartbeat sent")

	return nil
}

// userSyncLoop periodically syncs users from the panel.
func (a *Agent) userSyncLoop() {
	ticker := time.NewTicker(a.cfg.Intervals.UserSync)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if err := a.syncUsers(); err != nil {
				a.logger.Errorf("Failed to sync users: %v", err)
			}
		}
	}
}

// trafficReportLoop periodically reports traffic to the panel.
func (a *Agent) trafficReportLoop() {
	ticker := time.NewTicker(a.cfg.Intervals.TrafficReport)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if err := a.reportTraffic(); err != nil {
				a.logger.Errorf("Failed to report traffic: %v", err)
			}
		}
	}
}

// initRealityKeys loads or generates Reality keypair for this node.
func (a *Agent) initRealityKeys() error {
	keyPath := a.cfg.Xray.RealityKeyPath
	if keyPath == "" {
		keyPath = "reality_private.key"
	}

	keys, isNew, err := LoadOrCreateRealityKeys(keyPath)
	if err != nil {
		return err
	}

	a.realityKeys = keys

	if isNew {
		a.logger.WithFields(logrus.Fields{
			"key_path":   keyPath,
			"public_key": keys.PublicKey,
		}).Info("Generated new Reality keypair")
	} else {
		a.logger.WithFields(logrus.Fields{
			"key_path":   keyPath,
			"public_key": keys.PublicKey,
		}).Info("Loaded existing Reality keypair")
	}

	return nil
}
