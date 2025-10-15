package harness

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// StdioTransport implements Transport interface for stdio-based MCP servers
type StdioTransport struct {
	name       string
	serverCmd  string
	serverArgs []string
	env        []string

	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    *bufio.Scanner
	stderr    io.ReadCloser
	responses chan response
	mu        sync.Mutex
	running   bool
}

type response struct {
	ID   interface{}
	Data json.RawMessage
	Err  error
}

// NewStdioTransport creates a new stdio transport adapter
func NewStdioTransport(name, serverCmd string, args []string, env []string) *StdioTransport {
	return &StdioTransport{
		name:       name,
		serverCmd:  serverCmd,
		serverArgs: args,
		env:        env,
		responses:  make(chan response, 100),
	}
}

// Name returns the transport type
func (s *StdioTransport) Name() string {
	return fmt.Sprintf("stdio-%s", s.name)
}

// Start launches the MCP server process
func (s *StdioTransport) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("transport already running")
	}

	// Create command
	s.cmd = exec.CommandContext(ctx, s.serverCmd, s.serverArgs...)
	if len(s.env) > 0 {
		s.cmd.Env = s.env
	}

	// Setup pipes
	stdin, err := s.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	s.stdin = stdin

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	s.stdout = bufio.NewScanner(stdout)

	stderr, err := s.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	s.stderr = stderr

	// Start server
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Start response reader goroutine
	go s.readResponses()

	s.running = true

	// Wait longer for server to initialize (npx needs time to download packages)
	time.Sleep(3 * time.Second)

	return nil
}

// Stop terminates the MCP server process
func (s *StdioTransport) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// Close stdin to signal shutdown
	if s.stdin != nil {
		s.stdin.Close()
	}

	// Give server time to shutdown gracefully
	done := make(chan error, 1)
	go func() {
		done <- s.cmd.Wait()
	}()

	select {
	case <-time.After(5 * time.Second):
		// Force kill if graceful shutdown fails
		if s.cmd.Process != nil {
			s.cmd.Process.Kill()
		}
	case <-done:
		// Clean shutdown
	}

	s.running = false
	close(s.responses)

	return nil
}

// Execute sends an MCP request via stdin and waits for response on stdout
func (s *StdioTransport) Execute(ctx context.Context, request json.RawMessage) (json.RawMessage, error) {
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

	// Write request to stdin
	if _, err := s.stdin.Write(append(request, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Wait for response with timeout
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

// IsHealthy checks if the transport is operational
func (s *StdioTransport) IsHealthy(ctx context.Context) bool {
	s.mu.Lock()
	running := s.running
	s.mu.Unlock()

	if !running {
		return false
	}

	// Try a simple initialize request (ping might not be supported)
	initReq := MustMarshal(MCPRequest{
		JSONRPC: "2.0",
		ID:      "health-check",
		Method:  "initialize",
		Params: MustMarshal(map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "slimmcp-test",
				"version": "0.1.0",
			},
		}),
	})

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.Execute(ctx, initReq)
	return err == nil
}

// readResponses continuously reads from stdout and dispatches responses
func (s *StdioTransport) readResponses() {
	for s.stdout.Scan() {
		line := s.stdout.Bytes()
		if len(line) == 0 {
			continue
		}

		// Parse JSON-RPC response
		var resp MCPResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			// Send error response
			s.responses <- response{
				Err: fmt.Errorf("failed to parse response: %w", err),
			}
			continue
		}

		// Send successful response
		s.responses <- response{
			ID:   resp.ID,
			Data: line,
		}
	}

	// Scanner finished (EOF or error)
	if err := s.stdout.Err(); err != nil {
		s.responses <- response{
			Err: fmt.Errorf("stdout read error: %w", err),
		}
	}
}
