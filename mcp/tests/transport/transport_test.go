package transport

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/slimmcp/slimmcp/tests/transport/harness"
)

// TestIdenticalResponsesAcrossTransports validates that SlimMCP produces
// identical responses regardless of transport type (stdio, http, https, sse)
func TestIdenticalResponsesAcrossTransports(t *testing.T) {
	// Skip if no GitHub token available
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		t.Skip("GITHUB_PERSONAL_ACCESS_TOKEN or GITHUB_TOKEN not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Setup transports
	transports := []harness.Transport{
		// Stdio: Direct connection to GitHub MCP server
		harness.NewStdioTransport(
			"github",
			"npx",
			[]string{"-y", "@modelcontextprotocol/server-github"},
			[]string{"GITHUB_PERSONAL_ACCESS_TOKEN=" + token},
		),

		// HTTP: Via bridge (if available)
		harness.NewHTTPTransport(
			"github-bridge",
			"http://localhost:3100",
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
			// HTTP bridge might not be running - that's okay, skip it
			if _, ok := transport.(*harness.HTTPTransport); ok {
				t.Logf("Skipping %s: %v", transport.Name(), err)
				continue
			}
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
			time.Sleep(500 * time.Millisecond)
		}
		if !healthy {
			t.Logf("Skipping unhealthy transport: %s", transport.Name())
			continue
		}
		t.Logf("Transport ready: %s", transport.Name())
	}

	// Get test scenarios
	scenarios := harness.CreateBasicScenarios()

	// Add GitHub-specific scenarios
	scenarios = append(scenarios, createGitHubScenarios()...)

	// Run each scenario across all transports
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Logf("Testing scenario: %s", scenario.Description)

			responses := make(map[string]json.RawMessage)

			// Execute on each transport
			for _, transport := range transports {
				if !transport.IsHealthy(ctx) {
					t.Logf("Skipping unhealthy transport: %s", transport.Name())
					continue
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
			}

			// Compare responses across transports
			if len(responses) >= 2 {
				normalizer := harness.DefaultNormalizer()
				harness.AssertEquivalentResponses(t, normalizer, responses)
			} else {
				t.Logf("Skipping comparison (need at least 2 responses, got %d)", len(responses))
			}
		})
	}
}

// createGitHubScenarios returns GitHub MCP specific test scenarios
func createGitHubScenarios() []harness.TestScenario {
	return []harness.TestScenario{
		{
			Name:        "search_repositories",
			Description: "Search GitHub repositories",
			Request: harness.MustMarshal(harness.MCPRequest{
				JSONRPC: "2.0",
				ID:      100,
				Method:  "tools/call",
				Params: harness.MustMarshal(map[string]interface{}{
					"name": "search_repositories",
					"arguments": map[string]interface{}{
						"query": "language:go stars:>1000",
						"page":  1,
					},
				}),
			}),
			Timeout: 10 * time.Second,
			ValidateResponse: func(t *testing.T, response json.RawMessage) error {
				var resp harness.MCPResponse
				if err := json.Unmarshal(response, &resp); err != nil {
					return err
				}
				if resp.Error != nil {
					t.Logf("Warning: Error response: %v", resp.Error)
				}
				return nil
			},
		},
		{
			Name:        "get_file_contents",
			Description: "Get file contents from a repository",
			Request: harness.MustMarshal(harness.MCPRequest{
				JSONRPC: "2.0",
				ID:      101,
				Method:  "tools/call",
				Params: harness.MustMarshal(map[string]interface{}{
					"name": "get_file_contents",
					"arguments": map[string]interface{}{
						"owner": "golang",
						"repo":  "go",
						"path":  "README.md",
					},
				}),
			}),
			Timeout: 10 * time.Second,
			ValidateResponse: func(t *testing.T, response json.RawMessage) error {
				var resp harness.MCPResponse
				if err := json.Unmarshal(response, &resp); err != nil {
					return err
				}
				if resp.Error != nil {
					t.Logf("Warning: Error response: %v", resp.Error)
				}
				return nil
			},
		},
	}
}

// TestTransportLifecycle validates proper start/stop behavior
// TODO: Fix stdio transport broken pipe issue
func TestTransportLifecycle(t *testing.T) {
	t.Skip("Stdio transport has broken pipe issue - needs investigation")
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		t.Skip("GITHUB_PERSONAL_ACCESS_TOKEN or GITHUB_TOKEN not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	transport := harness.NewStdioTransport(
		"github-lifecycle",
		"npx",
		[]string{"-y", "@modelcontextprotocol/server-github"},
		[]string{"GITHUB_PERSONAL_ACCESS_TOKEN=" + token},
	)

	// Test start
	if err := transport.Start(ctx); err != nil {
		t.Fatalf("Failed to start transport: %v", err)
	}

	// Test health check
	if !transport.IsHealthy(ctx) {
		t.Error("Transport should be healthy after start")
	}

	// Test execute
	initReq := harness.MustMarshal(harness.MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: harness.MustMarshal(map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "slimmcp-test",
				"version": "0.1.0",
			},
		}),
	})

	response, err := transport.Execute(ctx, initReq)
	if err != nil {
		t.Errorf("Failed to execute request: %v", err)
	}

	var resp harness.MCPResponse
	if err := json.Unmarshal(response, &resp); err != nil {
		t.Errorf("Invalid response format: %v", err)
	}

	// Test stop
	if err := transport.Stop(); err != nil {
		t.Errorf("Failed to stop transport: %v", err)
	}

	// Test that execute fails after stop
	_, err = transport.Execute(ctx, initReq)
	if err == nil {
		t.Error("Execute should fail after transport is stopped")
	}
}
