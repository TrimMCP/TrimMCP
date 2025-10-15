package filter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// Engine implements the Filter interface for filtering response data
type Engine struct {
	maxDepth int
}

// NewEngine creates a new filtering engine
func NewEngine() *Engine {
	return &Engine{
		maxDepth: 10, // Default max depth
	}
}

// Filter applies filtering to the given data based on the specified fields
func (e *Engine) Filter(ctx context.Context, data interface{}, fields []string) (interface{}, error) {
	// Convert fields slice to map for O(1) lookup
	keepFields := make(map[string]bool)
	for _, field := range fields {
		keepFields[field] = true
	}

	// Filter the data recursively
	filtered, err := e.filterValue(ctx, data, keepFields, 0)
	if err != nil {
		return nil, fmt.Errorf("filtering failed: %w", err)
	}

	return filtered, nil
}

// filterValue filters a value recursively
func (e *Engine) filterValue(ctx context.Context, value interface{}, keepFields map[string]bool, depth int) (interface{}, error) {
	// Check context for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check max depth
	if depth > e.maxDepth {
		return nil, types.ErrMaxDepthExceeded
	}

	// Handle different value types
	switch v := value.(type) {
	case map[string]interface{}:
		return e.filterObject(ctx, v, keepFields, depth)
	case []interface{}:
		return e.filterArray(ctx, v, keepFields, depth)
	default:
		// Primitive values are returned as-is
		return v, nil
	}
}

// filterObject filters an object (map) keeping only specified fields
func (e *Engine) filterObject(ctx context.Context, obj map[string]interface{}, keepFields map[string]bool, depth int) (map[string]interface{}, error) {
	filtered := make(map[string]interface{})

	for key, value := range obj {
		// Keep field if it's in the keep list or if list is empty (keep all)
		if len(keepFields) == 0 || keepFields[key] || types.IsCriticalField(key) {
			// Recursively filter nested values
			filteredValue, err := e.filterValue(ctx, value, keepFields, depth+1)
			if err != nil {
				return nil, err
			}
			filtered[key] = filteredValue
		}
	}

	return filtered, nil
}

// filterArray filters an array, recursively filtering each element
func (e *Engine) filterArray(ctx context.Context, arr []interface{}, keepFields map[string]bool, depth int) ([]interface{}, error) {
	filtered := make([]interface{}, len(arr))

	for i, elem := range arr {
		filteredElem, err := e.filterValue(ctx, elem, keepFields, depth+1)
		if err != nil {
			return nil, err
		}
		filtered[i] = filteredElem
	}

	return filtered, nil
}

// FilterJSON filters raw JSON data
func (e *Engine) FilterJSON(ctx context.Context, data []byte, fields []string) ([]byte, error) {
	// Parse JSON
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Filter
	filtered, err := e.Filter(ctx, parsed, fields)
	if err != nil {
		return nil, err
	}

	// Marshal back to JSON
	result, err := json.Marshal(filtered)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filtered data: %w", err)
	}

	return result, nil
}

// SetMaxDepth sets the maximum nesting depth for filtering
func (e *Engine) SetMaxDepth(depth int) {
	e.maxDepth = depth
}
