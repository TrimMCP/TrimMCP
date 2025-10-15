package metrics

import (
	"time"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// Collector defines the interface for collecting and recording metrics
type Collector interface {
	// RecordFiltering records statistics from a filtering operation
	RecordFiltering(stats types.FilteringStatistics)

	// RecordRequest records a request being processed
	RecordRequest(server, tool, strategy, status string)

	// RecordLatency records the latency of an operation
	RecordLatency(operation string, durationMs float64)

	// RecordError records an error occurrence
	RecordError(errorType, server string)
}

// PrometheusCollector implements the Collector interface using Prometheus
type PrometheusCollector struct {
	enabled bool
}

// NewPrometheusCollector creates a new Prometheus-based metrics collector
func NewPrometheusCollector(enabled bool) *PrometheusCollector {
	return &PrometheusCollector{
		enabled: enabled,
	}
}

// RecordFiltering records filtering statistics to Prometheus
func (c *PrometheusCollector) RecordFiltering(stats types.FilteringStatistics) {
	if !c.enabled {
		return
	}

	// Record token metrics
	RecordTokenMetrics(stats.Server, stats.Tool, stats.TokensOriginal, stats.TokensSaved)

	// Record filtering duration
	RecordFilteringDuration(stats.Strategy, stats.Duration)
}

// RecordRequest records a processed request
func (c *PrometheusCollector) RecordRequest(server, tool, strategy, status string) {
	if !c.enabled {
		return
	}

	RecordRequest(server, tool, status, strategy)
}

// RecordLatency records the latency of an operation
func (c *PrometheusCollector) RecordLatency(operation string, durationMs float64) {
	if !c.enabled {
		return
	}

	// Convert milliseconds to duration and record
	duration := time.Duration(durationMs * float64(time.Millisecond))
	RecordFilteringDuration(operation, duration)
}

// RecordError records an error occurrence
func (c *PrometheusCollector) RecordError(errorType, server string) {
	if !c.enabled {
		return
	}

	// Record as a failed request
	RecordRequest(server, "unknown", "error", errorType)
}

// NoOpCollector is a collector that does nothing (for when metrics are disabled)
type NoOpCollector struct{}

// NewNoOpCollector creates a no-op collector
func NewNoOpCollector() *NoOpCollector {
	return &NoOpCollector{}
}

// RecordFiltering does nothing
func (c *NoOpCollector) RecordFiltering(stats types.FilteringStatistics) {}

// RecordRequest does nothing
func (c *NoOpCollector) RecordRequest(server, tool, strategy, status string) {}

// RecordLatency does nothing
func (c *NoOpCollector) RecordLatency(operation string, durationMs float64) {}

// RecordError does nothing
func (c *NoOpCollector) RecordError(errorType, server string) {}
