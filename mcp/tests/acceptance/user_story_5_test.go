package acceptance

import (
	"encoding/json"
	"testing"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// TestUS5_NativeFieldSelectionDelegation tests that when a downstream server
// advertises native field selection support, the proxy delegates filtering
// to the server and passes the response through.
//
// Acceptance Criteria:
// - Given downstream server advertises fieldSelection support
// - When proxy receives request
// - Then adds _meta to downstream request and passes response through
func TestUS5_NativeFieldSelectionDelegation(t *testing.T) {

	// Setup: Start mock server that advertises field selection support
	mockServer := startCapableMockServer(t, 3020, map[string]interface{}{
		"test-tool": generateLargeResponse(50),
	}, true) // advertises field selection support
	defer mockServer.Close()

	// Setup: Start proxy
	proxy := startProxyServer(t, &proxyConfig{
		listenPort:  8093,
		metricsPort: 9096,
		strategy:    "smart",
		servers: map[string]string{
			"capable": mockServer.URL,
		},
	})
	defer proxy.Close()

	// Execute: Send request with explicit field selection
	req := &types.MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: &types.RequestParams{
			Name: "test-tool",
			Arguments: map[string]interface{}{
				"id": "test-123",
			},
		},
		Meta: &types.RequestMeta{
			FieldSelection: &types.FieldSelectionSpec{
				Include:  []string{"id", "title", "status"},
				Strategy: "explicit",
			},
		},
		ID: 1,
	}

	resp, err := sendRequest(proxy.URL+"/capable", req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Verify: Response successful
	if resp.Error != nil {
		t.Fatalf("Request returned error: %v", resp.Error)
	}

	// Verify: Response metadata indicates server-side filtering
	if resp.Meta == nil || resp.Meta.FieldSelection == nil {
		t.Fatal("Response missing filtering metadata")
	}

	meta := resp.Meta.FieldSelection

	// Verify: FilteredBy should be "server" not "proxy"
	if meta.FilteredBy != "server" {
		t.Errorf("Expected filteredBy='server', got '%s'", meta.FilteredBy)
	}

	// Verify: Response should be filtered by server
	var resultData map[string]interface{}
	if err := json.Unmarshal(resp.Result, &resultData); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Verify only requested fields are present (server did the filtering)
	expectedFields := map[string]bool{
		"id":     true,
		"title":  true,
		"status": true,
	}

	for field := range resultData {
		if !expectedFields[field] {
			t.Errorf("Unexpected field '%s' in response (should be filtered by server)", field)
		}
	}

	t.Log("✓ Proxy delegated filtering to capable server")
}

// TestUS5_FallbackToProxyFiltering tests that when a downstream server
// does not support field selection, the proxy applies filtering itself.
//
// Acceptance Criteria:
// - Given downstream server without field selection support
// - When proxy receives request
// - Then fetches full response and applies filtering
func TestUS5_FallbackToProxyFiltering(t *testing.T) {

	// Setup: Start mock server WITHOUT field selection support
	mockServer := startCapableMockServer(t, 3021, map[string]interface{}{
		"test-tool": generateLargeResponse(50),
	}, false) // does NOT advertise field selection support
	defer mockServer.Close()

	// Setup: Start proxy
	proxy := startProxyServer(t, &proxyConfig{
		listenPort:  8094,
		metricsPort: 9097,
		strategy:    "smart",
		servers: map[string]string{
			"legacy": mockServer.URL,
		},
	})
	defer proxy.Close()

	// Execute: Send request with explicit field selection
	req := &types.MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: &types.RequestParams{
			Name: "test-tool",
			Arguments: map[string]interface{}{
				"id": "test-456",
			},
		},
		Meta: &types.RequestMeta{
			FieldSelection: &types.FieldSelectionSpec{
				Include:  []string{"id", "title", "url"},
				Strategy: "explicit",
			},
		},
		ID: 1,
	}

	resp, err := sendRequest(proxy.URL+"/legacy", req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Verify: Response successful
	if resp.Error != nil {
		t.Fatalf("Request returned error: %v", resp.Error)
	}

	// Verify: Response metadata indicates proxy-side filtering
	if resp.Meta == nil || resp.Meta.FieldSelection == nil {
		t.Fatal("Response missing filtering metadata")
	}

	meta := resp.Meta.FieldSelection

	// Verify: FilteredBy should be "proxy"
	if meta.FilteredBy != "proxy" {
		t.Errorf("Expected filteredBy='proxy', got '%s'", meta.FilteredBy)
	}

	// Verify: Should have filtering statistics
	if meta.OriginalSize <= 0 {
		t.Error("Expected originalSize > 0 for proxy filtering")
	}

	if meta.FilteredSize <= 0 {
		t.Error("Expected filteredSize > 0 for proxy filtering")
	}

	if meta.Savings <= 0 {
		t.Error("Expected savings > 0 for proxy filtering")
	}

	// Verify: Response was filtered
	var resultData map[string]interface{}
	if err := json.Unmarshal(resp.Result, &resultData); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Should have filtered to only requested fields
	expectedFields := []string{"id", "title", "url"}
	for _, field := range expectedFields {
		if _, ok := resultData[field]; !ok {
			t.Errorf("Expected field '%s' not in response", field)
		}
	}

	// Should not have extra fields (proxy filtered them out)
	if len(resultData) > len(expectedFields)+2 { // Allow some tolerance
		t.Errorf("Too many fields in response: expected ~%d, got %d",
			len(expectedFields), len(resultData))
	}

	t.Log("✓ Proxy applied filtering for non-capable server")
}

// TestUS5_CapabilityCaching tests that once capability information is
// detected, it is cached and not re-queried on subsequent requests.
//
// Acceptance Criteria:
// - Given capability detection ran once
// - When subsequent requests arrive
// - Then uses cached capability info without re-querying
func TestUS5_CapabilityCaching(t *testing.T) {

	// Setup: Start mock server that tracks capability queries
	capabilityQueryCount := 0
	mockServer := startTrackingCapableMockServer(t, 3022, map[string]interface{}{
		"test-tool": generateLargeResponse(30),
	}, true, &capabilityQueryCount)
	defer mockServer.Close()

	// Setup: Start proxy
	proxy := startProxyServer(t, &proxyConfig{
		listenPort:  8095,
		metricsPort: 9098,
		strategy:    "smart",
		servers: map[string]string{
			"tracked": mockServer.URL,
		},
	})
	defer proxy.Close()

	// Execute: Send multiple requests
	for i := 0; i < 5; i++ {
		req := &types.MCPRequest{
			JSONRPC: "2.0",
			Method:  "tools/call",
			Params: &types.RequestParams{
				Name: "test-tool",
				Arguments: map[string]interface{}{
					"id": "test-loop",
				},
			},
			ID: i + 1,
		}

		resp, err := sendRequest(proxy.URL+"/tracked", req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}

		if resp.Error != nil {
			t.Fatalf("Request %d returned error: %v", i, resp.Error)
		}
	}

	// Verify: Capability should have been queried only once
	actualQueryCount := mockServer.mock.GetCapabilityQueryCount()

	if actualQueryCount > 1 {
		t.Errorf("Expected capability queried once, got %d queries", actualQueryCount)
	}

	if actualQueryCount == 0 {
		t.Error("Expected capability to be queried at least once")
	}

	t.Logf("✓ Capability cached after first query (%d queries for 5 requests)",
		actualQueryCount)
}

// TestUS5_CapabilityDetectionFallback tests that if capability detection
// fails, the proxy falls back to proxy-side filtering gracefully.
//
// Acceptance Criteria:
// - Given capability detection fails (server error, timeout, etc.)
// - When proxy processes request
// - Then falls back to proxy filtering without failing the request
func TestUS5_CapabilityDetectionFallback(t *testing.T) {

	// Setup: Start mock server that fails capability queries
	mockServer := startFaultyCapableMockServer(t, 3023, map[string]interface{}{
		"test-tool": generateLargeResponse(30),
	})
	defer mockServer.Close()

	// Setup: Start proxy
	proxy := startProxyServer(t, &proxyConfig{
		listenPort:  8096,
		metricsPort: 9099,
		strategy:    "smart",
		servers: map[string]string{
			"faulty": mockServer.URL,
		},
	})
	defer proxy.Close()

	// Execute: Send request
	req := &types.MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: &types.RequestParams{
			Name: "test-tool",
			Arguments: map[string]interface{}{
				"id": "test-fallback",
			},
		},
		ID: 1,
	}

	resp, err := sendRequest(proxy.URL+"/faulty", req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Verify: Request succeeded despite capability detection failure
	if resp.Error != nil {
		t.Fatalf("Request should succeed with fallback, got error: %v", resp.Error)
	}

	// Verify: Should have filtering metadata indicating proxy filtering
	if resp.Meta == nil || resp.Meta.FieldSelection == nil {
		t.Fatal("Response missing filtering metadata")
	}

	meta := resp.Meta.FieldSelection

	// Verify: Proxy did the filtering (fallback behavior)
	if meta.FilteredBy != "proxy" {
		t.Errorf("Expected fallback to proxy filtering, got filteredBy='%s'", meta.FilteredBy)
	}

	// Verify: Response was filtered successfully
	if meta.Applied != true {
		t.Error("Expected filtering to be applied despite capability detection failure")
	}

	t.Log("✓ Proxy gracefully fell back to proxy filtering when capability detection failed")
}

// Helper functions are in helpers.go
