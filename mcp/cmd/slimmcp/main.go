package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/slimmcp/slimmcp/pkg/config"
	"github.com/slimmcp/slimmcp/pkg/metrics"
	"github.com/slimmcp/slimmcp/pkg/proxy"
)

const version = "0.1.0"

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "examples/dev.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help message")
	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("SlimMCP v%s\n", version)
		os.Exit(0)
	}

	// Show help
	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("SlimMCP v%s starting...", version)
	log.Printf("Configuration loaded from: %s", *configPath)
	log.Printf("Listen address: %s", cfg.Proxy.ListenAddress)
	log.Printf("Default strategy: %s", cfg.Proxy.Strategy)
	log.Printf("Configured servers: %d", len(cfg.Proxy.Servers))

	// Create and start proxy server
	server := proxy.NewServer(cfg)

	// Start metrics server if enabled (T052)
	var metricsServer *metrics.PrometheusServer
	if cfg.Monitoring.Enabled {
		metricsServer = metrics.NewPrometheusServer(cfg.Monitoring.MetricsPort)
		go func() {
			log.Printf("Starting metrics server on port %d", cfg.Monitoring.MetricsPort)
			if err := metricsServer.Start(); err != nil {
				log.Printf("Metrics server error: %v", err)
			}
		}()
	}

	// Setup graceful shutdown
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Println("SlimMCP proxy server started successfully")
	if cfg.Monitoring.Enabled {
		log.Printf("Metrics available at http://localhost:%d/metrics", cfg.Monitoring.MetricsPort)
	}
	log.Println("Press Ctrl+C to shutdown")

	// Wait for interrupt signal
	<-quit
	log.Println("Server is shutting down...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown metrics server
	if metricsServer != nil {
		if err := metricsServer.Shutdown(ctx); err != nil {
			log.Printf("Metrics server shutdown error: %v", err)
		}
	}

	if err := server.Stop(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	close(done)
	log.Println("Server stopped gracefully")
}

func printHelp() {
	fmt.Println("SlimMCP - MCP Response Optimizer Proxy")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  slim [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -config string")
	fmt.Println("        Path to configuration file (default \"examples/dev.yaml\")")
	fmt.Println("  -version")
	fmt.Println("        Show version information")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Start with default configuration")
	fmt.Println("  slim")
	fmt.Println()
	fmt.Println("  # Start with custom configuration")
	fmt.Println("  slim -config /path/to/config.yaml")
	fmt.Println()
	fmt.Println("  # Show version")
	fmt.Println("  slim -version")
}
