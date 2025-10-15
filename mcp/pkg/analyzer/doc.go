// Package analyzer provides intelligent field inference for MCP responses.
//
// The analyzer examines request parameters and context to determine which
// response fields are likely to be used, enabling smart filtering decisions.
//
// # Overview
//
// This package implements the "smart" filtering strategy by analyzing:
//
//   - Request parameters (which fields are mentioned)
//   - Tool/method being called
//   - Historical usage patterns (planned)
//   - Common field combinations
//
// # Basic Usage
//
//	analyzer := analyzer.NewSmartAnalyzer()
//
//	fields, confidence := analyzer.InferFields(request)
//	fmt.Printf("Inferred fields: %v (confidence: %.2f)\n", fields, confidence)
//
// # Analyzer Types
//
// # MinimalAnalyzer
//
// Returns a fixed set of critical fields for all requests:
//
//   - id, type, url, title, object
//
// # ExplicitAnalyzer
//
// Uses client-provided field selection from request metadata.
//
// # SmartAnalyzer (planned for v0.2)
//
// Analyzes request context to infer likely needed fields:
//
//   - Parameter inspection (e.g., "status" param suggests keeping status field)
//   - Tool-specific patterns (e.g., notion-search typically needs title, url)
//   - Recent access patterns (learning mode)
//
// # Field Inference Rules
//
// The smart analyzer applies these heuristics:
//
//  1. Always include critical fields (id, type, url)
//  2. Include fields mentioned in request parameters
//  3. Include common fields for the specific tool
//  4. Include related fields (e.g., if keeping "assignee", also keep "assignee.name")
//
// # Confidence Scores
//
// Inference results include confidence scores (0.0-1.0):
//
//   - 1.0: Explicitly requested by client
//   - 0.8-0.9: High confidence from parameter analysis
//   - 0.6-0.7: Medium confidence from tool patterns
//   - 0.4-0.5: Low confidence from heuristics
//
// Higher confidence fields are preserved even under aggressive filtering.
//
// # Thread Safety
//
// Analyzer implementations are safe for concurrent use.
package analyzer
