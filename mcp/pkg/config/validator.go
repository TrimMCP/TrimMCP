package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate proxy configuration
	if err := c.validateProxy(); err != nil {
		return fmt.Errorf("proxy config: %w", err)
	}

	// Validate optimization configuration
	if err := c.validateOptimization(); err != nil {
		return fmt.Errorf("optimization config: %w", err)
	}

	// Validate monitoring configuration
	if err := c.validateMonitoring(); err != nil {
		return fmt.Errorf("monitoring config: %w", err)
	}

	return nil
}

// validateProxy validates proxy configuration
func (c *Config) validateProxy() error {
	// Validate listen address
	if c.Proxy.ListenAddress == "" {
		return types.ErrInvalidListenAddress
	}

	// Validate strategy
	validStrategies := map[string]bool{
		"smart":       true,
		"minimal":     true,
		"explicit":    true,
		"passthrough": true,
	}
	if !validStrategies[c.Proxy.Strategy] {
		return types.ErrInvalidStrategy
	}

	// Validate servers
	if len(c.Proxy.Servers) == 0 {
		return fmt.Errorf("at least one server must be configured")
	}

	for name, server := range c.Proxy.Servers {
		if err := validateServerConfig(name, server); err != nil {
			return err
		}
	}

	// Validate TLS if enabled
	if c.Proxy.TLS != nil && c.Proxy.TLS.Enabled {
		if c.Proxy.TLS.CertFile == "" || c.Proxy.TLS.KeyFile == "" {
			return fmt.Errorf("TLS cert and key files are required when TLS is enabled")
		}
	}

	return nil
}

// validateServerConfig validates a server configuration
func validateServerConfig(name string, server *ServerConfig) error {
	if server.URL == "" {
		return fmt.Errorf("server '%s': %w", name, types.ErrMissingServerURL)
	}

	// Validate URL format
	parsedURL, err := url.Parse(server.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return fmt.Errorf("server '%s': %w", name, types.ErrInvalidServerURL)
	}

	// Validate timeout
	if server.Timeout < 0 {
		return fmt.Errorf("server '%s': %w", name, types.ErrInvalidTimeout)
	}

	// Validate strategy override if specified
	if server.Strategy != "" {
		validStrategies := map[string]bool{
			"smart":       true,
			"minimal":     true,
			"explicit":    true,
			"passthrough": true,
		}
		if !validStrategies[server.Strategy] {
			return fmt.Errorf("server '%s': %w", name, types.ErrInvalidStrategy)
		}
	}

	// Validate aggressiveness
	if server.Aggressiveness != "" {
		validLevels := map[string]bool{
			"low":    true,
			"medium": true,
			"high":   true,
		}
		if !validLevels[server.Aggressiveness] {
			return fmt.Errorf("server '%s': invalid aggressiveness level '%s' (must be low, medium, or high)",
				name, server.Aggressiveness)
		}
	}

	return nil
}

// validateOptimization validates optimization configuration
func (c *Config) validateOptimization() error {
	// Check for conflicts between alwaysKeep and alwaysRemove
	keepMap := make(map[string]bool)
	for _, field := range c.Optimization.AlwaysKeep {
		keepMap[field] = true
	}

	for _, field := range c.Optimization.AlwaysRemove {
		if keepMap[field] {
			return fmt.Errorf("field '%s' is in both alwaysKeep and alwaysRemove", field)
		}
	}

	// Validate circuit breaker
	if c.Optimization.Safety.CircuitBreaker.Enabled {
		if c.Optimization.Safety.CircuitBreaker.Threshold <= 0 {
			return fmt.Errorf("circuit breaker threshold must be positive")
		}
		if c.Optimization.Safety.CircuitBreaker.Timeout <= 0 {
			return fmt.Errorf("circuit breaker timeout must be positive")
		}
	}

	return nil
}

// validateMonitoring validates monitoring configuration
func (c *Config) validateMonitoring() error {
	// Validate metrics port
	if c.Monitoring.Enabled {
		if c.Monitoring.MetricsPort < 1024 || c.Monitoring.MetricsPort > 65535 {
			return types.ErrInvalidPort
		}
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[strings.ToLower(c.Monitoring.LogLevel)] {
		return fmt.Errorf("invalid log level '%s' (must be debug, info, warn, or error)", c.Monitoring.LogLevel)
	}

	return nil
}
