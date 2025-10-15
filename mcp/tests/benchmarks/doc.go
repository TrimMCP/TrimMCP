// Package benchmarks contains performance and load tests for SlimMCP proxy.
//
// These tests validate the proxy meets performance targets defined in
// specs/001-trim/performance-targets.md.
//
// # Test Suite
//
// TestLatencyBenchmark - Measures request latency across filtering strategies.
// Validates p95 latency < 10ms for minimal/explicit strategies.
//
// TestThroughputTest - Validates sustained request throughput under load.
// Target: 500+ requests/second per instance with 95%+ success rate.
//
// TestTokenSavingsTest - Verifies filtering achieves 70-90% token reduction.
// Tests both Notion and Slack responses with realistic data structures.
//
// TestConcurrentRequestsTest - Validates safe concurrent request handling.
// Sends 100 parallel requests, verifies all succeed without race conditions.
//
// # Running Tests
//
//	# Run all performance tests
//	go test -v ./tests/benchmarks/
//
//	# Skip long-running tests
//	go test -short ./tests/benchmarks/
//
//	# Run with race detector
//	go test -race ./tests/benchmarks/
//
//	# Run specific test
//	go test -run=TestLatencyBenchmark ./tests/benchmarks/
//
// # Performance Targets
//
// From specs/001-trim/performance-targets.md:
//   - Latency: p95 <10ms (minimal), p95 <15ms (smart)
//   - Throughput: 500+ req/sec per instance
//   - Token Savings: 70-90% reduction
//   - Concurrency: 100+ simultaneous requests
//   - Availability: 99.9% (three nines)
package benchmarks
