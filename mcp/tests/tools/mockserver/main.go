package main

import (
	"flag"
	"log"

	"github.com/slimmcp/slimmcp/tests/util"
)

func main() {
	port := flag.Int("port", 3000, "Port to run the mock server on")
	flag.Parse()

	// Create mock server
	server := util.NewMockMCPServer()

	// Load all Notion fixtures
	log.Println("Loading Notion fixtures...")
	if err := util.SetupMockServerWithFixtures(server, "notion"); err != nil {
		log.Printf("Warning: Failed to load Notion fixtures: %v", err)
	}

	// Load all GitHub fixtures
	log.Println("Loading GitHub fixtures...")
	if err := util.SetupMockServerWithFixtures(server, "github"); err != nil {
		log.Printf("Warning: Failed to load GitHub fixtures: %v", err)
	}

	// Add some default responses for testing
	server.SetRawResponse("test-tool", `{"id":"test-123","type":"test","url":"http://example.com","title":"Test Response"}`)

	// Start server
	log.Printf("Starting mock MCP server on port %d...", *port)
	log.Printf("Try: curl http://localhost:%d -X POST -d '{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"test-tool\",\"arguments\":{}},\"id\":1}'", *port)

	if err := server.Start(*port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
