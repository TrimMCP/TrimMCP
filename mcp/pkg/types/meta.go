package types

import "time"

// ToolSchema describes a tool's structure, fields, and validation rules
type ToolSchema struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Fields      map[string]FieldInfo `json:"fields"`
	Required    []string             `json:"required"`
	Defaults    []string             `json:"defaults"` // Default fields for this tool
}

// FieldInfo contains metadata about a specific field in a tool response
type FieldInfo struct {
	Type      string   `json:"type"`      // Field type (string, number, object, array)
	Required  bool     `json:"required"`  // Is this field required?
	RelatedTo []string `json:"relatedTo"` // Related field names
	TokenCost int      `json:"tokenCost"` // Estimated token cost of this field
	Critical  bool     `json:"critical"`  // Is this a critical field (id, url, etc)?
}

// ServerCapabilities contains information about what a downstream MCP server supports
type ServerCapabilities struct {
	URL                    string        `json:"url"`
	SupportsFieldSelection bool          `json:"supportsFieldSelection"`
	Version                string        `json:"version"`
	ProtocolVersion        string        `json:"protocolVersion"`
	SupportedStrategies    []string      `json:"supportedStrategies"`
	LastChecked            time.Time     `json:"lastChecked"`
	CacheTTL               time.Duration `json:"-"` // Not serialized
}

// FilterContext contains combined context for making filtering decisions
type FilterContext struct {
	Request       *MCPRequest
	Response      []byte // Raw response bytes
	Schema        *ToolSchema
	Strategy      string
	KeepFields    map[string]bool // Set of fields to keep
	ExcludeFields map[string]bool // Set of fields to exclude
	MaxDepth      int
	KeepRelated   bool
}

// FieldCategory represents the category of a field for filtering decisions
type FieldCategory int

const (
	FieldCategoryCritical  FieldCategory = iota // Always keep (id, type, url)
	FieldCategoryRequired                       // Required by schema
	FieldCategoryRequested                      // Explicitly requested
	FieldCategoryRelated                        // Related to requested field
	FieldCategoryOptional                       // Can be removed
)

// String returns the string representation of FieldCategory
func (c FieldCategory) String() string {
	switch c {
	case FieldCategoryCritical:
		return "critical"
	case FieldCategoryRequired:
		return "required"
	case FieldCategoryRequested:
		return "requested"
	case FieldCategoryRelated:
		return "related"
	case FieldCategoryOptional:
		return "optional"
	default:
		return "unknown"
	}
}

// IsCriticalField checks if a field name is a critical field
func IsCriticalField(field string) bool {
	criticalFields := map[string]bool{
		"id":    true,
		"type":  true,
		"url":   true,
		"title": true,
		"name":  true,
	}
	return criticalFields[field]
}

// FilteringStatistics contains real-time filtering performance statistics
type FilteringStatistics struct {
	TokensOriginal int           // Original response tokens
	TokensFiltered int           // Filtered response tokens
	TokensSaved    int           // Tokens saved
	SavingsPercent float64       // Savings percentage
	FieldsKept     int           // Number of fields kept
	FieldsRemoved  int           // Number of fields removed
	Duration       time.Duration // Filtering duration
	Strategy       string        // Strategy used
	Server         string        // Server name
	Tool           string        // Tool name
}

// CalculateSavings calculates the savings percentage
func (s *FilteringStatistics) CalculateSavings() {
	if s.TokensOriginal > 0 {
		s.TokensSaved = s.TokensOriginal - s.TokensFiltered
		s.SavingsPercent = float64(s.TokensSaved) / float64(s.TokensOriginal) * 100
	}
}

// NewFilterContext creates a new filter context
func NewFilterContext(req *MCPRequest, response []byte, schema *ToolSchema, strategy string) *FilterContext {
	return &FilterContext{
		Request:       req,
		Response:      response,
		Schema:        schema,
		Strategy:      strategy,
		KeepFields:    make(map[string]bool),
		ExcludeFields: make(map[string]bool),
		MaxDepth:      10,
		KeepRelated:   true,
	}
}

// ShouldKeepField determines if a field should be kept based on the context
func (c *FilterContext) ShouldKeepField(field string) bool {
	// Critical fields always kept
	if IsCriticalField(field) {
		return true
	}

	// Explicit exclude list
	if c.ExcludeFields[field] {
		return false
	}

	// Explicit include list
	if len(c.KeepFields) > 0 {
		return c.KeepFields[field]
	}

	// If schema available, check if required
	if c.Schema != nil {
		for _, required := range c.Schema.Required {
			if required == field {
				return true
			}
		}
	}

	// Default: keep unless explicitly excluded
	return true
}

// AddKeepField adds a field to the keep list
func (c *FilterContext) AddKeepField(field string) {
	c.KeepFields[field] = true
}

// AddExcludeField adds a field to the exclude list
func (c *FilterContext) AddExcludeField(field string) {
	c.ExcludeFields[field] = true
}
