package proxy

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/slimmcp/slimmcp/pkg/analyzer"
	"github.com/slimmcp/slimmcp/pkg/filter"
	"github.com/slimmcp/slimmcp/pkg/types"
)

// handleProxy handles incoming MCP requests
func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, nil, types.ErrorCodeInvalidRequest, "Failed to read request body", nil)
		return
	}
	defer r.Body.Close()

	// Parse request
	var request types.MCPRequest
	if err := json.Unmarshal(body, &request); err != nil {
		sendError(w, nil, types.ErrorCodeParseError, "Invalid JSON", nil)
		return
	}

	// Validate request
	if err := request.Validate(); err != nil {
		sendError(w, request.ID, types.ErrorCodeInvalidRequest, err.Error(), nil)
		return
	}

	// Extract server name from path
	serverName := s.extractServerName(r.URL.Path)
	if serverName == "" {
		sendError(w, request.ID, types.ErrorCodeInvalidRequest, "No server specified", nil)
		return
	}

	// Get server config
	serverConfig, ok := s.config.GetServerConfig(serverName)
	if !ok {
		sendError(w, request.ID, types.ErrorCodeInvalidRequest, "Unknown server: "+serverName, nil)
		return
	}

	// Forward request to downstream server
	ctx := r.Context()
	headers := extractHeaders(r.Header)

	startTime := time.Now()
	responseBody, err := s.client.Forward(ctx, serverConfig.URL, &request, headers)
	if err != nil {
		log.Printf("Error forwarding request: %v", err)
		sendError(w, request.ID, types.ErrorCodeDownstreamError, "Downstream server error", nil)
		return
	}

	// Parse downstream response
	var downstreamResponse types.MCPResponse
	if err := json.Unmarshal(responseBody, &downstreamResponse); err != nil {
		sendError(w, request.ID, types.ErrorCodeInternalError, "Failed to parse downstream response", nil)
		return
	}

	// If downstream returned an error, pass it through
	if downstreamResponse.Error != nil {
		sendJSONResponse(w, &downstreamResponse)
		return
	}

	// T069: Check if downstream server supports native field selection
	capabilities, err := s.capabilityCache.GetCapabilities(ctx, serverConfig.URL)
	if err != nil {
		log.Printf("Warning: Failed to detect capabilities for %s: %v", serverConfig.URL, err)
		// Continue with proxy-side filtering on error
	}

	// If server supports native field selection and request has field selection metadata
	if capabilities != nil && capabilities.SupportsFieldSelection && request.Meta != nil && request.Meta.FieldSelection != nil {
		// T070: Server handles filtering natively - pass through response
		log.Printf("Server %s supports native field selection - delegating to server", serverName)

		// Update metadata to indicate server-side filtering
		if downstreamResponse.Meta != nil && downstreamResponse.Meta.FieldSelection != nil {
			downstreamResponse.Meta.FieldSelection.FilteredBy = "server"
		}

		sendJSONResponse(w, &downstreamResponse)
		return
	}

	// Get effective strategy for this request
	strategy := s.config.GetStrategy(serverName)

	// Check if request overrides strategy
	effectiveStrategy := strategy
	if request.Meta != nil && request.Meta.FieldSelection != nil && request.Meta.FieldSelection.Strategy != "" {
		effectiveStrategy = request.Meta.FieldSelection.Strategy
	}

	// Create analyzer based on request and config
	analyzerCfg := &analyzer.Config{
		AlwaysKeep:   s.config.Optimization.AlwaysKeep,
		AlwaysRemove: s.config.Optimization.AlwaysRemove,
	}

	requestAnalyzer, err := analyzer.NewAnalyzerFromRequest(&request, strategy, analyzerCfg)
	if err != nil {
		log.Printf("Error creating analyzer: %v", err)
		sendJSONResponse(w, &downstreamResponse)
		return
	}

	// Infer fields to keep using analyzer
	fieldsToKeep, err := requestAnalyzer.InferFields(&request)
	if err != nil {
		log.Printf("Error inferring fields: %v", err)
		// Fallback: return unfiltered response
		sendJSONResponse(w, &downstreamResponse)
		return
	}

	// Parse result data
	var resultData interface{}
	if err := json.Unmarshal(downstreamResponse.Result, &resultData); err != nil {
		log.Printf("Error parsing result data: %v", err)
		// Fallback: return unfiltered response
		sendJSONResponse(w, &downstreamResponse)
		return
	}

	// Apply filtering based on strategy
	var filteredData interface{}

	// Check if we're using explicit strategy with exclude list
	if request.Meta != nil && request.Meta.FieldSelection != nil && len(request.Meta.FieldSelection.Exclude) > 0 {
		// Use exclude filtering
		filteredData, err = filter.FilterWithExclude(ctx, resultData, request.Meta.FieldSelection.Exclude, 0, 10)
	} else if len(fieldsToKeep) > 0 && hasNestedFields(fieldsToKeep) {
		// Use nested field filtering for dot notation
		filteredData, err = filter.FilterNested(ctx, resultData, fieldsToKeep, 0, 10)
	} else {
		// Use standard filtering
		filteredData, err = s.filter.Filter(ctx, resultData, fieldsToKeep)
	}

	if err != nil {
		log.Printf("Error filtering data: %v", err)
		// Fallback: return unfiltered response if configured
		if s.config.Optimization.Safety.FallbackOnError {
			sendJSONResponse(w, &downstreamResponse)
			return
		}
		sendError(w, request.ID, types.ErrorCodeFilteringError, "Filtering failed", nil)
		return
	}

	// Calculate filtering statistics
	filteringDuration := time.Since(startTime)
	filteringStats := filter.CalculateFilteringStats(resultData, filteredData, effectiveStrategy, filteringDuration)

	// Record filtering metrics (T049)
	if s.metricsCollector != nil {
		toolName := "unknown"
		if request.Params != nil {
			toolName = request.Params.Name
		}

		stats := types.FilteringStatistics{
			TokensOriginal: filteringStats.OriginalSize,
			TokensFiltered: filteringStats.FilteredSize,
			TokensSaved:    filteringStats.OriginalSize - filteringStats.FilteredSize,
			SavingsPercent: filteringStats.Savings,
			FieldsKept:     len(filteringStats.FieldsKept),
			FieldsRemoved:  filteringStats.FieldsRemoved,
			Duration:       filteringDuration,
			Strategy:       effectiveStrategy,
			Server:         serverName,
			Tool:           toolName,
		}
		s.metricsCollector.RecordFiltering(stats)
		s.metricsCollector.RecordRequest(serverName, toolName, effectiveStrategy, "success")
	}

	// Marshal filtered data
	filteredJSON, err := json.Marshal(filteredData)
	if err != nil {
		log.Printf("Error marshaling filtered data: %v", err)
		sendJSONResponse(w, &downstreamResponse)
		return
	}

	// Build response with filtering metadata
	response := &types.MCPResponse{
		JSONRPC: "2.0",
		Result:  filteredJSON,
		ID:      request.ID,
		Meta: &types.ResponseMeta{
			FieldSelection:  &filteringStats,
			Version:         "0.1.0",
			ProtocolVersion: "1.0",
		},
	}

	sendJSONResponse(w, response)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Build server status information (T063) with actual health checks (T062)
	servers := make(map[string]interface{})
	ctx := r.Context()
	allHealthy := true

	for serverName, serverConfig := range s.config.Proxy.Servers {
		// Perform actual health check
		serverHealth := s.client.CheckServerHealth(ctx, serverConfig.URL)

		// Add configuration info
		serverHealth["strategy"] = s.config.GetStrategy(serverName)
		if serverConfig.Timeout > 0 {
			serverHealth["timeout"] = serverConfig.Timeout.String()
		}

		// Track overall health
		if available, ok := serverHealth["available"].(bool); ok && !available {
			allHealthy = false
		}

		servers[serverName] = serverHealth
	}

	// Determine overall status
	overallStatus := "ok"
	if !allHealthy {
		overallStatus = "degraded"
	}

	health := map[string]interface{}{
		"status":  overallStatus,
		"version": "0.1.0",
		"proxy":   "SlimMCP",
		"servers": servers,
	}

	// Return 503 if any server is unhealthy
	if !allHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// extractServerName extracts the server name from the URL path
func (s *Server) extractServerName(path string) string {
	// Remove leading slash
	path = strings.TrimPrefix(path, "/")

	// If empty or just "/", return empty (will use default if configured)
	if path == "" {
		// Return first configured server as default
		for serverName := range s.config.Proxy.Servers {
			return serverName
		}
		return ""
	}

	return path
}

// extractHeaders extracts relevant headers to forward
func extractHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)

	// Forward authentication headers
	if auth := headers.Get("Authorization"); auth != "" {
		result["Authorization"] = auth
	}

	// Forward API key headers (common variations)
	if apiKey := headers.Get("X-API-Key"); apiKey != "" {
		result["X-API-Key"] = apiKey
	}
	if apiKey := headers.Get("Api-Key"); apiKey != "" {
		result["Api-Key"] = apiKey
	}

	// Forward user agent
	if ua := headers.Get("User-Agent"); ua != "" {
		result["User-Agent"] = ua
	}

	return result
}

// sendError sends a JSON-RPC error response
func sendError(w http.ResponseWriter, id interface{}, code int, message string, data interface{}) {
	response := types.NewErrorResponse(id, code, message, data)
	sendJSONResponse(w, response)
}

// sendJSONResponse sends a JSON response
func sendJSONResponse(w http.ResponseWriter, response *types.MCPResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// hasNestedFields checks if any field uses dot notation
func hasNestedFields(fields []string) bool {
	for _, field := range fields {
		if strings.Contains(field, ".") {
			return true
		}
	}
	return false
}
