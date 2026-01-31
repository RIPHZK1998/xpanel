// Package xray provides integration with xray-core VPN nodes.
package xray

// UserConfig represents a user configuration for xray-core.
type UserConfig struct {
	UUID     string `json:"id"`
	Email    string `json:"email"`    // Used as identifier in xray
	Level    int    `json:"level"`    // User access level
	AlterId  int    `json:"alterId"`  // For VMess protocol
	Flow     string `json:"flow"`     // For VLESS protocol (e.g., "xtls-rprx-vision")
	Password string `json:"password"` // For Trojan protocol
}

// InboundUser represents a user in xray inbound configuration.
type InboundUser struct {
	Email   string      `json:"email"`
	Account UserAccount `json:"account,omitempty"`
}

// UserAccount represents the account details for different protocols.
type UserAccount struct {
	// Common
	ID string `json:"id,omitempty"` // UUID

	// VLESS specific
	Flow string `json:"flow,omitempty"`

	// VMess specific
	AlterID  int `json:"alterId,omitempty"`
	Security int `json:"security,omitempty"`

	// Trojan specific
	Password string `json:"password,omitempty"`
}

// AddUserRequest represents a request to add a user to xray.
type AddUserRequest struct {
	Tag  string      `json:"tag"` // Inbound tag
	User InboundUser `json:"user"`
}

// RemoveUserRequest represents a request to remove a user from xray.
type RemoveUserRequest struct {
	Tag   string `json:"tag"`   // Inbound tag
	Email string `json:"email"` // User identifier
}

// StatsRequest represents a request for user statistics.
type StatsRequest struct {
	Name   string `json:"name"`
	Reset_ bool   `json:"reset"`
}

// StatsResponse represents the response from stats query.
type StatsResponse struct {
	Stat Stat `json:"stat"`
}

// Stat represents a single statistic entry.
type Stat struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

// UserStats represents aggregated user statistics.
type UserStats struct {
	Email         string `json:"email"`
	UploadBytes   int64  `json:"upload_bytes"`
	DownloadBytes int64  `json:"download_bytes"`
}

// NodeInfo represents information about an xray node.
type NodeInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	APIEndpoint string `json:"api_endpoint"`
	APIPort     int    `json:"api_port"`
	InboundTag  string `json:"inbound_tag"`
	TLSEnabled  bool   `json:"tls_enabled"`
	SNI         string `json:"sni"`
}

// ClientConfig represents the VPN configuration for a client app.
type ClientConfig struct {
	Protocol  string          `json:"protocol"`
	Address   string          `json:"address"`
	Port      int             `json:"port"`
	UUID      string          `json:"uuid,omitempty"`
	Password  string          `json:"password,omitempty"`
	Flow      string          `json:"flow,omitempty"`
	TLS       TLSConfig       `json:"tls,omitempty"`
	Transport TransportConfig `json:"transport,omitempty"`
	ShareLink string          `json:"share_link,omitempty"` // vless://, vmess://, or trojan:// link
}

// TLSConfig represents TLS configuration for client.
type TLSConfig struct {
	Enabled     bool     `json:"enabled"`
	ServerName  string   `json:"server_name,omitempty"`
	ALPN        []string `json:"alpn,omitempty"`
	Fingerprint string   `json:"fingerprint,omitempty"`
}

// TransportConfig represents transport configuration for client.
type TransportConfig struct {
	Type    string            `json:"type"` // tcp, ws, grpc, etc.
	Path    string            `json:"path,omitempty"`
	Host    string            `json:"host,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}
