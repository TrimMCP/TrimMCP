package benchmark

import (
	"fmt"
	"os"
	"sort"

	"github.com/olekukonko/tablewriter"
)

// Reporter formats and displays benchmark results.
type Reporter struct {
	verbose bool
}

// NewReporter creates a new Reporter.
func NewReporter(verbose bool) *Reporter {
	return &Reporter{
		verbose: verbose,
	}
}

// PrintResults outputs the benchmark results as formatted tables.
func (r *Reporter) PrintResults(results []BenchmarkResult, stats *SummaryStats) {
	fmt.Println()
	fmt.Println("================================================================================")
	fmt.Println("Token Savings Benchmark Results")
	fmt.Println("================================================================================")
	fmt.Println()

	// Print individual operation results
	r.printOperationResults(results)

	fmt.Println()

	// Print summary statistics
	r.printSummary(stats)

	fmt.Println()
}

// printOperationResults prints a table with results for each operation.
func (r *Reporter) printOperationResults(results []BenchmarkResult) {
	fmt.Println("Operation Results:")
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Server", "Operation", "Size Reduction", "Token Reduction (tiktoken)", "Validation")

	for _, result := range results {
		// Get tiktoken reduction (primary metric)
		tiktokenReduction := result.TokenReduction["tiktoken-cl100k"]

		validation := "✓ PASS"
		if !result.ValidationPassed {
			validation = "✗ FAIL"
		}

		table.Append(
			result.Operation.Server,
			result.Operation.Operation,
			fmt.Sprintf("%.1f%%", result.SizeReduction),
			fmt.Sprintf("%.1f%%", tiktokenReduction),
			validation,
		)
	}

	table.Render()

	// Print detailed tokenizer comparison if verbose
	if r.verbose {
		fmt.Println()
		r.printDetailedTokenComparison(results)
	}
}

// printDetailedTokenComparison prints token counts for all tokenizers.
func (r *Reporter) printDetailedTokenComparison(results []BenchmarkResult) {
	fmt.Println("Detailed Token Comparison:")
	fmt.Println()

	// Get tokenizer names from first result
	if len(results) == 0 || len(results[0].OriginalTokens) == 0 {
		return
	}

	tokenizerNames := make([]string, len(results[0].OriginalTokens))
	for i, m := range results[0].OriginalTokens {
		tokenizerNames[i] = m.TokenizerName
	}

	table := tablewriter.NewWriter(os.Stdout)
	headers := []interface{}{"Server", "Operation"}
	for _, name := range tokenizerNames {
		headers = append(headers, name+" (original)")
		headers = append(headers, name+" (filtered)")
		headers = append(headers, name+" (reduction)")
	}
	table.Header(headers...)

	for _, result := range results {
		row := []interface{}{result.Operation.Server, result.Operation.Operation}

		for _, name := range tokenizerNames {
			// Find original and filtered counts
			var originalCount, filteredCount int
			for _, m := range result.OriginalTokens {
				if m.TokenizerName == name {
					originalCount = m.TokenCount
					break
				}
			}
			for _, m := range result.FilteredTokens {
				if m.TokenizerName == name {
					filteredCount = m.TokenCount
					break
				}
			}

			reduction := result.TokenReduction[name]
			row = append(row, fmt.Sprintf("%d", originalCount))
			row = append(row, fmt.Sprintf("%d", filteredCount))
			row = append(row, fmt.Sprintf("%.1f%%", reduction))
		}

		table.Append(row...)
	}

	table.Render()
}

// printSummary prints aggregate statistics.
func (r *Reporter) printSummary(stats *SummaryStats) {
	fmt.Println("Summary Statistics:")
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Metric", "Value")

	table.Append("Total Operations", fmt.Sprintf("%d", stats.TotalOperations))
	table.Append("Average Size Reduction", fmt.Sprintf("%.1f%%", stats.AvgSizeReduction))

	// Sort tokenizer names for consistent output
	tokenizerNames := make([]string, 0, len(stats.AvgTokenReduction))
	for name := range stats.AvgTokenReduction {
		tokenizerNames = append(tokenizerNames, name)
	}
	sort.Strings(tokenizerNames)

	for _, name := range tokenizerNames {
		avg := stats.AvgTokenReduction[name]
		min := stats.MinTokenReduction[name]
		max := stats.MaxTokenReduction[name]

		table.Append(
			fmt.Sprintf("Average Token Reduction (%s)", name),
			fmt.Sprintf("%.1f%% (min: %.1f%%, max: %.1f%%)", avg, min, max),
		)
	}

	table.Append("Validation Pass Rate", fmt.Sprintf("%.1f%%", stats.ValidationPassRate))

	table.Render()

	// Print success/failure message
	fmt.Println()
	if stats.ValidationPassRate == 100.0 {
		fmt.Println("✓ All operations passed validation")
	} else {
		fmt.Printf("⚠️  %d operations failed validation\n",
			stats.TotalOperations-int(float64(stats.TotalOperations)*stats.ValidationPassRate/100.0))
	}

	// Check if target met
	avgTiktokenReduction := stats.AvgTokenReduction["tiktoken-cl100k"]
	if avgTiktokenReduction >= 50.0 {
		fmt.Printf("✓ Target met: %.1f%% average token reduction (≥50%% required)\n", avgTiktokenReduction)
	} else {
		fmt.Printf("✗ Target NOT met: %.1f%% average token reduction (<50%% required)\n", avgTiktokenReduction)
	}
}
