package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// tokensOriginalTotal tracks the total number of tokens in original (unfiltered) responses
	tokensOriginalTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_optimizer_tokens_original_total",
			Help: "Total tokens in original (unfiltered) responses",
		},
		[]string{"server", "tool"},
	)

	// tokensSavedTotal tracks the total number of tokens saved by filtering
	tokensSavedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_optimizer_tokens_saved_total",
			Help: "Total tokens saved by filtering",
		},
		[]string{"server", "tool"},
	)

	// requestsTotal tracks the total number of requests processed
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_optimizer_requests_total",
			Help: "Total requests processed",
		},
		[]string{"server", "tool", "status", "strategy"},
	)

	// filteringDuration tracks the time spent filtering responses
	filteringDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_optimizer_filtering_duration_seconds",
			Help:    "Time spent filtering responses in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0}, // 1ms to 1s
		},
		[]string{"strategy"},
	)

	// savingsPercentage tracks the current token savings percentage per server
	savingsPercentage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_optimizer_savings_percentage",
			Help: "Current token savings percentage",
		},
		[]string{"server"},
	)
)

func init() {
	// Register all metrics with the default registry
	prometheus.MustRegister(tokensOriginalTotal)
	prometheus.MustRegister(tokensSavedTotal)
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(filteringDuration)
	prometheus.MustRegister(savingsPercentage)
}

// PrometheusServer wraps an HTTP server for serving Prometheus metrics
type PrometheusServer struct {
	server *http.Server
	port   int
}

// NewPrometheusServer creates a new Prometheus metrics server
func NewPrometheusServer(port int) *PrometheusServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &PrometheusServer{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
		port: port,
	}
}

// Start begins serving Prometheus metrics
func (s *PrometheusServer) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the metrics server
func (s *PrometheusServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Port returns the port the server is listening on
func (s *PrometheusServer) Port() int {
	return s.port
}

// RecordTokenMetrics records token usage metrics
func RecordTokenMetrics(server, tool string, original, saved int) {
	tokensOriginalTotal.WithLabelValues(server, tool).Add(float64(original))
	tokensSavedTotal.WithLabelValues(server, tool).Add(float64(saved))

	// Update savings percentage gauge
	if original > 0 {
		savingsPercent := float64(saved) / float64(original) * 100.0
		savingsPercentage.WithLabelValues(server).Set(savingsPercent)
	}
}

// RecordRequest records a processed request
func RecordRequest(server, tool, status, strategy string) {
	requestsTotal.WithLabelValues(server, tool, status, strategy).Inc()
}

// RecordFilteringDuration records the time spent filtering
func RecordFilteringDuration(strategy string, duration time.Duration) {
	filteringDuration.WithLabelValues(strategy).Observe(duration.Seconds())
}
