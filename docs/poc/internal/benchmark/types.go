package benchmark

// BenchmarkOperation represents a single MCP operation to benchmark.
type BenchmarkOperation struct {
	Server    string // "github" or "notion"
	Operation string // e.g., "list_pull_requests", "create_page"
}

// TokenMeasurement represents token counts for a specific text and tokenizer.
type TokenMeasurement struct {
	TokenizerName string // e.g., "tiktoken-cl100k", "word-count"
	TokenCount    int    // Number of tokens counted
}

// BenchmarkResult represents the complete benchmark results for one operation.
type BenchmarkResult struct {
	Operation         BenchmarkOperation
	OriginalSize      int                // Bytes in original response
	FilteredSize      int                // Bytes in filtered response
	SizeReduction     float64            // Percentage: (original - filtered) / original * 100
	OriginalTokens    []TokenMeasurement // Token counts for original response
	FilteredTokens    []TokenMeasurement // Token counts for filtered response
	TokenReduction    map[string]float64 // Percentage reduction per tokenizer
	ValidationPassed  bool               // Whether filtered response passed validation
	ValidationError   string             // Error message if validation failed
}

// SummaryStats represents aggregate statistics across all operations.
type SummaryStats struct {
	TotalOperations   int                // Total number of operations benchmarked
	AvgSizeReduction  float64            // Average size reduction percentage
	AvgTokenReduction map[string]float64 // Average token reduction per tokenizer
	MinTokenReduction map[string]float64 // Minimum token reduction per tokenizer
	MaxTokenReduction map[string]float64 // Maximum token reduction per tokenizer
	ValidationPassRate float64           // Percentage of operations that passed validation
}
