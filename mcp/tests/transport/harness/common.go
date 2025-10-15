package harness

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// Transport defines the interface for MCP transport adapters
type Transport interface {
	// Name returns the transport type (stdio, http, https, sse)
	Name() string

	// Start initializes the transport and any required backend servers
	Start(ctx context.Context) error

	// Stop cleanly shuts down the transport
	Stop() error

	// Execute sends an MCP request and returns the response
	Execute(ctx context.Context, request json.RawMessage) (json.RawMessage, error)

	// IsHealthy checks if the transport is ready to handle requests
	IsHealthy(ctx context.Context) bool
}

// TestScenario represents a reusable test case across all transports
type TestScenario struct {
	Name        string          // Scenario name (e.g., "list_tools")
	Description string          // Human-readable description
	Request     json.RawMessage // MCP JSON-RPC request
	Timeout     time.Duration   // Max execution time (default 5s)

	// ValidateResponse is called with the raw response to validate transport-agnostic behavior
	ValidateResponse func(t *testing.T, response json.RawMessage) error
}

// MCPRequest represents a generic MCP JSON-RPC request
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents a generic MCP JSON-RPC response
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC error object
type MCPError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// ResponseNormalizer removes transport-specific variations for comparison
type ResponseNormalizer struct {
	// IgnoreFields contains JSON paths to ignore during comparison (e.g., "result.timing")
	IgnoreFields []string

	// SortArrays indicates whether to sort arrays before comparison
	SortArrays bool
}

// Normalize applies normalization rules to an MCP response
func (n *ResponseNormalizer) Normalize(response json.RawMessage) (json.RawMessage, error) {
	// Parse response
	var data map[string]interface{}
	if err := json.Unmarshal(response, &data); err != nil {
		return nil, err
	}

	// Remove ignored fields
	for _, field := range n.IgnoreFields {
		removeJSONPath(data, field)
	}

	// Re-marshal
	return json.Marshal(data)
}

// AssertEquivalentResponses compares responses across transports
func AssertEquivalentResponses(t *testing.T, normalizer *ResponseNormalizer, responses map[string]json.RawMessage) {
	t.Helper()

	if len(responses) < 2 {
		t.Fatal("Need at least 2 responses to compare")
	}

	// Normalize all responses
	normalized := make(map[string]json.RawMessage)
	for transport, response := range responses {
		norm, err := normalizer.Normalize(response)
		if err != nil {
			t.Errorf("%s: failed to normalize response: %v", transport, err)
			continue
		}
		normalized[transport] = norm
	}

	// Compare each transport against the first one (baseline)
	var baseline string
	var baselineResponse json.RawMessage
	for transport, response := range normalized {
		baseline = transport
		baselineResponse = response
		break
	}

	for transport, response := range normalized {
		if transport == baseline {
			continue
		}

		if diff := cmp.Diff(string(baselineResponse), string(response)); diff != "" {
			t.Errorf("Response mismatch between %s and %s (-want +got):\n%s", baseline, transport, diff)
		}
	}
}

// DefaultNormalizer provides sensible defaults for response comparison
func DefaultNormalizer() *ResponseNormalizer {
	return &ResponseNormalizer{
		IgnoreFields: []string{
			// Common transport-specific metadata that shouldn't affect equivalence
			"_meta.timing",
			"_meta.transport",
			"_meta.server_version",
		},
		SortArrays: true,
	}
}

// removeJSONPath removes a nested field from a map using dot notation
func removeJSONPath(data map[string]interface{}, path string) {
	// Simple implementation - can be enhanced for array indices
	if len(path) == 0 {
		return
	}

	// Split on first dot
	parts := splitFirst(path, ".")
	if len(parts) == 1 {
		delete(data, parts[0])
		return
	}

	// Recursively navigate
	if nested, ok := data[parts[0]].(map[string]interface{}); ok {
		removeJSONPath(nested, parts[1])
	}
}

// splitFirst splits a string on the first occurrence of sep
func splitFirst(s, sep string) []string {
	idx := -1
	for i := 0; i < len(s)-len(sep)+1; i++ {
		if s[i:i+len(sep)] == sep {
			idx = i
			break
		}
	}
	if idx == -1 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+len(sep):]}
}

// CreateBasicScenarios returns common test scenarios for all transports
func CreateBasicScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:        "initialize",
			Description: "MCP initialize handshake",
			Request: MustMarshal(MCPRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "initialize",
				Params: MustMarshal(map[string]interface{}{
					"protocolVersion": "2024-11-05",
					"capabilities":    map[string]interface{}{},
					"clientInfo": map[string]interface{}{
						"name":    "slimmcp-test",
						"version": "0.1.0",
					},
				}),
			}),
			Timeout: 10 * time.Second,
			ValidateResponse: func(t *testing.T, response json.RawMessage) error {
				var resp MCPResponse
				if err := json.Unmarshal(response, &resp); err != nil {
					return err
				}
				if resp.Error != nil {
					t.Errorf("Expected success, got error: %v", resp.Error)
				}
				return nil
			},
		},
		{
			Name:        "list_tools",
			Description: "List available MCP tools",
			Request: MustMarshal(MCPRequest{
				JSONRPC: "2.0",
				ID:      2,
				Method:  "tools/list",
			}),
			Timeout: 5 * time.Second,
			ValidateResponse: func(t *testing.T, response json.RawMessage) error {
				var resp MCPResponse
				if err := json.Unmarshal(response, &resp); err != nil {
					return err
				}
				if resp.Error != nil {
					t.Errorf("Expected success, got error: %v", resp.Error)
				}
				return nil
			},
		},
	}
}

// MustMarshal marshals a value to JSON or panics on error
func MustMarshal(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
