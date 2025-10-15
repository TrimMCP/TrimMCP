package capability

import (
	"context"
	"encoding/json"
	"time"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// Detector defines the interface for detecting server capabilities
type Detector interface {
	// DetectCapabilities queries a server for its capabilities
	DetectCapabilities(ctx context.Context, serverURL string) (*types.ServerCapabilities, error)
}

// MCPCapabilityDetector detects MCP server capabilities
type MCPCapabilityDetector struct {
	client HTTPClient
}

// HTTPClient interface for making HTTP requests (for testing)
type HTTPClient interface {
	Forward(ctx context.Context, serverURL string, request *types.MCPRequest, headers map[string]string) ([]byte, error)
}

// NewDetector creates a new capability detector
func NewDetector(client HTTPClient) *MCPCapabilityDetector {
	return &MCPCapabilityDetector{
		client: client,
	}
}

// DetectCapabilities queries the server for its capabilities
func (d *MCPCapabilityDetector) DetectCapabilities(ctx context.Context, serverURL string) (*types.ServerCapabilities, error) {
	// Try to query server capabilities via MCP protocol
	// First, try a tools/list request to see if server responds
	request := &types.MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/list",
		Params:  nil,
		ID:      "capability-check",
	}

	// Send request with short timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	responseBody, err := d.client.Forward(timeoutCtx, serverURL, request, nil)
	if err != nil {
		// If server doesn't respond, assume no capabilities
		return &types.ServerCapabilities{
			URL:                    serverURL,
			SupportsFieldSelection: false,
			Version:                "unknown",
			ProtocolVersion:        "unknown",
			SupportedStrategies:    []string{},
			LastChecked:            time.Now(),
			CacheTTL:               1 * time.Hour,
		}, nil
	}

	// Parse response
	var response types.MCPResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		// Can't parse response, assume no capabilities
		return &types.ServerCapabilities{
			URL:                    serverURL,
			SupportsFieldSelection: false,
			Version:                "unknown",
			ProtocolVersion:        "unknown",
			SupportedStrategies:    []string{},
			LastChecked:            time.Now(),
			CacheTTL:               1 * time.Hour,
		}, nil
	}

	// If response has error, assume no capabilities
	if response.Error != nil {
		return &types.ServerCapabilities{
			URL:                    serverURL,
			SupportsFieldSelection: false,
			Version:                "unknown",
			ProtocolVersion:        "unknown",
			SupportedStrategies:    []string{},
			LastChecked:            time.Now(),
			CacheTTL:               1 * time.Hour,
		}, nil
	}

	// Check if response metadata indicates field selection support
	supportsFieldSelection := false
	version := "unknown"
	protocolVersion := "unknown"
	supportedStrategies := []string{}

	if response.Meta != nil {
		// Check for field selection capability in metadata
		if response.Meta.Version != "" {
			version = response.Meta.Version
		}
		if response.Meta.ProtocolVersion != "" {
			protocolVersion = response.Meta.ProtocolVersion
		}

		// Parse capabilities from metadata
		if response.Meta.Capabilities != nil {
			// Check for fieldSelection capability
			if fsCapRaw, ok := response.Meta.Capabilities["fieldSelection"]; ok {
				if fsCapMap, ok := fsCapRaw.(map[string]interface{}); ok {
					if supported, ok := fsCapMap["supported"].(bool); ok {
						supportsFieldSelection = supported
					}
					// Parse supported strategies if available
					if strategiesRaw, ok := fsCapMap["strategies"]; ok {
						if strategiesList, ok := strategiesRaw.([]interface{}); ok {
							for _, s := range strategiesList {
								if str, ok := s.(string); ok {
									supportedStrategies = append(supportedStrategies, str)
								}
							}
						}
					}
				}
			}
		}
	}

	return &types.ServerCapabilities{
		URL:                    serverURL,
		SupportsFieldSelection: supportsFieldSelection,
		Version:                version,
		ProtocolVersion:        protocolVersion,
		SupportedStrategies:    supportedStrategies,
		LastChecked:            time.Now(),
		CacheTTL:               1 * time.Hour,
	}, nil
}

// detectFromToolsList attempts to detect capabilities from tools/list response
func detectFromToolsList(result json.RawMessage) (bool, []string) {
	// Parse tools list to see if any tools advertise field selection support
	var toolsList struct {
		Tools []struct {
			Name         string                 `json:"name"`
			Capabilities map[string]interface{} `json:"capabilities,omitempty"`
		} `json:"tools"`
	}

	if err := json.Unmarshal(result, &toolsList); err != nil {
		return false, []string{}
	}

	// Check if any tools advertise field selection capability
	for _, tool := range toolsList.Tools {
		if cap, ok := tool.Capabilities["fieldSelection"]; ok {
			if supported, ok := cap.(bool); ok && supported {
				return true, []string{"explicit", "minimal"}
			}
		}
	}

	return false, []string{}
}

// NoOpDetector is a detector that always returns no capabilities
type NoOpDetector struct{}

// NewNoOpDetector creates a no-op detector
func NewNoOpDetector() *NoOpDetector {
	return &NoOpDetector{}
}

// DetectCapabilities always returns no capabilities
func (d *NoOpDetector) DetectCapabilities(ctx context.Context, serverURL string) (*types.ServerCapabilities, error) {
	return &types.ServerCapabilities{
		URL:                    serverURL,
		SupportsFieldSelection: false,
		Version:                "unknown",
		ProtocolVersion:        "unknown",
		SupportedStrategies:    []string{},
		LastChecked:            time.Now(),
		CacheTTL:               1 * time.Hour,
	}, nil
}
