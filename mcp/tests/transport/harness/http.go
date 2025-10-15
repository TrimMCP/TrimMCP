package harness

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPTransport implements Transport interface for HTTP-based MCP servers
type HTTPTransport struct {
	name     string
	endpoint string
	client   *http.Client
	useTLS   bool
	headers  map[string]string
}

// NewHTTPTransport creates a new HTTP transport adapter
func NewHTTPTransport(name, endpoint string, headers map[string]string) *HTTPTransport {
	return &HTTPTransport{
		name:     name,
		endpoint: endpoint,
		useTLS:   false,
		headers:  headers,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewHTTPSTransport creates a new HTTPS transport adapter
func NewHTTPSTransport(name, endpoint string, headers map[string]string, skipVerify bool) *HTTPTransport {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipVerify,
		},
	}

	return &HTTPTransport{
		name:     name,
		endpoint: endpoint,
		useTLS:   true,
		headers:  headers,
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// Name returns the transport type
func (h *HTTPTransport) Name() string {
	protocol := "http"
	if h.useTLS {
		protocol = "https"
	}
	return fmt.Sprintf("%s-%s", protocol, h.name)
}

// Start is a no-op for HTTP transports (server assumed to be running)
func (h *HTTPTransport) Start(ctx context.Context) error {
	// Wait for server to be healthy
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for server to be ready: %w", ctx.Err())
		case <-ticker.C:
			if h.IsHealthy(ctx) {
				return nil
			}
		}
	}
}

// Stop is a no-op for HTTP transports
func (h *HTTPTransport) Stop() error {
	return nil
}

// Execute sends an MCP request via HTTP POST
func (h *HTTPTransport) Execute(ctx context.Context, request json.RawMessage) (json.RawMessage, error) {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", h.endpoint, bytes.NewReader(request))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range h.headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Validate JSON-RPC response format
	var mcpResp MCPResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return nil, fmt.Errorf("invalid JSON-RPC response: %w", err)
	}

	return body, nil
}

// IsHealthy checks if the HTTP endpoint is reachable
func (h *HTTPTransport) IsHealthy(ctx context.Context) bool {
	// Try a simple initialize request
	initReq := MustMarshal(MCPRequest{
		JSONRPC: "2.0",
		ID:      "health-check",
		Method:  "initialize",
		Params: MustMarshal(map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "slimmcp-test",
				"version": "0.1.0",
			},
		}),
	})

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", h.endpoint, bytes.NewReader(initReq))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range h.headers {
		req.Header.Set(key, value)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
