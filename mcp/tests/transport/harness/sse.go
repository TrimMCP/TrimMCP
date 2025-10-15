package harness

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SSETransport implements Transport interface for SSE (Server-Sent Events) based MCP servers
type SSETransport struct {
	name         string
	endpoint     string
	postEndpoint string // Endpoint for sending requests (if different from SSE endpoint)
	client       *http.Client
	headers      map[string]string

	sseConn   *http.Response
	responses chan response
	mu        sync.Mutex
	running   bool
}

// NewSSETransport creates a new SSE transport adapter
func NewSSETransport(name, sseEndpoint, postEndpoint string, headers map[string]string) *SSETransport {
	if postEndpoint == "" {
		postEndpoint = sseEndpoint
	}

	return &SSETransport{
		name:         name,
		endpoint:     sseEndpoint,
		postEndpoint: postEndpoint,
		headers:      headers,
		responses:    make(chan response, 100),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the transport type
func (s *SSETransport) Name() string {
	return fmt.Sprintf("sse-%s", s.name)
}

// Start establishes SSE connection
func (s *SSETransport) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("transport already running")
	}

	// Create SSE connection request
	req, err := http.NewRequestWithContext(ctx, "GET", s.endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	for key, value := range s.headers {
		req.Header.Set(key, value)
	}

	// Establish SSE connection
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to establish SSE connection: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("unexpected SSE status: %d", resp.StatusCode)
	}

	s.sseConn = resp
	s.running = true

	// Start reading SSE events
	go s.readSSEEvents()

	// Wait a moment for connection to stabilize
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Stop closes the SSE connection
func (s *SSETransport) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.sseConn != nil {
		s.sseConn.Body.Close()
	}

	s.running = false
	close(s.responses)

	return nil
}

// Execute sends an MCP request and waits for response via SSE
func (s *SSETransport) Execute(ctx context.Context, request json.RawMessage) (json.RawMessage, error) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil, fmt.Errorf("transport not running")
	}
	s.mu.Unlock()

	// Parse request to get ID
	var req MCPRequest
	if err := json.Unmarshal(request, &req); err != nil {
		return nil, fmt.Errorf("invalid request format: %w", err)
	}

	// Send request via POST
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.postEndpoint, bytes.NewReader(request))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range s.headers {
		httpReq.Header.Set(key, value)
	}

	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("unexpected HTTP status: %d", httpResp.StatusCode)
	}

	// Wait for response from SSE channel
	timeout := 5 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			return nil, fmt.Errorf("request timeout after %v", timeout)
		case resp := <-s.responses:
			// Check if this response matches our request ID
			if resp.ID == req.ID {
				if resp.Err != nil {
					return nil, resp.Err
				}
				return resp.Data, nil
			}
			// Not our response, keep waiting
		}
	}
}

// IsHealthy checks if the SSE connection is operational
func (s *SSETransport) IsHealthy(ctx context.Context) bool {
	s.mu.Lock()
	running := s.running
	s.mu.Unlock()

	if !running {
		return false
	}

	// Try a simple ping request
	pingReq := MustMarshal(MCPRequest{
		JSONRPC: "2.0",
		ID:      "health-check",
		Method:  "ping",
	})

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := s.Execute(ctx, pingReq)
	return err == nil
}

// readSSEEvents continuously reads SSE events and dispatches responses
func (s *SSETransport) readSSEEvents() {
	scanner := bufio.NewScanner(s.sseConn.Body)
	var eventData strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// SSE protocol: empty line means end of event
		if line == "" {
			if eventData.Len() > 0 {
				s.processSSEEvent(eventData.String())
				eventData.Reset()
			}
			continue
		}

		// Parse SSE field
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			if eventData.Len() > 0 {
				eventData.WriteString("\n")
			}
			eventData.WriteString(data)
		}
		// Ignore other SSE fields (event:, id:, retry:, etc.)
	}

	// Scanner finished (connection closed or error)
	// Don't send if we're already stopped (channel closed)
	s.mu.Lock()
	running := s.running
	s.mu.Unlock()

	if running && scanner.Err() != nil {
		select {
		case s.responses <- response{
			Err: fmt.Errorf("SSE read error: %w", scanner.Err()),
		}:
		default:
			// Channel full or closed, ignore
		}
	}
}

// processSSEEvent parses and dispatches a complete SSE event
func (s *SSETransport) processSSEEvent(data string) {
	// Check if we're still running
	s.mu.Lock()
	running := s.running
	s.mu.Unlock()

	if !running {
		return
	}

	// Try to parse as JSON-RPC response
	var resp MCPResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		select {
		case s.responses <- response{
			Err: fmt.Errorf("failed to parse SSE event: %w", err),
		}:
		default:
			// Channel full or closed, ignore
		}
		return
	}

	// Send successful response
	select {
	case s.responses <- response{
		ID:   resp.ID,
		Data: json.RawMessage(data),
	}:
	default:
		// Channel full or closed, ignore
	}
}
