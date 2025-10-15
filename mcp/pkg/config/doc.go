// Package config provides configuration management for SlimMCP.
//
// This package handles loading, validating, and accessing SlimMCP configuration
// from YAML files, environment variables, and programmatic sources.
//
// # Overview
//
// Configuration defines proxy behavior, server routing, filtering strategies,
// and operational settings.
//
// # Basic Usage
//
//	// Load from YAML file
//	cfg, err := config.Load("slim.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Access configuration
//	fmt.Println("Listen address:", cfg.ListenAddress)
//	fmt.Println("Strategy:", cfg.Strategy)
//
// # Configuration Structure
//
// Example YAML configuration:
//
//	proxy:
//	  listenAddress: "0.0.0.0:8080"
//	  strategy: smart
//
//	  servers:
//	    notion:
//	      url: "http://localhost:3000"
//	      strategy: explicit
//	      defaults:
//	        notion-fetch:
//	          fields: [id, title, url, status]
//
//	    slack:
//	      url: "http://localhost:3001"
//
//	optimization:
//	  alwaysKeep:
//	    - id
//	    - type
//	    - url
//	    - title
//
//	  alwaysRemove:
//	    - icon
//	    - cover
//	    - created_by
//
//	  safety:
//	    fallbackOnError: true
//	    circuitBreaker:
//	      enabled: true
//	      threshold: 10
//
//	monitoring:
//	  enabled: true
//	  metricsPort: 9090
//	  logLevel: info
//
// # Environment Variables
//
// Configuration supports environment variable interpolation:
//
//	proxy:
//	  servers:
//	    notion:
//	      url: "http://localhost:3000"
//	      token: ${NOTION_TOKEN}
//
// # Validation
//
// The configuration is automatically validated on load:
//
//   - Required fields checked
//   - URL formats validated
//   - Port numbers verified
//   - Strategy values validated
//   - Circular dependencies detected
//
// # Defaults
//
// If not specified, these defaults apply:
//
//   - listenAddress: "0.0.0.0:8080"
//   - strategy: "smart"
//   - metricsPort: 9090
//   - logLevel: "info"
//   - fallbackOnError: true
//
// # Hot Reload
//
// Configuration can be reloaded without restarting:
//
//	if err := cfg.Reload(); err != nil {
//		log.Error("Failed to reload config:", err)
//	}
//
// Changes take effect immediately for new requests. In-flight requests use
// the previous configuration.
//
// # Thread Safety
//
// Config instances are safe for concurrent read access. Modifications require
// external synchronization or use of the Reload method.
package config
