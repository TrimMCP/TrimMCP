package proxy

import (
	"log"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status code
		written:        false,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// metricsMiddleware wraps handlers to automatically record request metrics
// T050: Middleware for request metrics collection
func (s *Server) metricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only apply metrics to proxy endpoints, not health or metrics endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
			next(w, r)
			return
		}

		startTime := time.Now()

		// Wrap response writer to capture status code
		rw := newResponseWriter(w)

		// Call the next handler
		next(rw, r)

		// Record metrics after request completes
		duration := time.Since(startTime)

		// Extract server name from path
		serverName := s.extractServerName(r.URL.Path)

		// Determine status based on HTTP status code
		status := "success"
		if rw.statusCode >= 400 && rw.statusCode < 500 {
			status = "client_error"
		} else if rw.statusCode >= 500 {
			status = "server_error"
		}

		// Record request metrics if collector is available
		if s.metricsCollector != nil {
			// We don't have tool name at this level, so use "unknown"
			// The actual filtering metrics with tool name are recorded in handleProxy
			strategy := s.config.GetStrategy(serverName)
			s.metricsCollector.RecordRequest(serverName, "unknown", strategy, status)

			// Log request completion
			if s.config.Monitoring.LogLevel == "debug" {
				log.Printf("Request completed: server=%s status=%s duration=%v",
					serverName, status, duration)
			}
		}
	}
}

// loggingMiddleware adds request logging
func (s *Server) loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only log if log level is debug or info
		if s.config.Monitoring.LogLevel == "debug" || s.config.Monitoring.LogLevel == "info" {
			log.Printf("Incoming request: method=%s path=%s remote=%s",
				r.Method, r.URL.Path, r.RemoteAddr)
		}
		next(w, r)
	}
}

// corsMiddleware adds CORS headers if configured
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers if configured
		if s.config.Proxy.CORS != nil {
			if len(s.config.Proxy.CORS.AllowedOrigins) > 0 {
				origin := s.config.Proxy.CORS.AllowedOrigins[0]
				if origin == "*" || contains(s.config.Proxy.CORS.AllowedOrigins, r.Header.Get("Origin")) {
					w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			if len(s.config.Proxy.CORS.AllowedMethods) > 0 {
				methods := ""
				for i, method := range s.config.Proxy.CORS.AllowedMethods {
					if i > 0 {
						methods += ", "
					}
					methods += method
				}
				w.Header().Set("Access-Control-Allow-Methods", methods)
			}

			if len(s.config.Proxy.CORS.AllowedHeaders) > 0 {
				headers := ""
				for i, header := range s.config.Proxy.CORS.AllowedHeaders {
					if i > 0 {
						headers += ", "
					}
					headers += header
				}
				w.Header().Set("Access-Control-Allow-Headers", headers)
			}

			if s.config.Proxy.CORS.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
