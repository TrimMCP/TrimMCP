package analyzer

import (
	"fmt"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// NewAnalyzer creates an analyzer based on the strategy
func NewAnalyzer(strategy string, cfg *Config) (Analyzer, error) {
	switch strategy {
	case "minimal":
		return NewMinimalAnalyzer(), nil
	case "explicit":
		return NewExplicitAnalyzer(cfg), nil
	case "smart":
		// Smart analyzer not yet implemented (Phase 8)
		return NewMinimalAnalyzer(), nil // Fallback to minimal for now
	case "passthrough":
		// Passthrough analyzer not yet implemented (Phase 7 - User Story 5)
		return NewPassthroughAnalyzer(), nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategy)
	}
}

// NewAnalyzerFromRequest creates an analyzer based on request metadata or config
func NewAnalyzerFromRequest(req *types.MCPRequest, defaultStrategy string, cfg *Config) (Analyzer, error) {
	strategy := defaultStrategy

	// Check if request specifies a strategy override
	if req.Meta != nil && req.Meta.FieldSelection != nil && req.Meta.FieldSelection.Strategy != "" {
		strategy = req.Meta.FieldSelection.Strategy
	}

	return NewAnalyzer(strategy, cfg)
}

// PassthroughAnalyzer is a placeholder for when no filtering should be applied
type PassthroughAnalyzer struct{}

// NewPassthroughAnalyzer creates a new passthrough analyzer
func NewPassthroughAnalyzer() *PassthroughAnalyzer {
	return &PassthroughAnalyzer{}
}

// InferFields returns empty list (no filtering)
func (a *PassthroughAnalyzer) InferFields(req *types.MCPRequest) ([]string, error) {
	return []string{}, nil
}
