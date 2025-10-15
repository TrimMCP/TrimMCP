package cache

import "time"

// CachedResponse represents a stored MCP response.
type CachedResponse struct {
	Operation  string                 `json:"operation"`
	Server     string                 `json:"server"`
	Request    map[string]interface{} `json:"request"`
	Response   map[string]interface{} `json:"response"`
	CapturedAt time.Time              `json:"captured_at"`
}

// CacheLoader loads cached MCP responses from storage.
type CacheLoader interface {
	// Load retrieves a cached response for the given server and operation.
	// Returns error if cache file missing or corrupted.
	Load(server, operation string) (*CachedResponse, error)

	// Exists checks if a cached response exists without loading it.
	// Used for validation before attempting to load.
	Exists(server, operation string) bool
}
