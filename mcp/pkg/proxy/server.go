package proxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/slimmcp/slimmcp/pkg/analyzer"
	"github.com/slimmcp/slimmcp/pkg/capability"
	"github.com/slimmcp/slimmcp/pkg/client"
	"github.com/slimmcp/slimmcp/pkg/config"
	"github.com/slimmcp/slimmcp/pkg/filter"
	"github.com/slimmcp/slimmcp/pkg/metrics"
)

// Server is the HTTP proxy server
type Server struct {
	config           *config.Config
	httpServer       *http.Server
	analyzer         analyzer.Analyzer
	filter           filter.Filter
	client           *client.MCPClient
	metricsCollector metrics.Collector
	capabilityCache  *capability.Cache // T069: Capability detection integration
}

// NewServer creates a new proxy server
func NewServer(cfg *config.Config) *Server {
	// Create components
	analyzer := analyzer.NewMinimalAnalyzer() // Default to minimal strategy
	filterEngine := filter.NewEngine()
	mcpClient := client.NewMCPClient(30 * time.Second)

	// Create metrics collector (T052)
	var collector metrics.Collector
	if cfg.Monitoring.Enabled {
		collector = metrics.NewPrometheusCollector(true)
	} else {
		collector = metrics.NewNoOpCollector()
	}

	// Create capability detector and cache (T069)
	detector := capability.NewDetector(mcpClient)
	capCache := capability.NewCache(detector, 1*time.Hour)

	server := &Server{
		config:           cfg,
		analyzer:         analyzer,
		filter:           filterEngine,
		client:           mcpClient,
		metricsCollector: collector,
		capabilityCache:  capCache,
	}

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:              cfg.Proxy.ListenAddress,
		Handler:           server.setupRouter(),
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return server
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting SlimMCP proxy server on %s", s.config.Proxy.ListenAddress)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Shutting down proxy server...")
	return s.httpServer.Shutdown(ctx)
}

// setupRouter sets up the HTTP routes
func (s *Server) setupRouter() http.Handler {
	mux := http.NewServeMux()

	// Chain middleware for proxy endpoints (T050)
	proxyHandler := s.corsMiddleware(s.loggingMiddleware(s.metricsMiddleware(s.handleProxy)))

	// Main proxy endpoint (POST /)
	mux.HandleFunc("/", proxyHandler)

	// Server-specific endpoints (POST /{server})
	for serverName := range s.config.Proxy.Servers {
		pattern := fmt.Sprintf("/%s", serverName)
		mux.HandleFunc(pattern, proxyHandler)
	}

	// Health check endpoint (no middleware needed)
	mux.HandleFunc("/health", s.handleHealth)

	return mux
}
