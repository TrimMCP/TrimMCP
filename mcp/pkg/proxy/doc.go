// Package proxy implements an HTTP proxy server for the Model Context Protocol (MCP).
//
// The proxy intercepts MCP requests, applies intelligent filtering strategies to reduce
// response sizes, and returns optimized responses while maintaining full protocol compatibility.
//
// # Overview
//
// SlimMCP proxy acts as a transparent middleware layer between MCP clients and servers,
// reducing token usage by 60-85% through smart field filtering and response optimization.
//
// # Basic Usage
//
//	cfg := &config.Config{
//		ListenAddress: "0.0.0.0:8080",
//		Strategy:      config.StrategyMinimal,
//		Servers: map[string]config.ServerConfig{
//			"notion": {
//				URL: "http://localhost:3000",
//			},
//		},
//	}
//
//	server, err := proxy.NewServer(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	if err := server.Start(); err != nil {
//		log.Fatal(err)
//	}
//
// # Features
//
//   - Multi-server routing with path-based server selection
//   - Configurable filtering strategies (minimal, smart, explicit)
//   - Built-in Prometheus metrics collection
//   - Health check endpoints
//   - Graceful shutdown handling
//   - Request/response logging
//
// # Filtering Strategies
//
// The proxy supports multiple filtering strategies:
//
//   - Minimal: Returns only critical fields (id, type, url, title)
//   - Smart: Analyzes request context to infer needed fields
//   - Explicit: Client specifies fields via _meta.fieldSelection
//
// # Performance
//
// The proxy adds minimal overhead:
//
//   - <10ms p95 latency for minimal strategy
//   - <15ms p95 latency for explicit strategy
//   - 500+ req/sec throughput on 2 vCPU, 4GB RAM
//
// # Thread Safety
//
// All exported types and functions in this package are safe for concurrent use.
package proxy
