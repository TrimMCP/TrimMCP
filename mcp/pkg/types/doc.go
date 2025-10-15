// Package types provides common type definitions for SlimMCP.
//
// This package defines the core data structures used throughout SlimMCP,
// including requests, responses, field selection metadata, and error types.
//
// # Overview
//
// The types package serves as a shared foundation for all SlimMCP packages,
// ensuring consistent data structures and interfaces across the codebase.
//
// # Core Types
//
// # Request Types
//
// MCPRequest represents a JSON-RPC 2.0 MCP request:
//
//	type MCPRequest struct {
//		JSONRPC string          `json:"jsonrpc"`
//		ID      interface{}     `json:"id"`
//		Method  string          `json:"method"`
//		Params  json.RawMessage `json:"params"`
//		Meta    *RequestMeta    `json:"_meta,omitempty"`
//	}
//
// # Response Types
//
// MCPResponse represents a JSON-RPC 2.0 MCP response:
//
//	type MCPResponse struct {
//		JSONRPC string          `json:"jsonrpc"`
//		ID      interface{}     `json:"id"`
//		Result  json.RawMessage `json:"result,omitempty"`
//		Error   *MCPError       `json:"error,omitempty"`
//		Meta    *ResponseMeta   `json:"_meta,omitempty"`
//	}
//
// # Field Selection
//
// FieldSelection specifies which fields to keep or remove:
//
//	type FieldSelection struct {
//		Include     []string `json:"include,omitempty"`
//		Exclude     []string `json:"exclude,omitempty"`
//		Strategy    Strategy `json:"strategy"`
//		KeepRelated bool     `json:"keepRelated,omitempty"`
//	}
//
// Example usage:
//
//	selection := &FieldSelection{
//		Include:  []string{"id", "title", "status"},
//		Strategy: StrategyExplicit,
//	}
//
// # Metadata Types
//
// ResponseMeta contains filtering metadata attached to responses:
//
//	type ResponseMeta struct {
//		FieldSelection *FieldSelectionMeta `json:"fieldSelection,omitempty"`
//		Timestamp      time.Time           `json:"timestamp"`
//		Version        string              `json:"version"`
//	}
//
// FieldSelectionMeta contains detailed filtering statistics:
//
//	type FieldSelectionMeta struct {
//		Applied       bool     `json:"applied"`
//		Strategy      Strategy `json:"strategy"`
//		FilteredBy    string   `json:"filteredBy"`
//		OriginalSize  int      `json:"originalSize"`
//		FilteredSize  int      `json:"filteredSize"`
//		Savings       float64  `json:"savings"`
//		FieldsKept    []string `json:"fieldsKept,omitempty"`
//		FieldsRemoved int      `json:"fieldsRemoved"`
//		FilteringTime float64  `json:"filteringTime"`
//	}
//
// # Error Types
//
// MCPError represents a JSON-RPC 2.0 error:
//
//	type MCPError struct {
//		Code    int             `json:"code"`
//		Message string          `json:"message"`
//		Data    json.RawMessage `json:"data,omitempty"`
//	}
//
// Standard error codes:
//
//   - -32700: Parse error
//   - -32600: Invalid request
//   - -32601: Method not found
//   - -32602: Invalid params
//   - -32603: Internal error
//
// # Strategy Types
//
// Strategy defines the filtering approach:
//
//	type Strategy string
//
//	const (
//		StrategyMinimal  Strategy = "minimal"
//		StrategySmart    Strategy = "smart"
//		StrategyExplicit Strategy = "explicit"
//		StrategyLearning Strategy = "learning"
//	)
//
// # Validation
//
// All types include validation methods:
//
//	if err := request.Validate(); err != nil {
//		log.Error("Invalid request:", err)
//	}
//
// # JSON Marshaling
//
// All types support standard JSON encoding/decoding:
//
//	data, err := json.Marshal(response)
//	if err != nil {
//		log.Error("Marshal error:", err)
//	}
//
// # Thread Safety
//
// Type instances are safe for concurrent read access. Modifications require
// external synchronization.
package types
