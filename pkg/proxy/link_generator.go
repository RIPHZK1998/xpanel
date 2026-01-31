package proxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// GenerateVLESSLink generates a VLESS protocol connection link.
// Format: vless://UUID@ADDRESS:PORT?params#REMARK
func GenerateVLESSLink(uuid, address string, port int, sni, remark string, realityEnabled bool, realityPublicKey, realityShortID string) string {
	// Build query parameters - only include necessary ones
	var params []string

	// Required parameters for TCP transport
	params = append(params, "type=tcp")
	params = append(params, "encryption=none")

	if realityEnabled {
		// Reality configuration (XTLS)
		params = append(params, "security=reality")
		params = append(params, "flow=xtls-rprx-vision")
		params = append(params, "fp=chrome")

		if sni != "" {
			params = append(params, "sni="+url.QueryEscape(sni))
		}
		if realityPublicKey != "" {
			params = append(params, "pbk="+url.QueryEscape(realityPublicKey))
		}
		if realityShortID != "" {
			params = append(params, "sid="+url.QueryEscape(realityShortID))
		}
	} else {
		// TLS configuration
		params = append(params, "security=tls")
		params = append(params, "fp=chrome")

		if sni != "" {
			params = append(params, "sni="+url.QueryEscape(sni))
		}
	}

	queryString := strings.Join(params, "&")
	link := fmt.Sprintf("vless://%s@%s:%d?%s", uuid, address, port, queryString)

	if remark != "" {
		link += "#" + url.QueryEscape(remark)
	}

	return link
}

// GenerateTrojanLink generates a Trojan protocol connection link.
// Format: trojan://PASSWORD@ADDRESS:PORT?params#REMARK
func GenerateTrojanLink(password, address string, port int, sni, remark string) string {
	params := url.Values{}
	params.Set("security", "tls")
	params.Set("type", "tcp")
	params.Set("headerType", "none")
	params.Set("sni", sni)

	link := fmt.Sprintf("trojan://%s@%s:%d?%s", password, address, port, params.Encode())

	if remark != "" {
		link += "#" + url.QueryEscape(remark)
	}

	return link
}

// VMessConfig represents VMess configuration for JSON encoding.
type VMessConfig struct {
	V    string `json:"v"`    // Protocol version
	PS   string `json:"ps"`   // Remark/name
	Add  string `json:"add"`  // Address
	Port int    `json:"port"` // Port
	ID   string `json:"id"`   // UUID
	Aid  int    `json:"aid"`  // Alter ID (usually 0)
	Net  string `json:"net"`  // Network type (tcp, ws, etc)
	Type string `json:"type"` // Header type
	Host string `json:"host"` // Host/SNI
	Path string `json:"path"` // Path (for websocket)
	TLS  string `json:"tls"`  // TLS setting
	SNI  string `json:"sni"`  // Server Name Indication
}

// GenerateVMessLink generates a VMess protocol connection link.
// Format: vmess://BASE64(JSON)
func GenerateVMessLink(uuid, address string, port int, sni, remark string) string {
	config := VMessConfig{
		V:    "2",
		PS:   remark,
		Add:  address,
		Port: port,
		ID:   uuid,
		Aid:  0,
		Net:  "tcp",
		Type: "none",
		Host: sni,
		Path: "",
		TLS:  "tls",
		SNI:  sni,
	}

	jsonData, _ := json.Marshal(config)
	encoded := base64.StdEncoding.EncodeToString(jsonData)

	return "vmess://" + encoded
}

// GenerateSubscriptionBase64 encodes multiple proxy links as a base64 subscription.
// This is the standard format used by proxy clients for bulk import.
func GenerateSubscriptionBase64(links []string) string {
	combined := strings.Join(links, "\n")
	return base64.StdEncoding.EncodeToString([]byte(combined))
}

// NodeLinkInfo represents connection information for a single node.
type NodeLinkInfo struct {
	NodeID   uint   `json:"node_id"`
	NodeName string `json:"node_name"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Link     string `json:"link"`
	QRData   string `json:"qr_data"` // Same as link, for QR code generation
}

// GenerateNodeLink creates a connection link for a node based on its protocol.
func GenerateNodeLink(uuid, nodeAddress string, nodePort int, protocol, nodeName, sni string, realityEnabled bool, realityPublicKey, realityShortID string) string {
	remark := nodeName

	switch protocol {
	case "vless":
		return GenerateVLESSLink(uuid, nodeAddress, nodePort, sni, remark, realityEnabled, realityPublicKey, realityShortID)
	case "trojan":
		return GenerateTrojanLink(uuid, nodeAddress, nodePort, sni, remark)
	case "vmess":
		return GenerateVMessLink(uuid, nodeAddress, nodePort, sni, remark)
	default:
		// Default to VLESS if protocol is unknown
		return GenerateVLESSLink(uuid, nodeAddress, nodePort, sni, remark, realityEnabled, realityPublicKey, realityShortID)
	}
}
