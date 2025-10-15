package analyzer

import (
	"strings"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// ExplicitAnalyzer implements the explicit filtering strategy
// It extracts fields from _meta.fieldSelection.include/exclude
type ExplicitAnalyzer struct {
	config *Config
}

// Config contains configuration for analyzers
type Config struct {
	AlwaysKeep   []string
	AlwaysRemove []string
}

// NewExplicitAnalyzer creates a new explicit analyzer
func NewExplicitAnalyzer(cfg *Config) *ExplicitAnalyzer {
	if cfg == nil {
		cfg = &Config{
			AlwaysKeep: []string{"id", "type", "url", "title", "object"},
		}
	}
	return &ExplicitAnalyzer{config: cfg}
}

// InferFields extracts fields from the request's field selection specification
func (a *ExplicitAnalyzer) InferFields(req *types.MCPRequest) ([]string, error) {
	if req.Meta == nil || req.Meta.FieldSelection == nil {
		// No field selection specified, return critical fields only
		return a.config.AlwaysKeep, nil
	}

	fieldSelection := req.Meta.FieldSelection

	// Validate field selection
	if err := fieldSelection.ParseFieldSelection(); err != nil {
		return nil, err
	}

	var fields []string

	// Add critical fields (always kept)
	fields = append(fields, a.config.AlwaysKeep...)

	if len(fieldSelection.Include) > 0 {
		// Include mode: add specified fields
		for _, field := range fieldSelection.Include {
			// Add the field and any parent fields if using dot notation
			fields = append(fields, a.expandDotNotation(field)...)
		}
	} else if len(fieldSelection.Exclude) > 0 {
		// Exclude mode: we'll handle this differently in the filter
		// Return empty list to signal "include all except excluded"
		// The filter will use the exclude list from the context
		return []string{}, nil
	} else {
		// No include/exclude specified, return critical fields
		fields = append(fields, a.config.AlwaysKeep...)
	}

	// Remove duplicates
	return removeDuplicates(fields), nil
}

// expandDotNotation expands a dot notation field into parent fields
// For example: "properties.Status" -> ["properties", "properties.Status"]
func (a *ExplicitAnalyzer) expandDotNotation(field string) []string {
	if !strings.Contains(field, ".") {
		return []string{field}
	}

	parts := strings.Split(field, ".")
	expanded := make([]string, 0, len(parts))

	// Add each level of nesting
	for i := range parts {
		expanded = append(expanded, strings.Join(parts[:i+1], "."))
	}

	return expanded
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
