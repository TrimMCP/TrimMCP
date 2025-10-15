package transport

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/slimmcp/slimmcp/tests/transport/harness"
)

// TestBasicTransports validates HTTP, HTTPS, and SSE transports work correctly
// without requiring external credentials or services
func TestBasicTransports(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Setup transports - all using local test servers
	transports := []harness.Transport{
		// HTTP: Plain HTTP test server
		harness.NewHTTPTransport(
			"test-http",
			"http://localhost:3200",
			nil,
		),

		// HTTPS: Test server with self-signed cert
		harness.NewHTTPSTransport(
			"test-https",
			"https://localhost:3201",
			nil,
			true, // skip TLS verification for self-signed cert
		),

		// SSE: Test server with Server-Sent Events
		harness.NewSSETransport(
			"test-sse",
			"http://localhost:3202", // SSE endpoint
			"http://localhost:3202", // POST endpoint (same)
			nil,
		),
	}

	// Start all transports
	for _, transport := range transports {
		t.Logf("Starting transport: %s", transport.Name())
		if err := transport.Start(ctx); err != nil {
			t.Fatalf("Failed to start %s: %v", transport.Name(), err)
		}
		defer func(tr harness.Transport) {
			if err := tr.Stop(); err != nil {
				t.Logf("Failed to stop %s: %v", tr.Name(), err)
			}
		}(transport)
	}

	// Wait for all transports to be healthy
	for _, transport := range transports {
		deadline := time.Now().Add(10 * time.Second)
		healthy := false
		for time.Now().Before(deadline) {
			if transport.IsHealthy(ctx) {
				healthy = true
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		if !healthy {
			t.Fatalf("Transport not healthy: %s", transport.Name())
		}
		t.Logf("Transport ready: %s", transport.Name())
	}

	// Get basic test scenarios
	scenarios := harness.CreateBasicScenarios()

	// Run each scenario across all transports
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Logf("Testing scenario: %s", scenario.Description)

			responses := make(map[string]json.RawMessage)

			// Execute on each transport
			for _, transport := range transports {
				if !transport.IsHealthy(ctx) {
					t.Fatalf("Transport became unhealthy: %s", transport.Name())
				}

				timeout := scenario.Timeout
				if timeout == 0 {
					timeout = 5 * time.Second
				}

				execCtx, cancel := context.WithTimeout(ctx, timeout)
				response, err := transport.Execute(execCtx, scenario.Request)
				cancel()

				if err != nil {
					t.Errorf("%s failed: %v", transport.Name(), err)
					continue
				}

				// Validate response if validator provided
				if scenario.ValidateResponse != nil {
					if err := scenario.ValidateResponse(t, response); err != nil {
						t.Errorf("%s validation failed: %v", transport.Name(), err)
					}
				}

				responses[transport.Name()] = response
				t.Logf("%s: success", transport.Name())
			}

			// Compare responses across transports
			if len(responses) >= 2 {
				normalizer := harness.DefaultNormalizer()
				harness.AssertEquivalentResponses(t, normalizer, responses)
				t.Logf("✓ All %d transports produced equivalent responses", len(responses))
			}
		})
	}
}

// TestHTTPTransport validates HTTP transport independently
func TestHTTPTransport(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transport := harness.NewHTTPTransport(
		"test-http",
		"http://localhost:3200",
		nil,
	)

	if err := transport.Start(ctx); err != nil {
		t.Fatalf("Failed to start transport: %v", err)
	}
	defer transport.Stop()

	// Test ping
	request := harness.MustMarshal(harness.MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "ping",
	})

	response, err := transport.Execute(ctx, request)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var resp harness.MCPResponse
	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("Expected success, got error: %v", resp.Error)
	}

	t.Logf("✓ HTTP transport works correctly")
}

// TestHTTPSTransport validates HTTPS transport with self-signed cert
func TestHTTPSTransport(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transport := harness.NewHTTPSTransport(
		"test-https",
		"https://localhost:3201",
		nil,
		true, // skip TLS verification
	)

	if err := transport.Start(ctx); err != nil {
		t.Fatalf("Failed to start transport: %v", err)
	}
	defer transport.Stop()

	// Test ping
	request := harness.MustMarshal(harness.MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "ping",
	})

	response, err := transport.Execute(ctx, request)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var resp harness.MCPResponse
	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("Expected success, got error: %v", resp.Error)
	}

	t.Logf("✓ HTTPS transport works correctly with self-signed cert")
}

// TestSSETransport validates SSE transport
func TestSSETransport(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transport := harness.NewSSETransport(
		"test-sse",
		"http://localhost:3202",
		"http://localhost:3202",
		nil,
	)

	if err := transport.Start(ctx); err != nil {
		t.Fatalf("Failed to start transport: %v", err)
	}
	defer transport.Stop()

	// Wait for SSE connection to stabilize
	time.Sleep(500 * time.Millisecond)

	// Test ping
	request := harness.MustMarshal(harness.MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "ping",
	})

	response, err := transport.Execute(ctx, request)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var resp harness.MCPResponse
	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("Expected success, got error: %v", resp.Error)
	}

	t.Logf("✓ SSE transport works correctly")
}
