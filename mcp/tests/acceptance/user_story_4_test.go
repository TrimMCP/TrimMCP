package acceptance

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// TestUS4_MultiServerRouting tests that requests to different server endpoints
// route correctly to their respective downstream servers.
//
// Acceptance Criteria:
// - Given proxy configured with Notion, Slack, Google Drive
// - When send requests to /notion, /slack, /gdrive
// - Then each routes to correct server and returns filtered responses
func TestUS4_MultiServerRouting(t *testing.T) {

	// Setup: Start mock servers for each service
	notionServer := startMockMCPServer(t, 3010, map[string]interface{}{
		"notion-fetch": generateNotionResponse(45),
	})
	defer notionServer.Close()

	slackServer := startMockMCPServer(t, 3011, map[string]interface{}{
		"slack-get-message": generateSlackResponse(30),
	})
	defer slackServer.Close()

	gdriveServer := startMockMCPServer(t, 3012, map[string]interface{}{
		"gdrive-get-file": generateGDriveResponse(35),
	})
	defer gdriveServer.Close()

	// Setup: Start proxy with multiple servers configured
	proxy := startProxyServer(t, &proxyConfig{
		listenPort:  8090,
		metricsPort: 9093,
		strategy:    "minimal",
		servers: map[string]string{
			"notion": notionServer.URL,
			"slack":  slackServer.URL,
			"gdrive": gdriveServer.URL,
		},
	})
	defer proxy.Close()

	// Test cases for each server
	testCases := []struct {
		name       string
		endpoint   string
		toolName   string
		expectedID string
	}{
		{
			name:       "Notion routing",
			endpoint:   "/notion",
			toolName:   "notion-fetch",
			expectedID: "notion-page-123",
		},
		{
			name:       "Slack routing",
			endpoint:   "/slack",
			toolName:   "slack-get-message",
			expectedID: "slack-msg-456",
		},
		{
			name:       "Google Drive routing",
			endpoint:   "/gdrive",
			toolName:   "gdrive-get-file",
			expectedID: "gdrive-file-789",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute: Send request to specific server endpoint
			req := &types.MCPRequest{
				JSONRPC: "2.0",
				Method:  "tools/call",
				Params: &types.RequestParams{
					Name: tc.toolName,
					Arguments: map[string]interface{}{
						"id": tc.expectedID,
					},
				},
				ID: 1,
			}

			resp, err := sendRequest(proxy.URL+tc.endpoint, req)
			if err != nil {
				t.Fatalf("Request to %s failed: %v", tc.endpoint, err)
			}

			// Verify: Response was successful
			if resp.Error != nil {
				t.Fatalf("Request returned error: %v", resp.Error)
			}

			// Verify: Response was filtered
			if resp.Meta == nil || resp.Meta.FieldSelection == nil {
				t.Error("Response missing filtering metadata")
			}

			// Verify: Response contains the expected data
			var resultData map[string]interface{}
			if err := json.Unmarshal(resp.Result, &resultData); err != nil {
				t.Fatalf("Failed to parse result: %v", err)
			}

			// Verify ID is preserved
			if id, ok := resultData["id"].(string); !ok || id != tc.expectedID {
				t.Errorf("Expected ID %s, got %v", tc.expectedID, resultData["id"])
			}

			t.Logf("✓ %s routed correctly and returned filtered response", tc.name)
		})
	}
}

// TestUS4_ToolSpecificFiltering tests that different tools with different schemas
// have tool-specific filtering rules applied correctly.
//
// Acceptance Criteria:
// - Given different tools have different schemas
// - When proxy processes requests for each
// - Then applies tool-specific filtering rules correctly
func TestUS4_ToolSpecificFiltering(t *testing.T) {

	// Setup: Start mock server with multiple tools
	mockServer := startMockMCPServer(t, 3013, map[string]interface{}{
		"notion-fetch":      generateNotionResponse(45),
		"slack-get-message": generateSlackResponse(30),
		"gdrive-get-file":   generateGDriveResponse(35),
	})
	defer mockServer.Close()

	// Setup: Start proxy with tool-specific defaults
	proxy := startProxyServer(t, &proxyConfig{
		listenPort:  8091,
		metricsPort: 9094,
		strategy:    "smart",
		servers: map[string]string{
			"multi": mockServer.URL,
		},
	})
	defer proxy.Close()

	// Test cases for different tools
	testCases := []struct {
		name              string
		toolName          string
		expectedMinFields []string
		expectFieldCount  int
	}{
		{
			name:              "Notion page",
			toolName:          "notion-fetch",
			expectedMinFields: []string{"id", "type", "url", "title"},
			expectFieldCount:  5, // Critical fields + minimal strategy
		},
		{
			name:              "Slack message",
			toolName:          "slack-get-message",
			expectedMinFields: []string{"id", "type", "url"},
			expectFieldCount:  4,
		},
		{
			name:              "Google Drive file",
			toolName:          "gdrive-get-file",
			expectedMinFields: []string{"id", "type", "url"},
			expectFieldCount:  4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute: Send request
			req := &types.MCPRequest{
				JSONRPC: "2.0",
				Method:  "tools/call",
				Params: &types.RequestParams{
					Name: tc.toolName,
					Arguments: map[string]interface{}{
						"id": "test-123",
					},
				},
				ID: 1,
			}

			resp, err := sendRequest(proxy.URL+"/multi", req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			// Verify: Response has filtering metadata
			if resp.Meta == nil || resp.Meta.FieldSelection == nil {
				t.Fatal("Response missing filtering metadata")
			}

			meta := resp.Meta.FieldSelection

			// Verify: Expected fields are present
			fieldMap := make(map[string]bool)
			for _, field := range meta.FieldsKept {
				fieldMap[field] = true
			}

			for _, expectedField := range tc.expectedMinFields {
				if !fieldMap[expectedField] {
					t.Errorf("Expected field '%s' not in kept fields: %v",
						expectedField, meta.FieldsKept)
				}
			}

			// Verify: Field count is reasonable
			if len(meta.FieldsKept) > tc.expectFieldCount+2 {
				t.Errorf("Too many fields kept: expected ~%d, got %d",
					tc.expectFieldCount, len(meta.FieldsKept))
			}

			t.Logf("✓ %s: kept %d fields with %.1f%% savings",
				tc.name, len(meta.FieldsKept), meta.Savings)
		})
	}
}

// TestUS4_AuthHeaderPassthrough tests that authentication headers are
// correctly passed through to downstream servers.
//
// Acceptance Criteria:
// - Given downstream server auth requirements
// - When client includes auth headers
// - Then proxy passes through unchanged and maintains session
func TestUS4_AuthHeaderPassthrough(t *testing.T) {

	// Setup: Start mock server that validates auth headers
	receivedHeaders := make(map[string]string)
	mockServer := startAuthenticatedMockServer(t, 3014, map[string]interface{}{
		"test-tool": generateLargeResponse(20),
	}, receivedHeaders)
	defer mockServer.Close()

	// Setup: Start proxy
	proxy := startProxyServer(t, &proxyConfig{
		listenPort:  8092,
		metricsPort: 9095,
		strategy:    "minimal",
		servers: map[string]string{
			"auth": mockServer.URL,
		},
	})
	defer proxy.Close()

	// Test cases for different auth methods
	testCases := []struct {
		name    string
		headers map[string]string
	}{
		{
			name: "Bearer token",
			headers: map[string]string{
				"Authorization": "Bearer test-token-12345",
			},
		},
		{
			name: "API Key header",
			headers: map[string]string{
				"X-API-Key": "api-key-67890",
			},
		},
		{
			name: "Multiple auth headers",
			headers: map[string]string{
				"Authorization": "Bearer multi-auth-token",
				"X-API-Key":     "backup-api-key",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear received headers
			for k := range receivedHeaders {
				delete(receivedHeaders, k)
			}

			// Execute: Send request with auth headers
			req := &types.MCPRequest{
				JSONRPC: "2.0",
				Method:  "tools/call",
				Params: &types.RequestParams{
					Name: "test-tool",
					Arguments: map[string]interface{}{
						"id": "test-123",
					},
				},
				ID: 1,
			}

			resp, err := sendRequestWithHeaders(proxy.URL+"/auth", req, tc.headers)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			// Verify: Response successful
			if resp.Error != nil {
				t.Fatalf("Request returned error: %v", resp.Error)
			}

			// Verify: All auth headers were passed through
			// Get captured headers from mock server
			capturedHeaders := mockServer.mock.GetCapturedHeaders()

			// Debug: print all captured headers
			t.Logf("Captured headers: %v", capturedHeaders)

			// Normalize captured headers to lowercase for case-insensitive comparison
			normalizedCaptured := make(map[string]string)
			for k, v := range capturedHeaders {
				normalizedCaptured[strings.ToLower(k)] = v
			}

			for headerName, expectedValue := range tc.headers {
				actualValue, ok := normalizedCaptured[strings.ToLower(headerName)]
				if !ok {
					t.Errorf("Header '%s' was not passed through. Available headers: %v", headerName, capturedHeaders)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("Header '%s': expected '%s', got '%s'",
						headerName, expectedValue, actualValue)
				}
			}

			t.Logf("✓ %s: all auth headers passed through correctly", tc.name)
		})
	}
}

// Helper functions are in helpers.go
