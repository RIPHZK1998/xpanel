package xray

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"xpanel/internal/models"
)

// Manager manages multiple xray-core nodes and user provisioning.
type Manager struct {
	nodes   map[uint]*Client
	nodesRW sync.RWMutex
}

// NewManager creates a new xray node manager.
func NewManager() *Manager {
	return &Manager{
		nodes: make(map[uint]*Client),
	}
}

// RegisterNode registers a node with the manager.
func (m *Manager) RegisterNode(node *models.Node) {
	m.nodesRW.Lock()
	defer m.nodesRW.Unlock()

	apiEndpoint := node.APIEndpoint
	if apiEndpoint == "" {
		apiEndpoint = node.Address
	}

	client := NewClient(apiEndpoint, node.APIPort, node.InboundTag)
	m.nodes[node.ID] = client
}

// UnregisterNode removes a node from the manager.
func (m *Manager) UnregisterNode(nodeID uint) {
	m.nodesRW.Lock()
	defer m.nodesRW.Unlock()
	delete(m.nodes, nodeID)
}

// GetClient retrieves the client for a specific node.
func (m *Manager) GetClient(nodeID uint) (*Client, error) {
	m.nodesRW.RLock()
	defer m.nodesRW.RUnlock()

	client, exists := m.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %d not registered", nodeID)
	}
	return client, nil
}

// ProvisionUser adds a user to a specific node.
func (m *Manager) ProvisionUser(nodeID uint, user *models.User, node *models.Node) error {
	client, err := m.GetClient(nodeID)
	if err != nil {
		return err
	}

	userConfig := &UserConfig{
		UUID:  user.UUID,
		Email: user.Email,
		Level: 0,
	}

	// Configure based on protocol
	switch node.Protocol {
	case models.ProtocolVLESS:
		userConfig.Flow = "xtls-rprx-vision"
	case models.ProtocolVMess:
		userConfig.AlterId = 0
	case models.ProtocolTrojan:
		userConfig.Password = user.UUID // Use UUID as password for Trojan
	}

	return client.AddUser(userConfig)
}

// DeprovisionUser removes a user from a specific node.
func (m *Manager) DeprovisionUser(nodeID uint, email string) error {
	client, err := m.GetClient(nodeID)
	if err != nil {
		return err
	}
	return client.RemoveUser(email)
}

// ProvisionUserToAllNodes provisions a user to all available nodes.
func (m *Manager) ProvisionUserToAllNodes(user *models.User, nodes []models.Node) error {
	m.nodesRW.RLock()
	defer m.nodesRW.RUnlock()

	var lastErr error
	for _, node := range nodes {
		if _, exists := m.nodes[node.ID]; !exists {
			continue
		}
		if err := m.ProvisionUser(node.ID, user, &node); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// DeprovisionUserFromAllNodes removes a user from all nodes.
func (m *Manager) DeprovisionUserFromAllNodes(email string) error {
	m.nodesRW.RLock()
	defer m.nodesRW.RUnlock()

	var lastErr error
	for nodeID := range m.nodes {
		if err := m.DeprovisionUser(nodeID, email); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// GetUserStats retrieves traffic statistics for a user from a specific node.
func (m *Manager) GetUserStats(nodeID uint, email string) (*UserStats, error) {
	client, err := m.GetClient(nodeID)
	if err != nil {
		return nil, err
	}
	return client.GetUserStats(email)
}

// HealthCheckNode performs a health check on a specific node.
func (m *Manager) HealthCheckNode(nodeID uint) error {
	client, err := m.GetClient(nodeID)
	if err != nil {
		return err
	}
	return client.HealthCheck()
}

// GenerateClientConfig generates a client configuration for a user and node.
func (m *Manager) GenerateClientConfig(user *models.User, node *models.Node) (*ClientConfig, error) {
	config := &ClientConfig{
		Protocol: string(node.Protocol),
		Address:  node.Address,
		Port:     node.Port,
		UUID:     user.UUID,
	}

	// Configure TLS
	if node.TLSEnabled {
		config.TLS = TLSConfig{
			Enabled:    true,
			ServerName: node.SNI,
			ALPN:       []string{"h2", "http/1.1"},
		}
	}

	// Configure transport (default to TCP)
	config.Transport = TransportConfig{
		Type: "tcp",
	}

	// Generate share link
	shareLink, err := m.generateShareLink(user, node)
	if err == nil {
		config.ShareLink = shareLink
	}

	// Protocol-specific configuration
	switch node.Protocol {
	case models.ProtocolVLESS:
		config.Flow = "xtls-rprx-vision"
	case models.ProtocolTrojan:
		config.Password = user.UUID
	}

	return config, nil
}

// generateShareLink creates a shareable link for the VPN configuration.
func (m *Manager) generateShareLink(user *models.User, node *models.Node) (string, error) {
	switch node.Protocol {
	case models.ProtocolVLESS:
		return m.generateVLESSLink(user, node), nil
	case models.ProtocolVMess:
		return m.generateVMessLink(user, node)
	case models.ProtocolTrojan:
		return m.generateTrojanLink(user, node), nil
	default:
		return "", fmt.Errorf("unsupported protocol: %s", node.Protocol)
	}
}

// generateVLESSLink generates a VLESS share link.
func (m *Manager) generateVLESSLink(user *models.User, node *models.Node) string {
	security := "none"
	if node.TLSEnabled {
		security = "tls"
	}

	link := fmt.Sprintf("vless://%s@%s:%d?type=tcp&security=%s",
		user.UUID, node.Address, node.Port, security)

	if node.TLSEnabled && node.SNI != "" {
		link += "&sni=" + node.SNI
	}

	link += "&flow=xtls-rprx-vision#" + node.Name

	return link
}

// generateVMessLink generates a VMess share link.
func (m *Manager) generateVMessLink(user *models.User, node *models.Node) (string, error) {
	config := map[string]interface{}{
		"v":    "2",
		"ps":   node.Name,
		"add":  node.Address,
		"port": node.Port,
		"id":   user.UUID,
		"aid":  0,
		"net":  "tcp",
		"type": "none",
		"host": "",
		"path": "",
		"tls":  "",
	}

	if node.TLSEnabled {
		config["tls"] = "tls"
		config["sni"] = node.SNI
	}

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	return "vmess://" + encoded, nil
}

// generateTrojanLink generates a Trojan share link.
func (m *Manager) generateTrojanLink(user *models.User, node *models.Node) string {
	link := fmt.Sprintf("trojan://%s@%s:%d", user.UUID, node.Address, node.Port)

	if node.TLSEnabled && node.SNI != "" {
		link += "?sni=" + node.SNI
	}

	link += "#" + node.Name

	return link
}
