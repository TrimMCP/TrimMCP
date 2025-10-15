package filter

import (
	"encoding/json"
	"time"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// EstimateTokens estimates the number of tokens in data using character count / 4
// This is a rough approximation commonly used for OpenAI models
func EstimateTokens(data interface{}) int {
	// Marshal to JSON to get character count
	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0
	}

	// Approximate: 1 token ≈ 4 characters
	return len(jsonData) / 4
}

// CalculateFilteringStats calculates filtering statistics by comparing original and filtered data
func CalculateFilteringStats(original, filtered interface{}, strategy string, duration time.Duration) types.FilteringMetadata {
	// Estimate tokens
	originalSize := EstimateTokens(original)
	filteredSize := EstimateTokens(filtered)

	// Calculate savings
	tokensSaved := originalSize - filteredSize
	savingsPercent := 0.0
	if originalSize > 0 {
		savingsPercent = float64(tokensSaved) / float64(originalSize) * 100
	}

	// Count fields kept and removed
	fieldsKept, fieldsRemoved := countFields(original, filtered)

	return types.FilteringMetadata{
		Applied:       true,
		Strategy:      strategy,
		FilteredBy:    "proxy",
		OriginalSize:  originalSize,
		FilteredSize:  filteredSize,
		Savings:       savingsPercent,
		FieldsKept:    fieldsKept,
		FieldsRemoved: fieldsRemoved,
		FilteringTime: float64(duration.Milliseconds()),
	}
}

// countFields counts the fields kept and removed
func countFields(original, filtered interface{}) ([]string, int) {
	originalFields := extractFieldNames(original)
	filteredFields := extractFieldNames(filtered)

	// Create map for quick lookup
	filteredMap := make(map[string]bool)
	for _, field := range filteredFields {
		filteredMap[field] = true
	}

	// Count removed fields
	removed := 0
	for _, field := range originalFields {
		if !filteredMap[field] {
			removed++
		}
	}

	return filteredFields, removed
}

// extractFieldNames recursively extracts all field names from data
func extractFieldNames(data interface{}) []string {
	var fields []string

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fields = append(fields, key)
			// Recursively get nested fields
			nestedFields := extractFieldNames(value)
			fields = append(fields, nestedFields...)
		}
	case []interface{}:
		// For arrays, extract fields from first element (representative)
		if len(v) > 0 {
			fields = extractFieldNames(v[0])
		}
	}

	return fields
}

// CreateNoFilteringMetadata creates metadata for when no filtering was applied
func CreateNoFilteringMetadata() types.FilteringMetadata {
	return types.FilteringMetadata{
		Applied:       false,
		Strategy:      "passthrough",
		FilteredBy:    "none",
		OriginalSize:  0,
		FilteredSize:  0,
		Savings:       0.0,
		FieldsKept:    []string{},
		FieldsRemoved: 0,
		FilteringTime: 0.0,
	}
}
