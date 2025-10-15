package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// JSONFileLoader implements CacheLoader by reading from JSON files in testdata/.
type JSONFileLoader struct {
	// BaseDir is the root directory for cached responses (default: "testdata")
	BaseDir string
}

// NewJSONFileLoader creates a new JSONFileLoader with the specified base directory.
// If baseDir is empty, defaults to "testdata".
func NewJSONFileLoader(baseDir string) *JSONFileLoader {
	if baseDir == "" {
		baseDir = "testdata"
	}

	return &JSONFileLoader{
		BaseDir: baseDir,
	}
}

// Load retrieves a cached response from testdata/{server}/{operation}.json.
// Returns error if file is missing or corrupted.
func (j *JSONFileLoader) Load(server, operation string) (*CachedResponse, error) {
	filePath := j.getFilePath(server, operation)

	// Check if file exists
	if !j.Exists(server, operation) {
		return nil, fmt.Errorf("cache file not found: %s", filePath)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file %s: %w", filePath, err)
	}

	// Parse JSON
	var cached CachedResponse
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("failed to parse cache file %s: %w", filePath, err)
	}

	// Validate structure
	if cached.Operation == "" {
		return nil, fmt.Errorf("cache file %s missing 'operation' field", filePath)
	}
	if cached.Server == "" {
		return nil, fmt.Errorf("cache file %s missing 'server' field", filePath)
	}
	if cached.Response == nil {
		return nil, fmt.Errorf("cache file %s missing 'response' field", filePath)
	}

	return &cached, nil
}

// Exists checks if a cached response file exists without loading it.
func (j *JSONFileLoader) Exists(server, operation string) bool {
	filePath := j.getFilePath(server, operation)
	_, err := os.Stat(filePath)
	return err == nil
}

// getFilePath constructs the file path for a cached response.
// Format: {BaseDir}/{server}/{operation}.json
func (j *JSONFileLoader) getFilePath(server, operation string) string {
	return filepath.Join(j.BaseDir, server, operation+".json")
}
