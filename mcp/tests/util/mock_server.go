package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// MockMCPServer is a mock MCP server for testing
type MockMCPServer struct {
	responses        map[string]json.RawMessage
	calls            []CallRecord
	capturedHeaders  map[string]string
	supportsCapabilities bool
	capabilityQueryCount int
	mutex            sync.Mutex
	server           *http.Server
}

// CallRecord records a call made to the mock server
type CallRecord struct {
	Method    string
	ToolName  string
	Arguments map[string]interface{}
	Headers   map[string]string
	Timestamp string
}

// NewMockMCPServer creates a new mock MCP server
func NewMockMCPServer() *MockMCPServer {
	return &MockMCPServer{
		responses:            make(map[string]json.RawMessage),
		calls:                make([]CallRecord, 0),
		capturedHeaders:      make(map[string]string),
		supportsCapabilities: false,
		capabilityQueryCount: 0,
	}
}

// SetSupportsCapabilities configures whether the server advertises field selection support
func (m *MockMCPServer) SetSupportsCapabilities(supports bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.supportsCapabilities = supports
}

// GetCapabilityQueryCount returns the number of times capabilities were queried
func (m *MockMCPServer) GetCapabilityQueryCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.capabilityQueryCount
}

// GetCapturedHeaders returns the headers captured from the last request
func (m *MockMCPServer) GetCapturedHeaders() map[string]string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Return a copy
	headers := make(map[string]string)
	for k, v := range m.capturedHeaders {
		headers[k] = v
	}
	return headers
}

// SetResponse configures a response for a specific tool
func (m *MockMCPServer) SetResponse(toolName string, response interface{}) error {
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.responses[toolName] = data
	return nil
}

// SetRawResponse configures a raw JSON response for a specific tool
func (m *MockMCPServer) SetRawResponse(toolName string, responseJSON string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.responses[toolName] = json.RawMessage(responseJSON)
}

// ServeHTTP implements http.Handler
func (m *MockMCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle GET requests for health checks
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	// Only accept POST requests for MCP operations
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Capture headers (will be stored for tool calls only, after we parse the request)
	currentHeaders := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			currentHeaders[key] = values[0]
		}
	}

	// Parse request
	var req struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		} `json:"params"`
		Meta *struct {
			FieldSelection *struct {
				Include  []string `json:"include,omitempty"`
				Exclude  []string `json:"exclude,omitempty"`
				Strategy string   `json:"strategy,omitempty"`
			} `json:"fieldSelection,omitempty"`
		} `json:"_meta,omitempty"`
		ID interface{} `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Mock server: failed to decode request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Mock server: received method=%s tool=%s", req.Method, req.Params.Name)

	// Handle capability detection queries
	if req.Method == "initialize" || req.Method == "ping" {
		m.mutex.Lock()
		m.capabilityQueryCount++
		supportsCapabilities := m.supportsCapabilities
		m.mutex.Unlock()

		capabilities := map[string]interface{}{
			"protocolVersion": "1.0",
			"capabilities": map[string]interface{}{
				"fieldSelection": map[string]interface{}{
					"supported": supportsCapabilities,
				},
			},
		}

		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"result":  capabilities,
			"id":      req.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Handle tools/list requests (capability detection) without capturing headers
	if req.Method == "tools/list" {
		m.mutex.Lock()
		supportsCapabilities := m.supportsCapabilities
		m.mutex.Unlock()

		toolsList := map[string]interface{}{
			"tools": []interface{}{},
		}

		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"result":  toolsList,
			"id":      req.ID,
		}

		// Add capability metadata if server supports capabilities
		if supportsCapabilities {
			response["_meta"] = map[string]interface{}{
				"version":         "1.0",
				"protocolVersion": "1.0",
				"capabilities": map[string]interface{}{
					"fieldSelection": map[string]interface{}{
						"supported":  true,
						"strategies": []string{"explicit", "minimal"},
					},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Capture headers for actual tool calls only (not capability queries)
	m.mutex.Lock()
	m.capturedHeaders = currentHeaders
	m.mutex.Unlock()

	// Record call
	m.recordCall(req.Method, req.Params.Name, req.Params.Arguments, currentHeaders)

	// Get response
	m.mutex.Lock()
	responseData, ok := m.responses[req.Params.Name]
	supportsCapabilities := m.supportsCapabilities
	m.mutex.Unlock()

	if !ok {
		// Return default response
		responseData = json.RawMessage(`{"id": "test-123", "type": "test", "url": "http://example.com", "title": "Test Response"}`)
	}

	// If the response is a map, try to inject the ID from the request arguments
	var responseMap map[string]interface{}
	if err := json.Unmarshal(responseData, &responseMap); err == nil {
		// If request has an ID in arguments, use it
		if argID, ok := req.Params.Arguments["id"].(string); ok && argID != "" {
			responseMap["id"] = argID
		}
		// Re-marshal
		responseData, _ = json.Marshal(responseMap)
	}

	// If server supports native field selection and request has _meta, apply filtering
	var responseMeta map[string]interface{}
	if supportsCapabilities && req.Meta != nil && req.Meta.FieldSelection != nil {
		// Server handles field selection natively (simplified mock behavior)
		// Parse the response to filter fields based on request
		var resultMap map[string]interface{}
		if err := json.Unmarshal(responseData, &resultMap); err == nil {
			// Apply filtering based on include list
			if len(req.Meta.FieldSelection.Include) > 0 {
				filteredResult := make(map[string]interface{})
				for _, field := range req.Meta.FieldSelection.Include {
					if val, ok := resultMap[field]; ok {
						filteredResult[field] = val
					}
				}
				responseData, _ = json.Marshal(filteredResult)
			}
		}

		// Add metadata indicating server-side filtering was applied
		responseMeta = map[string]interface{}{
			"fieldSelection": map[string]interface{}{
				"applied":    true,
				"filteredBy": "server",
				"strategy":   req.Meta.FieldSelection.Strategy,
			},
		}
	}

	// Build JSON-RPC response
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"result":  responseData,
		"id":      req.ID,
	}

	// Add metadata if present
	if responseMeta != nil {
		response["_meta"] = responseMeta
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// recordCall records a call to the mock server
func (m *MockMCPServer) recordCall(method, toolName string, arguments map[string]interface{}, headers map[string]string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.calls = append(m.calls, CallRecord{
		Method:    method,
		ToolName:  toolName,
		Arguments: arguments,
		Headers:   headers,
		Timestamp: "",
	})
}

// GetCalls returns all recorded calls
func (m *MockMCPServer) GetCalls() []CallRecord {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Return copy
	calls := make([]CallRecord, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// GetCallCount returns the number of calls made
func (m *MockMCPServer) GetCallCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.calls)
}

// GetCallsForTool returns calls for a specific tool
func (m *MockMCPServer) GetCallsForTool(toolName string) []CallRecord {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var toolCalls []CallRecord
	for _, call := range m.calls {
		if call.ToolName == toolName {
			toolCalls = append(toolCalls, call)
		}
	}
	return toolCalls
}

// Reset clears all recorded calls
func (m *MockMCPServer) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.calls = make([]CallRecord, 0)
}

// Start starts the mock server on the specified port
func (m *MockMCPServer) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	m.server = &http.Server{
		Addr:    addr,
		Handler: m,
	}

	log.Printf("Mock MCP server starting on %s", addr)
	return m.server.ListenAndServe()
}

// Stop stops the mock server
func (m *MockMCPServer) Stop() error {
	if m.server != nil {
		return m.server.Close()
	}
	return nil
}

// Helper functions for creating test responses

// CreateNotionPageResponse creates a mock Notion page response with many fields
func CreateNotionPageResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":               "page-123",
		"type":             "page",
		"url":              "https://notion.so/page-123",
		"title":            "Test Page",
		"created_time":     "2024-01-01T00:00:00Z",
		"last_edited_time": "2024-01-02T00:00:00Z",
		"archived":         false,
		"icon":             map[string]interface{}{"type": "emoji", "emoji": "📄"},
		"cover":            map[string]interface{}{"type": "external", "url": "https://example.com/cover.jpg"},
		"properties": map[string]interface{}{
			"Status": map[string]interface{}{
				"id":     "status",
				"type":   "select",
				"select": map[string]interface{}{"name": "In Progress", "color": "blue"},
			},
			"Assignee": map[string]interface{}{
				"id":   "assignee",
				"type": "people",
				"people": []interface{}{
					map[string]interface{}{"id": "user-1", "name": "John Doe", "email": "john@example.com"},
				},
			},
			"Priority": map[string]interface{}{
				"id":     "priority",
				"type":   "select",
				"select": map[string]interface{}{"name": "High", "color": "red"},
			},
			"Tags": map[string]interface{}{
				"id":   "tags",
				"type": "multi_select",
				"multi_select": []interface{}{
					map[string]interface{}{"name": "Backend", "color": "blue"},
					map[string]interface{}{"name": "Urgent", "color": "red"},
				},
			},
			"Description": map[string]interface{}{
				"id":   "description",
				"type": "rich_text",
				"rich_text": []interface{}{
					map[string]interface{}{"text": map[string]interface{}{"content": "This is a test page description"}},
				},
			},
		},
		"created_by": map[string]interface{}{
			"id":    "user-1",
			"name":  "John Doe",
			"email": "john@example.com",
		},
		"last_edited_by": map[string]interface{}{
			"id":    "user-2",
			"name":  "Jane Smith",
			"email": "jane@example.com",
		},
		"parent": map[string]interface{}{
			"type":      "workspace",
			"workspace": true,
		},
	}
}

// CreateSlackMessageResponse creates a mock Slack message response
func CreateSlackMessageResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":   "msg-123",
		"type": "message",
		"url":  "https://slack.com/archives/C123456/p1234567890123456",
		"text": "Hello team! I wanted to share an update on our Q4 roadmap. We've made significant progress on the user authentication system and the new dashboard features. Please review the attached document and provide feedback by end of week. Let me know if you have any questions or concerns about the timeline.",
		"user":    "U123456",
		"ts":      "1234567890.123456",
		"channel": "C123456",
		"team":    "T123456",
		"reactions": []interface{}{
			map[string]interface{}{"name": "thumbsup", "count": 5, "users": []string{"U001", "U002", "U003"}},
			map[string]interface{}{"name": "heart", "count": 3, "users": []string{"U004", "U005"}},
			map[string]interface{}{"name": "rocket", "count": 2, "users": []string{"U006"}},
		},
		"thread_ts":   "1234567890.123456",
		"reply_count": 12,
		"reply_users": []string{"U789", "U456", "U321"},
		"reply_users_count": 3,
		"latest_reply": "1234567899.123456",
		"is_starred": true,
		"pinned_to": []string{"C123456"},
		"pinned_info": map[string]interface{}{
			"channel": "C123456",
			"pinned_by": "U123456",
			"pinned_ts": 1234567891,
		},
		"attachments": []interface{}{
			map[string]interface{}{
				"id":         1,
				"color":      "36a64f",
				"fallback":   "Q4 Roadmap Document",
				"title":      "Q4 Roadmap - Engineering Team",
				"title_link": "https://docs.example.com/roadmap",
				"text":       "Complete roadmap for Q4 2024 including feature priorities, resource allocation, and timeline estimates",
				"fields": []interface{}{
					map[string]interface{}{"title": "Status", "value": "In Progress", "short": true},
					map[string]interface{}{"title": "Owner", "value": "Engineering Team", "short": true},
					map[string]interface{}{"title": "Due Date", "value": "Dec 31, 2024", "short": true},
					map[string]interface{}{"title": "Priority", "value": "High", "short": true},
				},
				"image_url": "https://example.com/roadmap-preview.png",
				"thumb_url": "https://example.com/roadmap-thumb.png",
				"footer":    "Shared from Google Docs",
				"footer_icon": "https://example.com/gdocs-icon.png",
				"ts": 1234567890,
			},
		},
		"files": []interface{}{
			map[string]interface{}{
				"id":        "F123456",
				"name":      "roadmap.pdf",
				"title":     "Q4 Roadmap PDF",
				"mimetype":  "application/pdf",
				"filetype":  "pdf",
				"size":      245678,
				"url_private": "https://files.slack.com/files-pri/T123456/roadmap.pdf",
				"permalink": "https://example.slack.com/files/U123456/F123456/roadmap.pdf",
			},
		},
		"blocks": []interface{}{
			map[string]interface{}{
				"type": "section",
				"text": map[string]interface{}{
					"type": "mrkdwn",
					"text": "*Q4 Roadmap Update* :rocket:",
				},
			},
			map[string]interface{}{
				"type": "divider",
			},
			map[string]interface{}{
				"type": "section",
				"fields": []interface{}{
					map[string]interface{}{"type": "mrkdwn", "text": "*Priority:*\nHigh"},
					map[string]interface{}{"type": "mrkdwn", "text": "*Due Date:*\nDec 31"},
				},
			},
		},
	}
}

// CreateGDriveFileResponse creates a mock Google Drive file response
func CreateGDriveFileResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":           "file-123",
		"type":         "file",
		"url":          "https://drive.google.com/file/d/file-123/view",
		"name":         "Test Document",
		"mimeType":     "application/vnd.google-apps.document",
		"webViewLink":  "https://drive.google.com/file/123",
		"createdTime":  "2024-01-01T00:00:00Z",
		"modifiedTime": "2024-01-02T00:00:00Z",
		"size":         "12345",
		"owners": []interface{}{
			map[string]interface{}{"displayName": "John Doe", "emailAddress": "john@example.com"},
		},
		"permissions": []interface{}{
			map[string]interface{}{"type": "user", "role": "owner"},
		},
		"thumbnailLink": "https://drive.google.com/thumbnail/123",
		"iconLink":      "https://drive.google.com/icon/123",
	}
}
