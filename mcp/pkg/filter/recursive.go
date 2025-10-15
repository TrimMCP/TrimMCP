package filter

import (
	"context"
	"strings"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// FilterNested handles nested field filtering with dot notation
// For example: "properties.Status" keeps only the Status field within properties
func FilterNested(ctx context.Context, data interface{}, fields []string, depth int, maxDepth int) (interface{}, error) {
	// Check context for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check max depth
	if depth > maxDepth {
		return nil, types.ErrMaxDepthExceeded
	}

	// Build nested field map: "properties.Status" -> map["properties"]["Status"] = true
	nestedFields := buildNestedFieldMap(fields)

	return filterWithNestedMap(ctx, data, nestedFields, depth, maxDepth)
}

// buildNestedFieldMap builds a nested map structure from dot notation fields
func buildNestedFieldMap(fields []string) map[string]interface{} {
	result := make(map[string]interface{})

	for _, field := range fields {
		if !strings.Contains(field, ".") {
			// Top-level field
			result[field] = true
			continue
		}

		// Nested field: split and build nested structure
		parts := strings.Split(field, ".")
		current := result

		for i, part := range parts {
			if i == len(parts)-1 {
				// Last part: mark as included
				current[part] = true
			} else {
				// Intermediate part: create nested map if doesn't exist
				if _, ok := current[part]; !ok {
					current[part] = make(map[string]interface{})
				}
				// Move to next level
				if nestedMap, ok := current[part].(map[string]interface{}); ok {
					current = nestedMap
				}
			}
		}
	}

	return result
}

// filterWithNestedMap filters data using a nested field map
func filterWithNestedMap(ctx context.Context, data interface{}, fieldMap map[string]interface{}, depth int, maxDepth int) (interface{}, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		filtered := make(map[string]interface{})

		for key, value := range v {
			// Check if field should be included
			if fieldSpec, ok := fieldMap[key]; ok {
				if fieldSpec == true {
					// Simple inclusion: keep the entire value
					filtered[key] = value
				} else if nestedMap, ok := fieldSpec.(map[string]interface{}); ok {
					// Nested inclusion: recursively filter
					filteredValue, err := filterWithNestedMap(ctx, value, nestedMap, depth+1, maxDepth)
					if err != nil {
						return nil, err
					}
					filtered[key] = filteredValue
				}
			} else if types.IsCriticalField(key) {
				// Always keep critical fields
				filtered[key] = value
			}
		}

		return filtered, nil

	case []interface{}:
		// For arrays, filter each element with the same field map
		filtered := make([]interface{}, len(v))
		for i, elem := range v {
			filteredElem, err := filterWithNestedMap(ctx, elem, fieldMap, depth+1, maxDepth)
			if err != nil {
				return nil, err
			}
			filtered[i] = filteredElem
		}
		return filtered, nil

	default:
		// Primitive values are returned as-is
		return v, nil
	}
}

// FilterWithExclude filters data by excluding specified fields
func FilterWithExclude(ctx context.Context, data interface{}, excludeFields []string, depth int, maxDepth int) (interface{}, error) {
	// Convert exclude list to map for O(1) lookup
	excludeMap := make(map[string]bool)
	for _, field := range excludeFields {
		excludeMap[field] = true
	}

	return filterWithExcludeMap(ctx, data, excludeMap, depth, maxDepth)
}

// filterWithExcludeMap filters data by excluding fields in the map
func filterWithExcludeMap(ctx context.Context, data interface{}, excludeMap map[string]bool, depth int, maxDepth int) (interface{}, error) {
	// Check context for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check max depth
	if depth > maxDepth {
		return nil, types.ErrMaxDepthExceeded
	}

	switch v := data.(type) {
	case map[string]interface{}:
		filtered := make(map[string]interface{})

		for key, value := range v {
			// Include field if NOT in exclude list
			if !excludeMap[key] {
				// Recursively filter nested values
				filteredValue, err := filterWithExcludeMap(ctx, value, excludeMap, depth+1, maxDepth)
				if err != nil {
					return nil, err
				}
				filtered[key] = filteredValue
			}
		}

		return filtered, nil

	case []interface{}:
		filtered := make([]interface{}, len(v))
		for i, elem := range v {
			filteredElem, err := filterWithExcludeMap(ctx, elem, excludeMap, depth+1, maxDepth)
			if err != nil {
				return nil, err
			}
			filtered[i] = filteredElem
		}
		return filtered, nil

	default:
		return v, nil
	}
}
