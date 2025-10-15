package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	setDefaults(&config)

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for unspecified configuration fields
func setDefaults(config *Config) {
	// Proxy defaults
	if config.Proxy.ListenAddress == "" {
		config.Proxy.ListenAddress = "0.0.0.0:8080"
	}
	if config.Proxy.Strategy == "" {
		config.Proxy.Strategy = "smart"
	}

	// Server defaults
	for _, server := range config.Proxy.Servers {
		if server.Timeout == 0 {
			server.Timeout = 30000000000 // 30 seconds in nanoseconds
		}
		if server.Aggressiveness == "" {
			server.Aggressiveness = "medium"
		}
	}

	// Optimization defaults
	if len(config.Optimization.AlwaysKeep) == 0 {
		config.Optimization.AlwaysKeep = []string{"id", "type", "url", "title"}
	}

	// Safety defaults
	if config.Optimization.Safety.CircuitBreaker.Threshold == 0 {
		config.Optimization.Safety.CircuitBreaker.Threshold = 10
	}
	if config.Optimization.Safety.CircuitBreaker.Timeout == 0 {
		config.Optimization.Safety.CircuitBreaker.Timeout = 60000000000 // 60 seconds
	}

	// Monitoring defaults
	if config.Monitoring.MetricsPort == 0 {
		config.Monitoring.MetricsPort = 9090
	}
	if config.Monitoring.LogLevel == "" {
		config.Monitoring.LogLevel = "info"
	}
	if config.Monitoring.Prometheus.Path == "" {
		config.Monitoring.Prometheus.Path = "/metrics"
	}
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
