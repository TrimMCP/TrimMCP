package types

import "encoding/json"

// MCPResponse represents a JSON-RPC 2.0 response to the client
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`          // Always "2.0"
	Result  json.RawMessage `json:"result,omitempty"` // Successful response data
	Error   *JSONRPCError   `json:"error,omitempty"`  // Error information if request failed
	ID      interface{}     `json:"id"`               // Request ID for correlation
	Meta    *ResponseMeta   `json:"_meta,omitempty"`  // SlimMCP-specific metadata
}

// JSONRPCError represents a standard JSON-RPC 2.0 error structure
type JSONRPCError struct {
	Code    int             `json:"code"`           // Error code
	Message string          `json:"message"`        // Error message
	Data    json.RawMessage `json:"data,omitempty"` // Additional error data
}

// ResponseMeta contains SlimMCP-specific response metadata
type ResponseMeta struct {
	FieldSelection  *FilteringMetadata     `json:"fieldSelection,omitempty"`  // Filtering statistics
	Version         string                 `json:"version,omitempty"`         // SlimMCP version
	ProtocolVersion string                 `json:"protocolVersion,omitempty"` // MCP protocol version
	Capabilities    map[string]interface{} `json:"capabilities,omitempty"`    // Server capabilities
}

// FilteringMetadata contains statistics and information about response filtering
type FilteringMetadata struct {
	Applied       bool     `json:"applied"`        // Was filtering applied?
	Strategy      string   `json:"strategy"`       // Strategy used
	FilteredBy    string   `json:"filteredBy"`     // "proxy" or "server"
	OriginalSize  int      `json:"originalSize"`   // Original response size (tokens)
	FilteredSize  int      `json:"filteredSize"`   // Filtered response size (tokens)
	Savings       float64  `json:"savings"`        // Savings percentage
	FieldsKept    []string `json:"fields.kept"`    // Fields that were kept
	FieldsRemoved int      `json:"fields.removed"` // Number of fields removed
	FilteringTime float64  `json:"filteringTime"`  // Time spent filtering (ms)
}

// ToJSON converts the response to JSON bytes
func (r *MCPResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON parses JSON bytes into an MCPResponse
func (r *MCPResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// Validate checks if the response is valid
func (r *MCPResponse) Validate() error {
	if r.JSONRPC != "2.0" {
		return ErrInvalidJSONRPC
	}

	// Exactly one of Result or Error must be present
	hasResult := len(r.Result) > 0
	hasError := r.Error != nil

	if hasResult && hasError {
		return ErrBothResultAndError
	}

	if !hasResult && !hasError {
		return ErrNeitherResultNorError
	}

	return nil
}

// IsError returns true if this response contains an error
func (r *MCPResponse) IsError() bool {
	return r.Error != nil
}

// IsSuccess returns true if this response contains a result
func (r *MCPResponse) IsSuccess() bool {
	return len(r.Result) > 0 && r.Error == nil
}

// Standard JSON-RPC 2.0 error codes
const (
	ErrorCodeParseError     = -32700 // Invalid JSON
	ErrorCodeInvalidRequest = -32600 // Invalid Request
	ErrorCodeMethodNotFound = -32601 // Method not found
	ErrorCodeInvalidParams  = -32602 // Invalid params
	ErrorCodeInternalError  = -32603 // Internal error
)

// Custom SlimMCP error codes
const (
	ErrorCodeDownstreamError = -32000 // Downstream server error
	ErrorCodeFilteringError  = -32001 // Filtering error
	ErrorCodeSchemaError     = -32002 // Schema validation error
	ErrorCodeTimeoutError    = -32003 // Timeout error
)

// NewJSONRPCError creates a new JSON-RPC error
func NewJSONRPCError(code int, message string, data interface{}) *JSONRPCError {
	var dataJSON json.RawMessage
	if data != nil {
		if jsonData, err := json.Marshal(data); err == nil {
			dataJSON = jsonData
		}
	}

	return &JSONRPCError{
		Code:    code,
		Message: message,
		Data:    dataJSON,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(id interface{}, code int, message string, data interface{}) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		Error:   NewJSONRPCError(code, message, data),
		ID:      id,
	}
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(id interface{}, result json.RawMessage) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
}
