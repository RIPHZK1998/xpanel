package xray

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with xray-core API.
type Client struct {
	apiEndpoint string
	apiPort     int
	httpClient  *http.Client
	inboundTag  string
}

// NewClient creates a new xray API client.
func NewClient(apiEndpoint string, apiPort int, inboundTag string) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		apiPort:     apiPort,
		inboundTag:  inboundTag,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// AddUser adds a user to the xray node.
func (c *Client) AddUser(userConfig *UserConfig) error {
	url := fmt.Sprintf("http://%s:%d/api/users/add", c.apiEndpoint, c.apiPort)

	user := InboundUser{
		Email: userConfig.Email,
		Account: UserAccount{
			ID:       userConfig.UUID,
			Flow:     userConfig.Flow,
			AlterID:  userConfig.AlterId,
			Password: userConfig.Password,
		},
	}

	req := AddUserRequest{
		Tag:  c.inboundTag,
		User: user,
	}

	return c.doRequest("POST", url, req, nil)
}

// RemoveUser removes a user from the xray node.
func (c *Client) RemoveUser(email string) error {
	url := fmt.Sprintf("http://%s:%d/api/users/remove", c.apiEndpoint, c.apiPort)

	req := RemoveUserRequest{
		Tag:   c.inboundTag,
		Email: email,
	}

	return c.doRequest("POST", url, req, nil)
}

// GetUserStats retrieves traffic statistics for a user.
func (c *Client) GetUserStats(email string) (*UserStats, error) {
	stats := &UserStats{
		Email: email,
	}

	// Get upload stats
	uploadName := fmt.Sprintf("user>>>%s>>>traffic>>>uplink", email)
	uploadStats, err := c.queryStat(uploadName, false)
	if err == nil {
		stats.UploadBytes = uploadStats.Value
	}

	// Get download stats
	downloadName := fmt.Sprintf("user>>>%s>>>traffic>>>downlink", email)
	downloadStats, err := c.queryStat(downloadName, false)
	if err == nil {
		stats.DownloadBytes = downloadStats.Value
	}

	return stats, nil
}

// ResetUserStats resets traffic statistics for a user.
func (c *Client) ResetUserStats(email string) error {
	uploadName := fmt.Sprintf("user>>>%s>>>traffic>>>uplink", email)
	_, err := c.queryStat(uploadName, true)
	if err != nil {
		return err
	}

	downloadName := fmt.Sprintf("user>>>%s>>>traffic>>>downlink", email)
	_, err = c.queryStat(downloadName, true)
	return err
}

// queryStat queries xray statistics API.
func (c *Client) queryStat(name string, reset bool) (*Stat, error) {
	url := fmt.Sprintf("http://%s:%d/api/stats/query", c.apiEndpoint, c.apiPort)

	req := StatsRequest{
		Name:   name,
		Reset_: reset,
	}

	var resp StatsResponse
	if err := c.doRequest("POST", url, req, &resp); err != nil {
		return nil, err
	}

	return &resp.Stat, nil
}

// HealthCheck performs a health check on the xray node.
func (c *Client) HealthCheck() error {
	url := fmt.Sprintf("http://%s:%d/api/health", c.apiEndpoint, c.apiPort)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// doRequest performs an HTTP request to xray API.
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

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("xray API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
