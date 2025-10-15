package benchmarks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/slimmcp/slimmcp/pkg/config"
	"github.com/slimmcp/slimmcp/pkg/filter"
	"github.com/slimmcp/slimmcp/pkg/proxy"
	"github.com/slimmcp/slimmcp/pkg/types"
	"github.com/slimmcp/slimmcp/tests/util"
)

// TestLatencyBenchmark verifies proxy latency meets performance targets.
// Target: Minimal <10ms p95, Smart <15ms p95 (from performance-targets.md)
func TestLatencyBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping latency test in short mode")
	}

	tests := []struct {
		name     string
		strategy string
		target   time.Duration // p95 target
	}{
		{"Minimal Strategy", "minimal", 10 * time.Millisecond},
		{"Explicit Strategy", "explicit", 10 * time.Millisecond},
		// Smart strategy would be 15ms when implemented (Phase 8)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start mock MCP server
			mockServer := util.NewMockMCPServer()
			go mockServer.Start(3200)
			defer mockServer.Stop()
			time.Sleep(100 * time.Millisecond)

			// Configure mock response
			response := util.CreateNotionPageResponse()
			if err := mockServer.SetResponse("test-tool", response); err != nil {
				t.Fatalf("Failed to set mock response: %v", err)
			}

			// Create proxy configuration
			cfg := &config.Config{
				Proxy: config.ProxyConfig{
					ListenAddress: "127.0.0.1:8200",
					Strategy:      tt.strategy,
					Servers: map[string]*config.ServerConfig{
						"test": {
							URL:      "http://127.0.0.1:3200",
							Strategy: tt.strategy,
						},
					},
				},
				Optimization: config.OptimizationConfig{
					AlwaysKeep:   []string{"id", "type", "url"},
					AlwaysRemove: []string{},
					Safety: config.SafetyConfig{
						FallbackOnError: true,
					},
				},
				Monitoring: config.MonitoringConfig{
					Enabled: false,
				},
			}

			// Start proxy
			proxyServer := proxy.NewServer(cfg)
			go proxyServer.Start()
			defer proxyServer.Stop(context.Background())
			time.Sleep(200 * time.Millisecond)

			// Measure latency over 100 requests (reduced for faster testing)
			const numRequests = 100
			latencies := make([]time.Duration, numRequests)

			for i := 0; i < numRequests; i++ {
				request := &types.MCPRequest{
					JSONRPC: "2.0",
					Method:  "tools/call",
					Params: &types.RequestParams{
						Name: "test-tool",
						Arguments: map[string]interface{}{
							"id": fmt.Sprintf("test-%d", i),
						},
					},
					ID: i + 1,
				}

				if tt.strategy == "explicit" {
					request.Meta = &types.RequestMeta{
						FieldSelection: &types.FieldSelectionSpec{
							Include:  []string{"id", "url", "type"},
							Strategy: "explicit",
						},
					}
				}

				start := time.Now()
				resp, err := sendRequestToProxy("http://127.0.0.1:8200/test", request)
				latency := time.Since(start)

				if err != nil {
					t.Fatalf("Request %d failed: %v", i, err)
				}

				if resp.Error != nil {
					t.Fatalf("Request %d returned error: %v", i, resp.Error)
				}

				latencies[i] = latency
			}

			// Calculate percentiles
			p50 := calculatePercentile(latencies, 50)
			p95 := calculatePercentile(latencies, 95)
			p99 := calculatePercentile(latencies, 99)
			avg := calculateAverage(latencies)

			t.Logf("Strategy: %s", tt.strategy)
			t.Logf("  Requests: %d", numRequests)
			t.Logf("  Average: %v", avg)
			t.Logf("  p50: %v", p50)
			t.Logf("  p95: %v (target: <%v)", p95, tt.target)
			t.Logf("  p99: %v", p99)

			if p95 > tt.target {
				t.Errorf("p95 latency %v exceeds target %v", p95, tt.target)
			}
		})
	}
}

// TestThroughputTest verifies proxy can handle sustained load.
// Target: 500+ req/sec per instance (from performance-targets.md)
func TestThroughputTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping throughput test in short mode")
	}

	// Start mock MCP server
	mockServer := util.NewMockMCPServer()
	go mockServer.Start(3201)
	defer mockServer.Stop()
	time.Sleep(100 * time.Millisecond)

	// Configure mock response
	response := util.CreateNotionPageResponse()
	if err := mockServer.SetResponse("test-tool", response); err != nil {
		t.Fatalf("Failed to set mock response: %v", err)
	}

	// Create proxy configuration
	cfg := &config.Config{
		Proxy: config.ProxyConfig{
			ListenAddress: "127.0.0.1:8201",
			Strategy:      "minimal",
			Servers: map[string]*config.ServerConfig{
				"test": {
					URL:      "http://127.0.0.1:3201",
					Strategy: "minimal",
				},
			},
		},
		Optimization: config.OptimizationConfig{
			AlwaysKeep: []string{"id", "type", "url"},
			Safety: config.SafetyConfig{
				FallbackOnError: true,
			},
		},
		Monitoring: config.MonitoringConfig{
			Enabled: false,
		},
	}

	// Start proxy
	proxyServer := proxy.NewServer(cfg)
	go proxyServer.Start()
	defer proxyServer.Stop(context.Background())
	time.Sleep(200 * time.Millisecond)

	// Test parameters (reduced for faster testing)
	const (
		duration    = 5 * time.Second
		targetRPS   = 500 // Target: 500+ req/sec
		concurrency = 20  // Concurrent clients
	)

	// Throughput test
	var (
		totalRequests atomic.Int64
		successCount  atomic.Int64
		errorCount    atomic.Int64
	)

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	wg := &sync.WaitGroup{}

	// Launch concurrent workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			client := &http.Client{
				Timeout: 5 * time.Second,
			}

			requestID := workerID * 10000

			for {
				select {
				case <-ctx.Done():
					return
				default:
					totalRequests.Add(1)
					requestID++

					request := &types.MCPRequest{
						JSONRPC: "2.0",
						Method:  "tools/call",
						Params: &types.RequestParams{
							Name: "test-tool",
							Arguments: map[string]interface{}{
								"id": fmt.Sprintf("test-%d", requestID),
							},
						},
						ID: requestID,
					}

					reqBody, _ := json.Marshal(request)
					httpReq, _ := http.NewRequest("POST", "http://127.0.0.1:8201/test", bytes.NewReader(reqBody))
					httpReq.Header.Set("Content-Type", "application/json")

					resp, err := client.Do(httpReq)
					if err != nil {
						errorCount.Add(1)
						continue
					}

					if resp.StatusCode == 200 {
						successCount.Add(1)
					} else {
						errorCount.Add(1)
					}

					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
			}
		}(i)
	}

	wg.Wait()

	// Calculate metrics
	total := totalRequests.Load()
	success := successCount.Load()
	errors := errorCount.Load()
	actualRPS := float64(total) / duration.Seconds()
	successRate := (float64(success) / float64(total)) * 100

	t.Logf("Throughput Test Results:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Concurrent Workers: %d", concurrency)
	t.Logf("  Total Requests: %d", total)
	t.Logf("  Successful: %d", success)
	t.Logf("  Errors: %d", errors)
	t.Logf("  Actual RPS: %.2f (target: >=%d)", actualRPS, targetRPS)
	t.Logf("  Success Rate: %.2f%%", successRate)

	if actualRPS < float64(targetRPS) {
		t.Errorf("Throughput %.2f req/sec is below target %d req/sec", actualRPS, targetRPS)
	}

	if successRate < 95.0 {
		t.Errorf("Success rate %.2f%% is below target 95%%", successRate)
	}
}

// TestTokenSavingsTest verifies filtering achieves target token reduction.
// Target: 70-90% savings (from performance-targets.md)
func TestTokenSavingsTest(t *testing.T) {
	tests := []struct {
		name       string
		response   interface{}
		fields     []string
		minSavings float64
		maxSavings float64
	}{
		{
			name:       "Notion Page Response",
			response:   util.CreateNotionPageResponse(),
			fields:     []string{"id", "object", "url"},
			minSavings: 70.0,
			maxSavings: 90.0,
		},
		{
			name:       "Slack Message Response",
			response:   util.CreateSlackMessageResponse(),
			fields:     []string{"id", "type", "text"},
			minSavings: 70.0,
			maxSavings: 90.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate original size
			originalTokens := filter.EstimateTokens(tt.response)

			// Apply filtering
			filterEngine := filter.NewEngine()

			filtered, err := filterEngine.Filter(context.Background(), tt.response, tt.fields)
			if err != nil {
				t.Fatal(err)
			}

			// Calculate filtered size
			filteredTokens := filter.EstimateTokens(filtered)

			// Calculate savings
			savings := (float64(originalTokens-filteredTokens) / float64(originalTokens)) * 100

			t.Logf("%s:", tt.name)
			t.Logf("  Original tokens: %d", originalTokens)
			t.Logf("  Filtered tokens: %d", filteredTokens)
			t.Logf("  Savings: %.2f%% (target: %.0f-%.0f%%)", savings, tt.minSavings, tt.maxSavings)

			if savings < tt.minSavings {
				t.Errorf("Token savings %.2f%% is below minimum target %.0f%%", savings, tt.minSavings)
			}

			if savings > tt.maxSavings {
				t.Logf("Warning: Savings %.2f%% exceeds maximum target %.0f%% (may be filtering too aggressively)", savings, tt.maxSavings)
			}
		})
	}
}

// TestConcurrentRequestsTest verifies proxy handles concurrent load safely.
// Target: 100 concurrent requests, all succeed, no race conditions
func TestConcurrentRequestsTest(t *testing.T) {
	const numConcurrent = 100

	// Start mock MCP server
	mockServer := util.NewMockMCPServer()
	go mockServer.Start(3202)
	defer func() {
		mockServer.Stop()
		time.Sleep(300 * time.Millisecond) // Allow ports to be released (TIME_WAIT)
	}()
	time.Sleep(200 * time.Millisecond) // Wait for mock server to start

	// Configure mock response
	response := util.CreateNotionPageResponse()
	if err := mockServer.SetResponse("test-tool", response); err != nil {
		t.Fatalf("Failed to set mock response: %v", err)
	}

	// Create proxy configuration
	cfg := &config.Config{
		Proxy: config.ProxyConfig{
			ListenAddress: "127.0.0.1:8202",
			Strategy:      "minimal",
			Servers: map[string]*config.ServerConfig{
				"test": {
					URL:      "http://127.0.0.1:3202",
					Strategy: "minimal",
				},
			},
		},
		Optimization: config.OptimizationConfig{
			AlwaysKeep: []string{"id", "type", "url"},
			Safety: config.SafetyConfig{
				FallbackOnError: true,
			},
		},
		Monitoring: config.MonitoringConfig{
			Enabled: false,
		},
	}

	// Start proxy
	proxyServer := proxy.NewServer(cfg)
	go proxyServer.Start()
	defer func() {
		proxyServer.Stop(context.Background())
		time.Sleep(300 * time.Millisecond) // Allow ports to be released (TIME_WAIT)
	}()
	time.Sleep(300 * time.Millisecond) // Wait for proxy to start

	// Send 100 concurrent requests
	var (
		wg           sync.WaitGroup
		successCount atomic.Int64
		errorCount   atomic.Int64
	)

	start := time.Now()

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			request := &types.MCPRequest{
				JSONRPC: "2.0",
				Method:  "tools/call",
				Params: &types.RequestParams{
					Name: "test-tool",
					Arguments: map[string]interface{}{
						"id": fmt.Sprintf("concurrent-%d", id),
					},
				},
				ID: id + 1,
			}

			resp, err := sendRequestToProxy("http://127.0.0.1:8202/test", request)
			if err != nil {
				errorCount.Add(1)
				t.Errorf("Request %d failed: %v", id, err)
				return
			}

			if resp.Error != nil {
				errorCount.Add(1)
				t.Errorf("Request %d returned error: %v", id, resp.Error)
				return
			}

			// Verify response has result
			if resp.Result == nil {
				errorCount.Add(1)
				t.Errorf("Request %d missing result", id)
				return
			}

			successCount.Add(1)
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	success := successCount.Load()
	errors := errorCount.Load()

	t.Logf("Concurrent Request Test Results:")
	t.Logf("  Concurrent Requests: %d", numConcurrent)
	t.Logf("  Successful: %d", success)
	t.Logf("  Errors: %d", errors)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Throughput: %.2f req/sec", float64(numConcurrent)/duration.Seconds())

	if success != numConcurrent {
		t.Errorf("Expected all %d requests to succeed, got %d successful, %d errors", numConcurrent, success, errors)
	}
}

// Helper functions

func sendRequestToProxy(url string, req *types.MCPRequest) (*types.MCPResponse, error) {
	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 5 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Parse response
	var resp types.MCPResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}

func calculatePercentile(durations []time.Duration, percentile float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	// Sort durations (simple bubble sort for small datasets)
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * percentile / 100.0)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

func calculateAverage(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations))
}
