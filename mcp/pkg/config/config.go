package config

import "time"

// Config represents the top-level configuration structure
type Config struct {
	Proxy        ProxyConfig        `yaml:"proxy"`
	Optimization OptimizationConfig `yaml:"optimization"`
	Monitoring   MonitoringConfig   `yaml:"monitoring"`
	Learning     *LearningConfig    `yaml:"learning,omitempty"`
}

// ProxyConfig contains HTTP proxy server configuration
type ProxyConfig struct {
	ListenAddress string                   `yaml:"listenAddress"` // e.g., "0.0.0.0:8080"
	Strategy      string                   `yaml:"strategy"`      // Default strategy
	Servers       map[string]*ServerConfig `yaml:"servers"`       // Downstream servers
	CORS          *CORSConfig              `yaml:"cors,omitempty"`
	TLS           *TLSConfig               `yaml:"tls,omitempty"`
}

// ServerConfig contains configuration for a specific downstream MCP server
type ServerConfig struct {
	URL            string              `yaml:"url"`                      // Server URL
	Timeout        time.Duration       `yaml:"timeout"`                  // Request timeout
	Strategy       string              `yaml:"strategy,omitempty"`       // Override global strategy
	Aggressiveness string              `yaml:"aggressiveness,omitempty"` // "low", "medium", "high"
	Defaults       map[string][]string `yaml:"defaults,omitempty"`       // Per-tool defaults
}

// OptimizationConfig contains filtering behavior configuration
type OptimizationConfig struct {
	AlwaysKeep   []string     `yaml:"alwaysKeep"`   // Always include these fields
	AlwaysRemove []string     `yaml:"alwaysRemove"` // Always exclude these fields
	Safety       SafetyConfig `yaml:"safety"`       // Safety settings
}

// SafetyConfig contains safety and fallback settings
type SafetyConfig struct {
	FallbackOnError bool                 `yaml:"fallbackOnError"` // Return full response on error
	CircuitBreaker  CircuitBreakerConfig `yaml:"circuitBreaker"`  // Circuit breaker settings
}

// CircuitBreakerConfig contains circuit breaker settings for fault tolerance
type CircuitBreakerConfig struct {
	Enabled   bool          `yaml:"enabled"`   // Enable circuit breaker
	Threshold int           `yaml:"threshold"` // Failures before opening
	Timeout   time.Duration `yaml:"timeout"`   // Time to wait before retry
}

// MonitoringConfig contains observability configuration
type MonitoringConfig struct {
	Enabled        bool                 `yaml:"enabled"`        // Enable monitoring
	MetricsPort    int                  `yaml:"metricsPort"`    // Prometheus metrics port
	LogLevel       string               `yaml:"logLevel"`       // Log level
	Prometheus     PrometheusConfig     `yaml:"prometheus"`     // Prometheus settings
	RequestLogging RequestLoggingConfig `yaml:"requestLogging"` // Request logging settings
}

// PrometheusConfig contains Prometheus-specific settings
type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled"` // Enable Prometheus metrics
	Path    string `yaml:"path"`    // Metrics endpoint path (default: /metrics)
}

// RequestLoggingConfig contains request/response logging configuration
type RequestLoggingConfig struct {
	Enabled    bool `yaml:"enabled"`    // Enable request logging
	LogHeaders bool `yaml:"logHeaders"` // Log headers
	LogBody    bool `yaml:"logBody"`    // Log request/response bodies
}

// CORSConfig contains CORS policy configuration
type CORSConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowedOrigins   []string `yaml:"allowedOrigins"`
	AllowedMethods   []string `yaml:"allowedMethods"`
	AllowedHeaders   []string `yaml:"allowedHeaders"`
	AllowCredentials bool     `yaml:"allowCredentials"`
}

// TLSConfig contains TLS certificate configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

// LearningConfig contains learning mode configuration (optional/future)
type LearningConfig struct {
	Enabled          bool   `yaml:"enabled"`
	CollectUsageData bool   `yaml:"collectUsageData"`
	ImproveInference bool   `yaml:"improveInference"`
	PersistencePath  string `yaml:"persistencePath,omitempty"`
}

// GetServerConfig returns the configuration for a specific server
func (c *Config) GetServerConfig(serverName string) (*ServerConfig, bool) {
	if c.Proxy.Servers == nil {
		return nil, false
	}
	server, ok := c.Proxy.Servers[serverName]
	return server, ok
}

// GetStrategy returns the effective strategy for a server (server override or global)
func (c *Config) GetStrategy(serverName string) string {
	if server, ok := c.GetServerConfig(serverName); ok && server.Strategy != "" {
		return server.Strategy
	}
	return c.Proxy.Strategy
}

// IsFieldAlwaysKept checks if a field is in the always-keep list
func (c *Config) IsFieldAlwaysKept(field string) bool {
	for _, f := range c.Optimization.AlwaysKeep {
		if f == field {
			return true
		}
	}
	return false
}

// IsFieldAlwaysRemoved checks if a field is in the always-remove list
func (c *Config) IsFieldAlwaysRemoved(field string) bool {
	for _, f := range c.Optimization.AlwaysRemove {
		if f == field {
			return true
		}
	}
	return false
}
