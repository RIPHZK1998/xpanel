package models

import (
	"time"

	"gorm.io/gorm"
)

// NodeStatus represents the operational status of a VPN node.
type NodeStatus string

const (
	NodeStatusOnline      NodeStatus = "online"
	NodeStatusOffline     NodeStatus = "offline"
	NodeStatusMaintenance NodeStatus = "maintenance"
)

// ProtocolType represents supported VPN protocols.
type ProtocolType string

const (
	ProtocolVLESS  ProtocolType = "vless"
	ProtocolVMess  ProtocolType = "vmess"
	ProtocolTrojan ProtocolType = "trojan"
)

// Node represents a VPN server node running xray-core.
type Node struct {
	ID            uint         `gorm:"primaryKey;index:idx_node_lookup,priority:1" json:"id"`
	Name          string       `gorm:"type:varchar(100);not null;index" json:"name"`
	Address       string       `gorm:"type:varchar(255);not null" json:"address"`       // Server IP or hostname
	Port          int          `gorm:"not null" json:"port"`                            // VPN port
	Protocol      ProtocolType `gorm:"type:varchar(20);not null" json:"protocol"`       // vless/vmess/trojan
	APIEndpoint   string       `gorm:"type:varchar(255)" json:"api_endpoint,omitempty"` // xray API endpoint
	APIPort       int          `gorm:"default:10085" json:"api_port"`                   // xray API port
	Status        NodeStatus   `gorm:"type:varchar(20);default:'offline';index" json:"status"`
	Country       string       `gorm:"type:varchar(50)" json:"country,omitempty"`
	City          string       `gorm:"type:varchar(100)" json:"city,omitempty"`
	Tags          string       `gorm:"type:varchar(255)" json:"tags,omitempty"` // Comma-separated tags
	MaxUsers      int          `gorm:"default:0" json:"max_users"`              // 0 = unlimited
	CurrentUsers  int          `gorm:"default:0" json:"current_users"`
	OnlineDevices int          `gorm:"-" json:"online_devices"` // Calculated field: total devices from online users

	// TLS/Reality Configuration
	TLSEnabled bool   `gorm:"default:true" json:"tls_enabled"`
	SNI        string `gorm:"type:varchar(255)" json:"sni,omitempty"` // Server Name Indication

	// Reality Protocol Settings (for VLESS-Reality)
	RealityEnabled     bool   `gorm:"default:false" json:"reality_enabled"`
	RealityDest        string `gorm:"type:varchar(255)" json:"reality_dest,omitempty"` // Reality destination (e.g., "www.microsoft.com:443")
	RealityServerNames string `gorm:"type:text" json:"reality_server_names,omitempty"` // Comma-separated server names
	RealityPrivateKey  string `gorm:"type:varchar(255)" json:"reality_private_key,omitempty"`
	RealityPublicKey   string `gorm:"type:varchar(255)" json:"reality_public_key,omitempty"`
	RealityShortIds    string `gorm:"type:text" json:"reality_short_ids,omitempty"` // Comma-separated short IDs

	InboundTag  string         `gorm:"type:varchar(50);default:'proxy'" json:"inbound_tag"`
	LastCheckAt *time.Time     `gorm:"index" json:"last_heartbeat,omitempty"` // Index for online node queries
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index:idx_node_lookup,priority:2" json:"-"`

	// Relationships
	Plans       []SubscriptionPlan `gorm:"many2many:plan_nodes" json:"-"`
	TrafficLogs []TrafficLog       `gorm:"foreignKey:NodeID" json:"-"`
}

// IsAvailable checks if the node can accept new users.
func (n *Node) IsAvailable() bool {
	if n.Status != NodeStatusOnline {
		return false
	}
	if n.MaxUsers > 0 && n.CurrentUsers >= n.MaxUsers {
		return false
	}
	return true
}

// NodeResponse is the node data structure for API responses.
type NodeResponse struct {
	ID         uint         `json:"id"`
	Name       string       `json:"name"`
	Address    string       `json:"address"`
	Port       int          `json:"port"`
	Protocol   ProtocolType `json:"protocol"`
	Status     NodeStatus   `json:"status"`
	Country    string       `json:"country,omitempty"`
	City       string       `json:"city,omitempty"`
	TLSEnabled bool         `json:"tls_enabled"`
}

// ToResponse converts Node to a safe response structure for clients.
func (n *Node) ToResponse() NodeResponse {
	return NodeResponse{
		ID:         n.ID,
		Name:       n.Name,
		Address:    n.Address,
		Port:       n.Port,
		Protocol:   n.Protocol,
		Status:     n.Status,
		Country:    n.Country,
		City:       n.City,
		TLSEnabled: n.TLSEnabled,
	}
}
