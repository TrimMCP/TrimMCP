package analyzer

import "github.com/slimmcp/slimmcp/pkg/types"

// MinimalAnalyzer implements the minimal filtering strategy
// It returns a fixed list of critical fields: id, type, url, title
type MinimalAnalyzer struct{}

// NewMinimalAnalyzer creates a new minimal analyzer
func NewMinimalAnalyzer() *MinimalAnalyzer {
	return &MinimalAnalyzer{}
}

// InferFields returns the fixed list of critical fields for minimal strategy
func (a *MinimalAnalyzer) InferFields(req *types.MCPRequest) ([]string, error) {
	// Minimal strategy always returns the same critical fields
	return []string{"id", "type", "url", "title", "object"}, nil
}
