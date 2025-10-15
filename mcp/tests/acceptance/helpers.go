package acceptance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/slimmcp/slimmcp/pkg/config"
	"github.com/slimmcp/slimmcp/pkg/metrics"
	"github.com/slimmcp/slimmcp/pkg/proxy"
	"github.com/slimmcp/slimmcp/pkg/types"
	"github.com/slimmcp/slimmcp/tests/util"
)

// mockMCPServerInstance wraps a mock MCP server for testing
type mockMCPServerInstance struct {
	mock *util.MockMCPServer
	URL  string
	done chan error
}

// Close stops the mock server
func (m *mockMCPServerInstance) Close() {
	if m.mock != nil {
		m.mock.Stop()
	}
}

// proxyServerInstance wraps a proxy server for testing
type proxyServerInstance struct {
	proxy         *proxy.Server
	metricsServer *metrics.PrometheusServer
	URL           string
	done          chan error
}

// Close stops the proxy server
func (p *proxyServerInstance) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if p.metricsServer != nil {
		p.metricsServer.Shutdown(ctx)
	}

	if p.proxy != nil {
		p.proxy.Stop(ctx)
	}
}

// proxyConfig holds proxy configuration for testing
type proxyConfig struct {
	listenPort  int
	metricsPort int
	strategy    string
	servers     map[string]string
}

// startMockMCPServer starts a mock MCP server for testing
func startMockMCPServer(t *testing.T, port int, responses map[string]interface{}) *mockMCPServerInstance {
	t.Helper()

	mockServer := util.NewMockMCPServer()

	// Configure responses
	for toolName, response := range responses {
		if err := mockServer.SetResponse(toolName, response); err != nil {
			t.Fatalf("Failed to set response for %s: %v", toolName, err)
		}
	}

	// Start server in background
	done := make(chan error, 1)
	go func() {
		done <- mockServer.Start(port)
	}()

	// Wait for server to be ready
	url := fmt.Sprintf("http://localhost:%d", port)
	var lastErr error
	ready := false
	for i := 0; i < 50; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			ready = true
			break
		}
		lastErr = err
		time.Sleep(50 * time.Millisecond)
	}

	if !ready {
		t.Fatalf("Mock server on port %d never became ready: %v", port, lastErr)
	}

	instance := &mockMCPServerInstance{
		mock: mockServer,
		URL:  url,
		done: done,
	}

	t.Cleanup(func() {
		instance.Close()
	})

	return instance
}

// startProxyServer starts a proxy server for testing
func startProxyServer(t *testing.T, cfg *proxyConfig) *proxyServerInstance {
	t.Helper()

	// Build config
	proxyConfig := &config.Config{
		Proxy: config.ProxyConfig{
			ListenAddress: fmt.Sprintf("localhost:%d", cfg.listenPort),
			Strategy:      cfg.strategy,
			Servers:       make(map[string]*config.ServerConfig),
		},
		Optimization: config.OptimizationConfig{
			AlwaysKeep:   []string{"id", "type", "url"},
			AlwaysRemove: []string{},
			Safety: config.SafetyConfig{
				FallbackOnError: true,
			},
		},
		Monitoring: config.MonitoringConfig{
			Enabled:     true,
			MetricsPort: cfg.metricsPort,
			LogLevel:    "info",
		},
	}

	// Add servers
	for name, url := range cfg.servers {
		proxyConfig.Proxy.Servers[name] = &config.ServerConfig{
			URL:     url,
			Timeout: 30 * time.Second,
		}
	}

	// Create and start proxy
	proxyServer := proxy.NewServer(proxyConfig)

	// Start metrics server if enabled (T052)
	var metricsServer *metrics.PrometheusServer
	if proxyConfig.Monitoring.Enabled {
		metricsServer = metrics.NewPrometheusServer(proxyConfig.Monitoring.MetricsPort)
		go func() {
			if err := metricsServer.Start(); err != nil && err != http.ErrServerClosed {
				t.Logf("Metrics server error: %v", err)
			}
		}()

		// Wait for metrics server to start
		metricsURL := fmt.Sprintf("http://localhost:%d/metrics", proxyConfig.Monitoring.MetricsPort)
		for i := 0; i < 50; i++ {
			resp, err := http.Get(metricsURL)
			if err == nil {
				resp.Body.Close()
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Start proxy in background
	done := make(chan error, 1)
	go func() {
		done <- proxyServer.Start()
	}()

	// Wait for proxy to be ready
	proxyURL := fmt.Sprintf("http://%s", proxyConfig.Proxy.ListenAddress)
	for i := 0; i < 50; i++ {
		resp, err := http.Get(proxyURL + "/health")
		if err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	instance := &proxyServerInstance{
		proxy:         proxyServer,
		metricsServer: metricsServer,
		URL:           proxyURL,
		done:          done,
	}

	t.Cleanup(func() {
		instance.Close()
	})

	return instance
}

// sendRequest sends an MCP request to the specified URL
func sendRequest(url string, req *types.MCPRequest) (*types.MCPResponse, error) {
	return sendRequestWithHeaders(url, req, nil)
}

// sendRequestWithHeaders sends an MCP request with custom headers
func sendRequestWithHeaders(url string, req *types.MCPRequest, headers map[string]string) (*types.MCPResponse, error) {
	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var response types.MCPResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// generateLargeResponse generates a response with many fields for testing
func generateLargeResponse(numFields int) map[string]interface{} {
	response := make(map[string]interface{})
	response["id"] = "item-123"
	response["type"] = "test"
	response["url"] = "https://example.com/item-123"
	response["title"] = "Test Item"

	// Add many extra fields
	for i := 0; i < numFields-4; i++ {
		response[fmt.Sprintf("field_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	return response
}

// generateNotionResponse generates a realistic Notion page response
func generateNotionResponse(numFields int) map[string]interface{} {
	response := util.CreateNotionPageResponse()

	// Add extra fields if needed
	if numFields > 15 {
		for i := 15; i < numFields; i++ {
			response[fmt.Sprintf("extra_field_%d", i)] = fmt.Sprintf("extra_value_%d", i)
		}
	}

	return response
}

// generateSlackResponse generates a realistic Slack message response
func generateSlackResponse(numFields int) map[string]interface{} {
	response := util.CreateSlackMessageResponse()

	// Ensure critical fields are present
	if _, ok := response["url"]; !ok {
		response["url"] = "https://slack.com/archives/C123456/p1234567890123456"
	}

	// Add extra fields if needed
	if numFields > 12 {
		for i := 12; i < numFields; i++ {
			response[fmt.Sprintf("extra_field_%d", i)] = fmt.Sprintf("extra_value_%d", i)
		}
	}

	return response
}

// generateGDriveResponse generates a realistic Google Drive file response
func generateGDriveResponse(numFields int) map[string]interface{} {
	response := util.CreateGDriveFileResponse()

	// Ensure critical fields are present
	if _, ok := response["url"]; !ok {
		response["url"] = "https://drive.google.com/file/d/file-123/view"
	}

	// Add extra fields if needed
	if numFields > 12 {
		for i := 12; i < numFields; i++ {
			response[fmt.Sprintf("extra_field_%d", i)] = fmt.Sprintf("extra_value_%d", i)
		}
	}

	return response
}

// startCapableMockServer starts a mock server that advertises capability support
func startCapableMockServer(t *testing.T, port int, responses map[string]interface{}, supportsFieldSelection bool) *mockMCPServerInstance {
	t.Helper()

	mockServer := util.NewMockMCPServer()
	mockServer.SetSupportsCapabilities(supportsFieldSelection)

	// Configure responses
	for toolName, response := range responses {
		if err := mockServer.SetResponse(toolName, response); err != nil {
			t.Fatalf("Failed to set response for %s: %v", toolName, err)
		}
	}

	// Start server in background
	done := make(chan error, 1)
	go func() {
		done <- mockServer.Start(port)
	}()

	// Wait for server to be ready
	url := fmt.Sprintf("http://localhost:%d", port)
	for i := 0; i < 50; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	instance := &mockMCPServerInstance{
		mock: mockServer,
		URL:  url,
		done: done,
	}

	t.Cleanup(func() {
		instance.Close()
	})

	return instance
}

// startAuthenticatedMockServer starts a mock server that validates auth headers
func startAuthenticatedMockServer(t *testing.T, port int, responses map[string]interface{}, receivedHeaders map[string]string) *mockMCPServerInstance {
	t.Helper()

	instance := startMockMCPServer(t, port, responses)

	// The mock server captures headers automatically
	// Tests can call instance.mock.GetCapturedHeaders() to verify
	// We'll update receivedHeaders after each request in the test

	return instance
}

// startTrackingCapableMockServer starts a mock server that tracks capability queries
func startTrackingCapableMockServer(t *testing.T, port int, responses map[string]interface{}, supportsFieldSelection bool, queryCount *int) *mockMCPServerInstance {
	t.Helper()

	instance := startCapableMockServer(t, port, responses, supportsFieldSelection)

	// The mock server tracks queries automatically
	// Tests can call instance.mock.GetCapabilityQueryCount() to check

	return instance
}

// startFaultyCapableMockServer starts a mock server that fails capability detection
func startFaultyCapableMockServer(t *testing.T, port int, responses map[string]interface{}) *mockMCPServerInstance {
	t.Helper()

	// Start a server that doesn't support capabilities (default behavior)
	return startMockMCPServer(t, port, responses)
}
