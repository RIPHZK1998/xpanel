package panel

import (
	"fmt"
)

// NodeConfigResponse represents the node configuration from panel.
type NodeConfigResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Node NodeConfig `json:"node"`
	} `json:"data"`
}

// NodeConfig represents node configuration from panel.
type NodeConfig struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	TLSEnabled  bool   `json:"tls_enabled"`
	SNI         string `json:"sni"`
	InboundTag  string `json:"inbound_tag"`
	APIEndpoint string `json:"api_endpoint"`
	APIPort     int    `json:"api_port"`

	// Reality settings
	RealityEnabled     bool   `json:"reality_enabled"`
	RealityDest        string `json:"reality_dest"`
	RealityServerNames string `json:"reality_server_names"`
	RealityPrivateKey  string `json:"reality_private_key"`
	RealityPublicKey   string `json:"reality_public_key"`
	RealityShortIds    string `json:"reality_short_ids"`
}

// FetchNodeConfig retrieves the node configuration from the panel.
func (c *Client) FetchNodeConfig(nodeID uint) (*NodeConfig, error) {
	url := fmt.Sprintf("%s/api/v1/node-agent/%d/config", c.baseURL, nodeID)

	var response NodeConfigResponse
	if err := c.doRequest("GET", url, nil, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to fetch node config")
	}

	return &response.Data.Node, nil
}
