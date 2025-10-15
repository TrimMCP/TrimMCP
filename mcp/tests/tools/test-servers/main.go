package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// TestServer provides HTTP, HTTPS, and SSE endpoints for MCP testing
type TestServer struct {
	httpPort  int
	httpsPort int
	ssePort   int

	// SSE connections
	sseClients map[string]chan MCPResponse
	mu         sync.Mutex
}

// MCPRequest represents an MCP JSON-RPC request
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

// MCPResponse represents an MCP JSON-RPC response
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
	ID      interface{}     `json:"id"`
}

// MCPError represents a JSON-RPC error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	httpPort := flag.Int("http", 3200, "HTTP port")
	httpsPort := flag.Int("https", 3201, "HTTPS port")
	ssePort := flag.Int("sse", 3202, "SSE port")
	flag.Parse()

	server := &TestServer{
		httpPort:   *httpPort,
		httpsPort:  *httpsPort,
		ssePort:    *ssePort,
		sseClients: make(map[string]chan MCPResponse),
	}

	// Start HTTP server
	go func() {
		log.Printf("Starting HTTP server on :%d", server.httpPort)
		if err := server.startHTTP(); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Start HTTPS server
	go func() {
		log.Printf("Starting HTTPS server on :%d", server.httpsPort)
		if err := server.startHTTPS(); err != nil {
			log.Printf("HTTPS server error: %v", err)
		}
	}()

	// Start SSE server
	go func() {
		log.Printf("Starting SSE server on :%d (GET for events, POST for commands)", server.ssePort)
		if err := server.startSSE(); err != nil {
			log.Printf("SSE server error: %v", err)
		}
	}()

	log.Println("Test servers ready!")
	select {} // Block forever
}

func (s *TestServer) startHTTP() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleMCP)

	return http.ListenAndServe(fmt.Sprintf(":%d", s.httpPort), mux)
}

func (s *TestServer) startHTTPS() error {
	// Generate self-signed certificate
	cert, err := generateSelfSignedCert()
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleMCP)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.httpsPort),
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	return server.ListenAndServeTLS("", "")
}

func (s *TestServer) startSSE() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.handleSSEConnect(w, r)
		} else if r.Method == http.MethodPost {
			s.handleSSECommand(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", s.ssePort), mux)
}

func (s *TestServer) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, nil, -32700, "Failed to read request")
		return
	}

	var req MCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, nil, -32700, "Parse error")
		return
	}

	// Handle based on method
	var response MCPResponse
	switch req.Method {
	case "initialize":
		response = MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: json.RawMessage(`{
				"protocolVersion": "2024-11-05",
				"capabilities": {},
				"serverInfo": {
					"name": "test-server",
					"version": "0.1.0"
				}
			}`),
		}

	case "tools/list":
		response = MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: json.RawMessage(`{
				"tools": [
					{
						"name": "test_tool",
						"description": "A test tool",
						"inputSchema": {
							"type": "object",
							"properties": {}
						}
					}
				]
			}`),
		}

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			sendError(w, req.ID, -32602, "Invalid params")
			return
		}

		result := map[string]interface{}{
			"success": true,
			"tool":    params.Name,
			"echo":    params.Arguments,
		}
		resultJSON, _ := json.Marshal(result)

		response = MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  resultJSON,
		}

	case "ping":
		response = MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"status": "ok"}`),
		}

	default:
		sendError(w, req.ID, -32601, "Method not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *TestServer) handleSSEConnect(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create client channel
	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	responseChan := make(chan MCPResponse, 10)

	s.mu.Lock()
	s.sseClients[clientID] = responseChan
	s.mu.Unlock()

	// Remove client on disconnect
	defer func() {
		s.mu.Lock()
		delete(s.sseClients, clientID)
		close(responseChan)
		s.mu.Unlock()
	}()

	// Send initial connection message
	fmt.Fprintf(w, "data: {\"type\":\"connected\",\"clientId\":\"%s\"}\n\n", clientID)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Send responses as they arrive
	for response := range responseChan {
		data, _ := json.Marshal(response)
		fmt.Fprintf(w, "data: %s\n\n", data)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func (s *TestServer) handleSSECommand(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}

	var req MCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Parse error", http.StatusBadRequest)
		return
	}

	// Process request (same logic as handleMCP for consistency)
	var response MCPResponse
	switch req.Method {
	case "initialize":
		response = MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: json.RawMessage(`{
				"protocolVersion": "2024-11-05",
				"capabilities": {},
				"serverInfo": {
					"name": "test-server",
					"version": "0.1.0"
				}
			}`),
		}
	case "tools/list":
		response = MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: json.RawMessage(`{
				"tools": [
					{
						"name": "test_tool",
						"description": "A test tool",
						"inputSchema": {
							"type": "object",
							"properties": {}
						}
					}
				]
			}`),
		}
	case "ping":
		response = MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"status": "ok"}`),
		}
	default:
		response = MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &MCPError{Code: -32601, Message: "Method not found"},
		}
	}

	// Broadcast response to all SSE clients
	s.mu.Lock()
	for _, client := range s.sseClients {
		select {
		case client <- response:
		default:
			// Client buffer full, skip
		}
	}
	s.mu.Unlock()

	// Also return HTTP 202 Accepted
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("OK"))
}

func sendError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateSelfSignedCert() (tls.Certificate, error) {
	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Create certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"SlimMCP Test"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Encode certificate and key to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return tls.X509KeyPair(certPEM, keyPEM)
}
