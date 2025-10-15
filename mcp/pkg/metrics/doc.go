// Package metrics provides Prometheus metrics collection for SlimMCP.
//
// This package instruments the proxy to collect operational metrics including
// request counts, token savings, latency, and throughput.
//
// # Overview
//
// Metrics are exposed via HTTP endpoint for Prometheus scraping and provide
// visibility into proxy performance and optimization effectiveness.
//
// # Basic Usage
//
//	collector := metrics.NewCollector()
//
//	// Record request
//	collector.RecordRequest("notion", "notion-fetch", 200)
//
//	// Record token savings
//	collector.RecordTokenSavings("notion", "notion-fetch", 2847, 421)
//
//	// Record latency
//	start := time.Now()
//	// ... process request ...
//	collector.RecordLatency("notion", time.Since(start))
//
// # Available Metrics
//
// # Request Metrics
//
//   - mcp_optimizer_requests_total: Total requests processed
//   - mcp_optimizer_requests_per_second: Current request rate
//   - mcp_optimizer_request_duration_seconds: Request latency histogram
//
// # Token Metrics
//
//   - mcp_optimizer_tokens_original_total: Original response tokens
//   - mcp_optimizer_tokens_filtered_total: Filtered response tokens
//   - mcp_optimizer_tokens_saved_total: Tokens saved
//   - mcp_optimizer_savings_percentage: Savings as percentage
//
// # Performance Metrics
//
//   - mcp_optimizer_filtering_duration_seconds: Time spent filtering
//   - mcp_optimizer_filtering_accuracy: Filtering accuracy (0.0-1.0)
//
// # Error Metrics
//
//   - mcp_optimizer_errors_total: Total errors by type
//   - mcp_optimizer_fallbacks_total: Fallback to unfiltered responses
//
// # Labels
//
// Metrics are labeled for detailed analysis:
//
//   - server: Upstream server name (notion, slack, etc.)
//   - tool: MCP tool name (notion-fetch, slack-search, etc.)
//   - strategy: Filtering strategy (minimal, smart, explicit)
//   - status: HTTP status code or error type
//
// # HTTP Endpoint
//
// Metrics are exposed at /metrics on the metrics port (default 9090):
//
//	curl http://localhost:9090/metrics
//
// # Prometheus Configuration
//
// Example Prometheus scrape config:
//
//	scrape_configs:
//	  - job_name: 'slimmcp'
//	    static_configs:
//	      - targets: ['localhost:9090']
//	    scrape_interval: 15s
//
// # Grafana Dashboard
//
// A pre-built Grafana dashboard is available in examples/grafana-dashboard.json
//
// # Performance Impact
//
// Metrics collection adds minimal overhead:
//
//   - <0.1ms per request
//   - Lock-free counters
//   - Async histogram updates
//
// # Thread Safety
//
// All metrics collectors are safe for concurrent use.
package metrics
