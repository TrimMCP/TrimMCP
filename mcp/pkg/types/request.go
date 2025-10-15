package types

import "encoding/json"

// MCPRequest represents an incoming JSON-RPC 2.0 request from an LLM client
type MCPRequest struct {
	JSONRPC string         `json:"jsonrpc"`          // Always "2.0"
	Method  string         `json:"method"`           // MCP method (e.g., "tools/call")
	Params  *RequestParams `json:"params,omitempty"` // Method-specific parameters
	ID      interface{}    `json:"id,omitempty"`     // Request identifier (string, number, or null)
	Meta    *RequestMeta   `json:"_meta,omitempty"`  // SlimMCP extension
}

// RequestParams contains method-specific parameters for tool invocation
type RequestParams struct {
	Name      string                 `json:"name"`      // Tool name (e.g., "notion-fetch")
	Arguments map[string]interface{} `json:"arguments"` // Tool-specific arguments
}

// RequestMeta contains SlimMCP-specific metadata for request processing
type RequestMeta struct {
	FieldSelection *FieldSelectionSpec `json:"fieldSelection,omitempty"`
}

// FieldSelectionSpec specifies which fields to include/exclude in responses
type FieldSelectionSpec struct {
	Include     []string `json:"include,omitempty"`     // Fields to explicitly include
	Exclude     []string `json:"exclude,omitempty"`     // Fields to explicitly exclude
	Strategy    string   `json:"strategy,omitempty"`    // Filtering strategy override
	KeepRelated bool     `json:"keepRelated,omitempty"` // Include related fields (default: true)
	MaxDepth    int      `json:"maxDepth,omitempty"`    // Maximum nesting depth (default: 10, max: 20)
}

// ParseFieldSelection validates and normalizes the field selection spec
func (f *FieldSelectionSpec) ParseFieldSelection() error {
	// Set defaults
	if f.MaxDepth == 0 {
		f.MaxDepth = 10
	}

	// Validate constraints
	if len(f.Include) > 0 && len(f.Exclude) > 0 {
		return ErrInvalidFieldSelection
	}

	if f.MaxDepth < 1 || f.MaxDepth > 20 {
		return ErrInvalidMaxDepth
	}

	// Validate strategy if specified
	if f.Strategy != "" {
		validStrategies := map[string]bool{
			"smart":       true,
			"minimal":     true,
			"explicit":    true,
			"passthrough": true,
		}
		if !validStrategies[f.Strategy] {
			return ErrInvalidStrategy
		}
	}

	return nil
}

// HasInclude returns true if include list is specified
func (f *FieldSelectionSpec) HasInclude() bool {
	return len(f.Include) > 0
}

// HasExclude returns true if exclude list is specified
func (f *FieldSelectionSpec) HasExclude() bool {
	return len(f.Exclude) > 0
}

// IsFieldIncluded checks if a field should be included based on the spec
func (f *FieldSelectionSpec) IsFieldIncluded(field string) bool {
	if f.HasInclude() {
		for _, included := range f.Include {
			if included == field {
				return true
			}
		}
		return false
	}

	if f.HasExclude() {
		for _, excluded := range f.Exclude {
			if excluded == field {
				return false
			}
		}
		return true
	}

	return true
}

// ToJSON converts the request to JSON bytes
func (r *MCPRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON parses JSON bytes into an MCPRequest
func (r *MCPRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// Validate checks if the request is valid
func (r *MCPRequest) Validate() error {
	if r.JSONRPC != "2.0" {
		return ErrInvalidJSONRPC
	}

	if r.Method == "" {
		return ErrMissingMethod
	}

	// ID is required for requests expecting a response (not notifications)
	// Notifications have ID == nil

	if r.Meta != nil && r.Meta.FieldSelection != nil {
		if err := r.Meta.FieldSelection.ParseFieldSelection(); err != nil {
			return err
		}
	}

	return nil
}

// IsNotification returns true if this is a notification (no response expected)
func (r *MCPRequest) IsNotification() bool {
	return r.ID == nil
}

// GetToolName returns the tool name from params if available
func (r *MCPRequest) GetToolName() string {
	if r.Params != nil {
		return r.Params.Name
	}
	return ""
}

// GetArguments returns the arguments map from params if available
func (r *MCPRequest) GetArguments() map[string]interface{} {
	if r.Params != nil {
		return r.Params.Arguments
	}
	return nil
}
