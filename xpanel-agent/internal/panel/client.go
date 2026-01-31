// Package panel provides a client for communicating with the management panel.
package panel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"xpanel-agent/internal/models"
)

// Client handles communication with the management panel.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new panel API client.
func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// SendHeartbeat sends a heartbeat to the panel.
func (c *Client) SendHeartbeat(heartbeat *models.HeartbeatRequest) error {
	url := fmt.Sprintf("%s/api/v1/node-agent/heartbeat", c.baseURL)
	return c.doRequest("POST", url, heartbeat, nil)
}

// FetchUserSync retrieves the list of users to provision on this node.
func (c *Client) FetchUserSync(nodeID uint) ([]models.UserConfig, error) {
	url := fmt.Sprintf("%s/api/v1/node-agent/%d/sync", c.baseURL, nodeID)

	var response models.UserSyncResponse
	if err := c.doRequest("GET", url, nil, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("sync request failed")
	}

	return response.Data.Users, nil
}

// ReportTraffic sends traffic statistics to the panel.
func (c *Client) ReportTraffic(report *models.TrafficReportRequest) error {
	url := fmt.Sprintf("%s/api/v1/node-agent/traffic", c.baseURL)
	return c.doRequest("POST", url, report, nil)
}

// ReportActivity sends user activity data to the panel.
func (c *Client) ReportActivity(report *models.ActivityReportRequest) error {
	url := fmt.Sprintf("%s/api/v1/node-agent/activity", c.baseURL)
	return c.doRequest("POST", url, report, nil)
}

// doRequest performs an HTTP request to the panel API.
func (c *Client) doRequest(method, url string, reqBody, respBody interface{}) error {
	var body io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("panel API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
