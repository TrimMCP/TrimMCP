// Package filter provides response filtering and optimization for MCP responses.
//
// This package implements the core filtering logic that reduces MCP response sizes
// by removing unnecessary fields while preserving essential data and structure.
//
// # Overview
//
// The filter package supports multiple filtering strategies and field selection modes,
// enabling fine-grained control over response optimization.
//
// # Basic Usage
//
//	engine := filter.NewEngine(&config.Config{
//		Strategy: config.StrategyMinimal,
//		AlwaysKeep: []string{"id", "type", "url", "title"},
//	})
//
//	filtered, metadata, err := engine.Filter(response, fieldSelection)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Token savings: %.1f%%\n", metadata.Savings)
//
// # Filtering Strategies
//
// # Minimal Strategy
//
// Returns only critical fields that are essential for MCP protocol compatibility:
//
//   - id: Unique identifier
//   - type: Object type
//   - url: Resource URL
//   - title: Human-readable name
//   - object: MCP object type
//
// # Explicit Strategy
//
// Clients specify exactly which fields to keep using include/exclude lists:
//
//	fieldSelection := &types.FieldSelection{
//		Include: []string{"id", "title", "status"},
//		Strategy: types.StrategyExplicit,
//	}
//
// # Smart Strategy (planned for v0.2)
//
// Analyzes request parameters to infer which fields are likely needed.
//
// # Nested Field Selection
//
// Supports dot notation for selecting nested fields:
//
//	fieldSelection := &types.FieldSelection{
//		Include: []string{"id", "properties.Status", "properties.Priority"},
//	}
//
// This preserves the parent structure while filtering child objects.
//
// # Metadata
//
// Filtering produces metadata about the operation:
//
//   - OriginalSize: Token count before filtering
//   - FilteredSize: Token count after filtering
//   - Savings: Percentage reduction
//   - FieldsKept: List of preserved fields
//   - FieldsRemoved: Count of removed fields
//   - FilteringTime: Processing duration in milliseconds
//
// # Performance
//
// The filter engine is highly optimized:
//
//   - <5ms filtering time for typical responses
//   - Memory-efficient recursive filtering
//   - Zero-allocation field matching for common cases
//
// # Thread Safety
//
// FilterEngine is safe for concurrent use. Multiple goroutines can call Filter
// simultaneously on the same engine instance.
package filter
