package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// MCPClient handles communication with downstream MCP servers
type MCPClient struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewMCPClient creates a new MCP client
func NewMCPClient(timeout time.Duration) *MCPClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &MCPClient{
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		timeout: timeout,
	}
}

// Forward forwards a request to the downstream MCP server
func (c *MCPClient) Forward(ctx context.Context, serverURL string, request *types.MCPRequest, headers map[string]string) ([]byte, error) {
	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, types.NewTimeoutError("forward request", c.timeout.String(), err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downstream server returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ForwardRaw forwards a raw JSON-RPC request to the downstream server
func (c *MCPClient) ForwardRaw(ctx context.Context, serverURL string, requestBody []byte, headers map[string]string) ([]byte, error) {
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, types.NewTimeoutError("forward request", c.timeout.String(), err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downstream server returned status %d", resp.StatusCode)
	}

	return body, nil
}

// Ping performs a health check on a downstream server
// T062: Connection health check
func (c *MCPClient) Ping(ctx context.Context, serverURL string) error {
	// Send a minimal request to check if server is responsive
	request := &types.MCPRequest{
		JSONRPC: "2.0",
		Method:  "ping",
		ID:      "health-check",
	}

	// Use short timeout for health checks
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.Forward(pingCtx, serverURL, request, nil)
	return err
}

// CheckServerHealth checks the health of a server and returns availability info
// T062: Connection health check
func (c *MCPClient) CheckServerHealth(ctx context.Context, serverURL string) map[string]interface{} {
	startTime := time.Now()
	err := c.Ping(ctx, serverURL)
	latency := time.Since(startTime)

	health := map[string]interface{}{
		"url":        serverURL,
		"available":  err == nil,
		"latency_ms": latency.Milliseconds(),
		"checked_at": startTime.Format(time.RFC3339),
	}

	if err != nil {
		health["error"] = err.Error()
	}

	return health
}

// ForwardWithNativeFieldSelection forwards a request with native field selection to the downstream server
// T070: Native field selection request forwarding
func (c *MCPClient) ForwardWithNativeFieldSelection(ctx context.Context, serverURL string, request *types.MCPRequest, headers map[string]string) ([]byte, error) {
	// Simply forward the request as-is - the _meta.fieldSelection should already be in the request
	// The downstream server will handle the field selection natively
	return c.Forward(ctx, serverURL, request, headers)
}
