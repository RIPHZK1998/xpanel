// Package xray provides xray-core integration and management.
package xray

import (
	"context"
	"fmt"

	"xpanel-agent/internal/models"

	handlerCmd "github.com/xtls/xray-core/app/proxyman/command"
	statsCmd "github.com/xtls/xray-core/app/stats/command"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/proxy/vless"
	"github.com/xtls/xray-core/proxy/vmess"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// APIClient handles communication with xray-core API via gRPC.
type APIClient struct {
	apiAddress string
	apiPort    int
	inboundTag string
	conn       *grpc.ClientConn
}

// NewAPIClient creates a new xray API client.
func NewAPIClient(apiAddress string, apiPort int, inboundTag string) *APIClient {
	return &APIClient{
		apiAddress: apiAddress,
		apiPort:    apiPort,
		inboundTag: inboundTag,
	}
}

// Connect establishes gRPC connection to xray API.
func (c *APIClient) Connect() error {
	target := fmt.Sprintf("%s:%d", c.apiAddress, c.apiPort)
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to xray API: %w", err)
	}
	c.conn = conn
	return nil
}

// Close closes the gRPC connection.
func (c *APIClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// AddUser adds a user to xray-core using gRPC.
func (c *APIClient) AddUser(user *models.XrayUser) error {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	client := handlerCmd.NewHandlerServiceClient(c.conn)

	// Build user account based on protocol
	var account *serial.TypedMessage
	if user.Protocol == "vless" || user.Protocol == "" {
		// VLESS account with optional flow for Reality/XTLS
		vlessAccount := &vless.Account{
			Id: user.UUID,
		}
		// Set flow for xtls-rprx-vision (required for Reality)
		if user.Flow != "" {
			vlessAccount.Flow = user.Flow
		}
		account = serial.ToTypedMessage(vlessAccount)
	} else if user.Protocol == "vmess" {
		// VMess account
		account = serial.ToTypedMessage(&vmess.Account{
			Id: user.UUID,
		})
	}

	// Create the user
	protoUser := &protocol.User{
		Level:   uint32(user.Level),
		Email:   user.Email,
		Account: account,
	}

	// Create AlterInboundRequest with AddUserOperation
	req := &handlerCmd.AlterInboundRequest{
		Tag: c.inboundTag,
		Operation: serial.ToTypedMessage(&handlerCmd.AddUserOperation{
			User: protoUser,
		}),
	}

	_, err := client.AlterInbound(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to add user: %w", err)
	}

	return nil
}

// RemoveUser removes a user from xray-core using gRPC.
func (c *APIClient) RemoveUser(email string) error {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	client := handlerCmd.NewHandlerServiceClient(c.conn)

	// Create AlterInboundRequest with RemoveUserOperation
	req := &handlerCmd.AlterInboundRequest{
		Tag: c.inboundTag,
		Operation: serial.ToTypedMessage(&handlerCmd.RemoveUserOperation{
			Email: email,
		}),
	}

	_, err := client.AlterInbound(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to remove user: %w", err)
	}

	return nil
}

// GetUserStats retrieves traffic statistics for a user using gRPC.
func (c *APIClient) GetUserStats(email string) (*models.XrayStats, error) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	statsclient := statsCmd.NewStatsServiceClient(c.conn)
	stats := &models.XrayStats{}

	// Get upload stats
	uploadPattern := fmt.Sprintf("user>>>%s>>>traffic>>>uplink", email)
	uploadReq := &statsCmd.QueryStatsRequest{
		Pattern: uploadPattern,
		Reset_:  false,
	}
	uploadResp, err := statsclient.QueryStats(context.Background(), uploadReq)
	if err == nil && len(uploadResp.Stat) > 0 {
		stats.UploadBytes = uploadResp.Stat[0].Value
	}

	// Get download stats
	downloadPattern := fmt.Sprintf("user>>>%s>>>traffic>>>downlink", email)
	downloadReq := &statsCmd.QueryStatsRequest{
		Pattern: downloadPattern,
		Reset_:  false,
	}
	downloadResp, err := statsclient.QueryStats(context.Background(), downloadReq)
	if err == nil && len(downloadResp.Stat) > 0 {
		stats.DownloadBytes = downloadResp.Stat[0].Value
	}

	return stats, nil
}

// ResetUserStats resets traffic statistics for a user.
func (c *APIClient) ResetUserStats(email string) error {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	statsclient := statsCmd.NewStatsServiceClient(c.conn)

	// Reset upload stats
	uploadPattern := fmt.Sprintf("user>>>%s>>>traffic>>>uplink", email)
	uploadReq := &statsCmd.QueryStatsRequest{
		Pattern: uploadPattern,
		Reset_:  true,
	}
	_, _ = statsclient.QueryStats(context.Background(), uploadReq)

	// Reset download stats
	downloadPattern := fmt.Sprintf("user>>>%s>>>traffic>>>downlink", email)
	downloadReq := &statsCmd.QueryStatsRequest{
		Pattern: downloadPattern,
		Reset_:  true,
	}
	_, err := statsclient.QueryStats(context.Background(), downloadReq)

	return err
}

// GetOnlineIPs returns the online IP addresses for a specific user.
// This uses Xray-core's built-in GetStatsOnlineIpList API.
// Returns a map of IP address -> last seen timestamp (unix).
func (c *APIClient) GetOnlineIPs(email string) (map[string]int64, error) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	statsclient := statsCmd.NewStatsServiceClient(c.conn)

	// Query for user's online IP list
	req := &statsCmd.GetStatsRequest{
		Name:   fmt.Sprintf("user>>>%s>>>online", email),
		Reset_: false,
	}

	resp, err := statsclient.GetStatsOnlineIpList(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to get online IPs: %w", err)
	}

	return resp.Ips, nil
}

// GetAllOnlineUsers returns a list of all currently online user emails.
// This uses Xray-core's built-in GetAllOnlineUsers API.
func (c *APIClient) GetAllOnlineUsers() ([]string, error) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	statsclient := statsCmd.NewStatsServiceClient(c.conn)

	resp, err := statsclient.GetAllOnlineUsers(
		context.Background(),
		&statsCmd.GetAllOnlineUsersRequest{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get online users: %w", err)
	}

	return resp.Users, nil
}
