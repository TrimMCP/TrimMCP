package schema

import "github.com/slimmcp/slimmcp/pkg/types"

// Manager defines the interface for managing tool schemas
type Manager interface {
	// GetSchema retrieves the schema for a specific tool from a server
	GetSchema(tool, serverURL string) (*types.ToolSchema, error)

	// LoadSchema loads a schema from configuration or cache
	LoadSchema(tool string) (*types.ToolSchema, error)

	// CacheSchema stores a schema in the cache
	CacheSchema(tool, serverURL string, schema *types.ToolSchema) error

	// InvalidateCache clears the schema cache for a specific tool or all tools
	InvalidateCache(tool string) error
}
