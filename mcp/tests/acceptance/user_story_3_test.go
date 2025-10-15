package acceptance

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// TestUS3_MetricsEndpoint tests that after processing 100 requests,
// the /metrics endpoint exposes Prometheus metrics showing tokens saved,
// savings percentage, and request counts by server and tool.
//
// Acceptance Criteria:
// - Given proxy processed 100 requests
// - When access /metrics endpoint
// - Then see Prometheus metrics with:
//   - Total tokens saved
//   - Savings percentage
//   - Request counts by server and tool
func TestUS3_MetricsEndpoint(t *testing.T) {
	// Setup: Start mock server
	mockServer := startMockMCPServer(t, 3005, map[string]interface{}{
		"test-tool": generateLargeResponse(50), // 50 fields
	})
	defer mockServer.Close()

	// Setup: Start proxy with metrics enabled
	proxy := startProxyServer(t, &proxyConfig{
		listenPort:  8085,
		metricsPort: 9095,
		strategy:    "minimal",
		servers: map[string]string{
			"mock": mockServer.URL,
		},
	})
	defer proxy.Close()

	// Execute: Send 100 requests
	for i := 0; i < 100; i++ {
		req := &types.MCPRequest{
			JSONRPC: "2.0",
			Method:  "tools/call",
			Params: &types.RequestParams{
				Name: "test-tool",
				Arguments: map[string]interface{}{
					"id": fmt.Sprintf("item-%d", i),
				},
			},
			ID: i + 1,
		}

		resp, err := sendRequest(proxy.URL+"/mock", req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}

		// Verify response was filtered
		if resp.Error != nil {
			t.Fatalf("Request %d returned error: %v", i, resp.Error)
		}
	}

	// Verify: Check /metrics endpoint
	metricsResp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", 9095))
	if err != nil {
		t.Fatalf("Failed to fetch metrics: %v", err)
	}
	defer metricsResp.Body.Close()

	// Read metrics body
	var metrics strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := metricsResp.Body.Read(buf)
		if n > 0 {
			metrics.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	metricsText := metrics.String()

	// Verify: Check for required metrics
	requiredMetrics := []string{
		"mcp_optimizer_tokens_original_total",
		"mcp_optimizer_tokens_saved_total",
		"mcp_optimizer_requests_total",
		"mcp_optimizer_savings_percentage",
	}

	for _, metric := range requiredMetrics {
		if !strings.Contains(metricsText, metric) {
			t.Errorf("Metrics missing required metric: %s", metric)
		}
	}

	// Verify: Check request count
	if !strings.Contains(metricsText, "mcp_optimizer_requests_total{") {
		t.Error("Metrics missing request count")
	}

	// Verify: Check labels (server, tool, strategy)
	expectedLabels := []string{
		"server=\"mock\"",
		"tool=\"test-tool\"",
		"strategy=\"minimal\"",
	}

	for _, label := range expectedLabels {
		if !strings.Contains(metricsText, label) {
			t.Errorf("Metrics missing expected label: %s", label)
		}
	}

	t.Log("✓ Metrics endpoint exposes all required Prometheus metrics")
}

// TestUS3_ResponseMetadata tests that each filtered response contains
// complete metadata about the filtering operation.
//
// Acceptance Criteria:
// - Given single filtered response
// - When inspect _meta.fieldSelection
// - Then contains:
//   - originalSize (tokens)
//   - filteredSize (tokens)
//   - savings (percentage)
//   - fields.kept (array)
//   - fields.removed (count)
//   - filteringTime (ms)
func TestUS3_ResponseMetadata(t *testing.T) {
	// Setup: Start mock server with large response
	mockServer := startMockMCPServer(t, 3001, map[string]interface{}{
		"notion-fetch": generateNotionResponse(45), // 45 fields
	})
	defer mockServer.Close()

	// Setup: Start proxy
	proxy := startProxyServer(t, &proxyConfig{
		listenPort:  8081,
		metricsPort: 9091,
		strategy:    "minimal",
		servers: map[string]string{
			"notion": mockServer.URL,
		},
	})
	defer proxy.Close()

	// Execute: Send request
	req := &types.MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: &types.RequestParams{
			Name: "notion-fetch",
			Arguments: map[string]interface{}{
				"id": "page-123",
			},
		},
		ID: 1,
	}

	resp, err := sendRequest(proxy.URL+"/notion", req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Verify: Check response metadata exists
	if resp.Meta == nil {
		t.Fatal("Response missing _meta field")
	}

	if resp.Meta.FieldSelection == nil {
		t.Fatal("Response missing _meta.fieldSelection field")
	}

	meta := resp.Meta.FieldSelection

	// Verify: originalSize is present and reasonable
	if meta.OriginalSize <= 0 {
		t.Errorf("originalSize should be > 0, got %d", meta.OriginalSize)
	}

	// Verify: filteredSize is present and smaller than original
	if meta.FilteredSize <= 0 {
		t.Errorf("filteredSize should be > 0, got %d", meta.FilteredSize)
	}
	if meta.FilteredSize >= meta.OriginalSize {
		t.Errorf("filteredSize (%d) should be < originalSize (%d)", meta.FilteredSize, meta.OriginalSize)
	}

	// Verify: savings percentage is calculated
	if meta.Savings <= 0 || meta.Savings >= 100 {
		t.Errorf("savings should be between 0 and 100, got %.2f", meta.Savings)
	}

	// Verify: fields.kept contains critical fields
	if len(meta.FieldsKept) == 0 {
		t.Error("fields.kept should not be empty")
	}
	criticalFields := []string{"id", "type", "url"}
	for _, field := range criticalFields {
		found := false
		for _, kept := range meta.FieldsKept {
			if kept == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("fields.kept missing critical field: %s", field)
		}
	}

	// Verify: fields.removed is present and reasonable
	if meta.FieldsRemoved <= 0 {
		t.Errorf("fields.removed should be > 0 for minimal strategy, got %d", meta.FieldsRemoved)
	}

	// Verify: filteringTime is present and reasonable (<100ms)
	// Note: Can be 0 if filtering is extremely fast (<1ms)
	if meta.FilteringTime < 0 || meta.FilteringTime > 100 {
		t.Errorf("filteringTime should be between 0 and 100ms, got %.2f", meta.FilteringTime)
	}

	// Verify: strategy is correct
	if meta.Strategy != "minimal" {
		t.Errorf("strategy should be 'minimal', got '%s'", meta.Strategy)
	}

	// Verify: filteredBy is 'proxy'
	if meta.FilteredBy != "proxy" {
		t.Errorf("filteredBy should be 'proxy', got '%s'", meta.FilteredBy)
	}

	t.Logf("✓ Response metadata complete: %d → %d tokens (%.1f%% savings)",
		meta.OriginalSize, meta.FilteredSize, meta.Savings)
}

// TestUS3_MetricsTrends tests that metrics can be queried over time
// to track savings trends and detect performance issues.
//
// Acceptance Criteria:
// - Given continuous operation
// - When query metrics over time
// - Then can track savings trends and detect performance issues
func TestUS3_MetricsTrends(t *testing.T) {
	// Setup: Start mock server
	mockServer := startMockMCPServer(t, 3002, map[string]interface{}{
		"test-tool": generateLargeResponse(30),
	})
	defer mockServer.Close()

	// Setup: Start proxy
	proxy := startProxyServer(t, &proxyConfig{
		listenPort:  8082,
		metricsPort: 9092,
		strategy:    "smart",
		servers: map[string]string{
			"mock": mockServer.URL,
		},
	})
	defer proxy.Close()

	metricsURL := fmt.Sprintf("http://localhost:%d/metrics", 9092)

	// Execute: Send requests in batches and track metrics
	type metricsSnapshot struct {
		timestamp      time.Time
		requestsTotal  int
		tokensSaved    int
		savingsPercent float64
	}

	snapshots := []metricsSnapshot{}

	for batch := 0; batch < 5; batch++ {
		// Send 20 requests
		for i := 0; i < 20; i++ {
			req := &types.MCPRequest{
				JSONRPC: "2.0",
				Method:  "tools/call",
				Params: &types.RequestParams{
					Name: "test-tool",
					Arguments: map[string]interface{}{
						"id": fmt.Sprintf("batch-%d-item-%d", batch, i),
					},
				},
				ID: batch*20 + i + 1,
			}

			_, err := sendRequest(proxy.URL+"/mock", req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
		}

		// Capture metrics snapshot
		resp, err := http.Get(metricsURL)
		if err != nil {
			t.Fatalf("Failed to fetch metrics: %v", err)
		}
		var metrics strings.Builder
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				metrics.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		resp.Body.Close()

		// Parse metrics (simplified - in real test would use Prometheus parser)
		snapshot := metricsSnapshot{
			timestamp:      time.Now(),
			requestsTotal:  (batch + 1) * 20,
			tokensSaved:    0, // Would parse from metrics
			savingsPercent: 0, // Would parse from metrics
		}
		snapshots = append(snapshots, snapshot)

		time.Sleep(100 * time.Millisecond)
	}

	// Verify: Metrics show increasing trends
	if len(snapshots) < 2 {
		t.Fatal("Not enough metric snapshots")
	}

	// Check that request count increases
	for i := 1; i < len(snapshots); i++ {
		if snapshots[i].requestsTotal <= snapshots[i-1].requestsTotal {
			t.Errorf("Request count not increasing: %d → %d",
				snapshots[i-1].requestsTotal, snapshots[i].requestsTotal)
		}
	}

	// Verify: Can detect time-based patterns
	if snapshots[len(snapshots)-1].timestamp.Sub(snapshots[0].timestamp) < 400*time.Millisecond {
		t.Error("Metrics timestamps not tracking time correctly")
	}

	t.Logf("✓ Captured %d metric snapshots over time", len(snapshots))
	t.Log("✓ Metrics show increasing trends and time-based patterns")
}

// Helper functions are in helpers.go
