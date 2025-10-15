package benchmark

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/slimmcp/slimmcp/pkg/cache"
	"github.com/slimmcp/slimmcp/pkg/filter"
	"github.com/slimmcp/slimmcp/pkg/token"
)

// Runner orchestrates the benchmark execution.
type Runner struct {
	cacheLoader cache.CacheLoader
	filter      filter.Filter
	tokenizers  []token.Tokenizer
	validator   Validator
	verbose     bool
}

// NewRunner creates a new benchmark runner with the specified components.
func NewRunner(
	cacheLoader cache.CacheLoader,
	filter filter.Filter,
	tokenizers []token.Tokenizer,
	validator Validator,
	verbose bool,
) *Runner {
	return &Runner{
		cacheLoader: cacheLoader,
		filter:      filter,
		tokenizers:  tokenizers,
		validator:   validator,
		verbose:     verbose,
	}
}

// Run executes the benchmark for the specified operations.
// Returns results for each operation and summary statistics.
func (r *Runner) Run(operations []BenchmarkOperation) ([]BenchmarkResult, *SummaryStats, error) {
	if r.verbose {
		log.Printf("Starting benchmark for %d operations...\n", len(operations))
	}

	results := make([]BenchmarkResult, 0, len(operations))

	for _, op := range operations {
		if r.verbose {
			log.Printf("Benchmarking %s/%s...\n", op.Server, op.Operation)
		}

		result, err := r.runOperation(op)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to benchmark %s/%s: %w", op.Server, op.Operation, err)
		}

		results = append(results, result)

		if r.verbose {
			r.logResult(result)
		}
	}

	// Calculate summary statistics
	stats := r.calculateSummary(results)

	if r.verbose {
		log.Println("Benchmark complete!")
	}

	return results, stats, nil
}

// runOperation executes the benchmark for a single operation.
func (r *Runner) runOperation(op BenchmarkOperation) (BenchmarkResult, error) {
	// Load cached response
	cached, err := r.cacheLoader.Load(op.Server, op.Operation)
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("failed to load cache: %w", err)
	}

	// Serialize original response to JSON for size/token measurement
	originalJSON, err := json.Marshal(cached.Response)
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("failed to marshal original response: %w", err)
	}

	// Apply filter
	filtered, err := r.filter.Apply(cached.Response)
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("failed to filter response: %w", err)
	}

	// Serialize filtered response
	filteredJSON, err := json.Marshal(filtered)
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("failed to marshal filtered response: %w", err)
	}

	// Measure tokens for original and filtered
	originalTokens, err := r.measureTokens(string(originalJSON))
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("failed to measure original tokens: %w", err)
	}

	filteredTokens, err := r.measureTokens(string(filteredJSON))
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("failed to measure filtered tokens: %w", err)
	}

	// Calculate reductions
	sizeReduction := calculateReduction(len(originalJSON), len(filteredJSON))
	tokenReduction := r.calculateTokenReduction(originalTokens, filteredTokens)

	// Validate filtered response
	validationPassed := true
	validationError := ""
	if r.validator != nil {
		if err := r.validator.Validate(op.Operation, filtered); err != nil {
			validationPassed = false
			validationError = err.Error()
		}
	}

	return BenchmarkResult{
		Operation:        op,
		OriginalSize:     len(originalJSON),
		FilteredSize:     len(filteredJSON),
		SizeReduction:    sizeReduction,
		OriginalTokens:   originalTokens,
		FilteredTokens:   filteredTokens,
		TokenReduction:   tokenReduction,
		ValidationPassed: validationPassed,
		ValidationError:  validationError,
	}, nil
}

// measureTokens counts tokens using all configured tokenizers.
func (r *Runner) measureTokens(text string) ([]TokenMeasurement, error) {
	measurements := make([]TokenMeasurement, 0, len(r.tokenizers))

	for _, tok := range r.tokenizers {
		count, err := tok.Count(text)
		if err != nil {
			return nil, fmt.Errorf("tokenizer %s failed: %w", tok.Name(), err)
		}

		measurements = append(measurements, TokenMeasurement{
			TokenizerName: tok.Name(),
			TokenCount:    count,
		})
	}

	return measurements, nil
}

// calculateTokenReduction computes percentage reduction for each tokenizer.
func (r *Runner) calculateTokenReduction(original, filtered []TokenMeasurement) map[string]float64 {
	reductions := make(map[string]float64)

	for i := range original {
		if i >= len(filtered) {
			break
		}

		name := original[i].TokenizerName
		reduction := calculateReduction(original[i].TokenCount, filtered[i].TokenCount)
		reductions[name] = reduction
	}

	return reductions
}

// calculateReduction computes percentage reduction: (original - filtered) / original * 100
func calculateReduction(original, filtered int) float64 {
	if original == 0 {
		return 0.0
	}

	return float64(original-filtered) / float64(original) * 100.0
}

// calculateSummary computes aggregate statistics from benchmark results.
func (r *Runner) calculateSummary(results []BenchmarkResult) *SummaryStats {
	if len(results) == 0 {
		return &SummaryStats{}
	}

	// Initialize accumulators
	totalSizeReduction := 0.0
	tokenReductionSums := make(map[string]float64)
	tokenReductionMins := make(map[string]float64)
	tokenReductionMaxs := make(map[string]float64)
	passedCount := 0

	// First pass: initialize min/max
	for tokName := range results[0].TokenReduction {
		tokenReductionMins[tokName] = 100.0
		tokenReductionMaxs[tokName] = 0.0
	}

	// Accumulate statistics
	for _, result := range results {
		totalSizeReduction += result.SizeReduction

		for tokName, reduction := range result.TokenReduction {
			tokenReductionSums[tokName] += reduction

			if reduction < tokenReductionMins[tokName] {
				tokenReductionMins[tokName] = reduction
			}
			if reduction > tokenReductionMaxs[tokName] {
				tokenReductionMaxs[tokName] = reduction
			}
		}

		if result.ValidationPassed {
			passedCount++
		}
	}

	// Calculate averages
	count := float64(len(results))
	avgTokenReduction := make(map[string]float64)
	for tokName, sum := range tokenReductionSums {
		avgTokenReduction[tokName] = sum / count
	}

	return &SummaryStats{
		TotalOperations:    len(results),
		AvgSizeReduction:   totalSizeReduction / count,
		AvgTokenReduction:  avgTokenReduction,
		MinTokenReduction:  tokenReductionMins,
		MaxTokenReduction:  tokenReductionMaxs,
		ValidationPassRate: float64(passedCount) / count * 100.0,
	}
}

// logResult logs a single benchmark result in verbose mode.
func (r *Runner) logResult(result BenchmarkResult) {
	log.Printf("  Size: %d → %d bytes (%.1f%% reduction)\n",
		result.OriginalSize, result.FilteredSize, result.SizeReduction)

	for _, measurement := range result.OriginalTokens {
		filtered := result.FilteredTokens[0] // Assumes same order
		for _, f := range result.FilteredTokens {
			if f.TokenizerName == measurement.TokenizerName {
				filtered = f
				break
			}
		}

		reduction := result.TokenReduction[measurement.TokenizerName]
		log.Printf("  %s: %d → %d tokens (%.1f%% reduction)\n",
			measurement.TokenizerName, measurement.TokenCount, filtered.TokenCount, reduction)
	}

	if !result.ValidationPassed {
		log.Printf("  ⚠️  Validation FAILED: %s\n", result.ValidationError)
	}
}
