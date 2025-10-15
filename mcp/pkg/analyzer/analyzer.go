package analyzer

import "github.com/slimmcp/slimmcp/pkg/types"

// Analyzer defines the interface for analyzing requests and inferring needed fields
type Analyzer interface {
	// InferFields analyzes a request and returns the list of fields that should be kept
	InferFields(req *types.MCPRequest) ([]string, error)
}
