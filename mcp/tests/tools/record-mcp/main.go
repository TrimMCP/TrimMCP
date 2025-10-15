package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

// RecordRequest represents an MCP request to record
type RecordRequest struct {
	Server    string                 `json:"server"`
	Operation string                 `json:"operation"`
	URL       string                 `json:"url"`
	Method    string                 `json:"method"`
	Params    map[string]interface{} `json:"params"`
}

func main() {
	var (
		envFile    = flag.String("env", ".env.test", "Path to environment file")
		configFile = flag.String("config", "tests/recording-scenarios.json", "Path to recording scenarios config")
		outputDir  = flag.String("output", "tests/cassettes", "Directory to save cassettes")
		mode       = flag.String("mode", "record", "VCR mode: record, replay, or auto")
	)
	flag.Parse()

	// Load environment variables
	if err := loadEnvFile(*envFile); err != nil {
		log.Fatalf("Failed to load env file: %v", err)
	}

	// Load recording scenarios
	scenarios, err := loadScenarios(*configFile)
	if err != nil {
		log.Fatalf("Failed to load scenarios: %v", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	log.Printf("Recording %d scenarios to %s", len(scenarios), *outputDir)

	// Record each scenario
	for i, scenario := range scenarios {
		log.Printf("[%d/%d] Recording %s/%s", i+1, len(scenarios), scenario.Server, scenario.Operation)

		if err := recordScenario(scenario, *outputDir, *mode); err != nil {
			log.Printf("  ERROR: %v", err)
			continue
		}

		log.Printf("  SUCCESS: Saved to %s/%s.yaml", *outputDir, cassetteName(scenario))
	}

	log.Println("Recording complete!")
}

func recordScenario(scenario RecordRequest, outputDir, mode string) error {
	// Create recorder (govcr v3 uses YAML format)
	cassettePath := filepath.Join(outputDir, cassetteName(scenario))

	var recMode recorder.Mode
	switch mode {
	case "record":
		recMode = recorder.ModeRecordOnly
	case "replay":
		recMode = recorder.ModeReplayOnly
	case "auto":
		if _, err := os.Stat(cassettePath); err == nil {
			recMode = recorder.ModeReplayOnly
		} else {
			recMode = recorder.ModeRecordOnly
		}
	default:
		return fmt.Errorf("invalid mode: %s", mode)
	}

	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       cassettePath,
		Mode:               recMode,
		SkipRequestLatency: true,
	})
	if err != nil {
		return fmt.Errorf("create recorder: %w", err)
	}
	defer r.Stop()

	// Sanitize sensitive headers before saving
	r.AddHook(func(i *cassette.Interaction) error {
		// Remove authorization tokens
		delete(i.Request.Headers, "Authorization")
		delete(i.Request.Headers, "Notion-Version")
		delete(i.Response.Headers, "Set-Cookie")

		// Sanitize cookies
		if cookies := i.Request.Headers.Values("Cookie"); len(cookies) > 0 {
			i.Request.Headers.Set("Cookie", "REDACTED")
		}

		return nil
	}, recorder.BeforeSaveHook)

	// Build MCP request
	mcpRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  scenario.Method,
		"params":  scenario.Params,
		"id":      1,
	}

	reqBody, err := json.Marshal(mcpRequest)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", scenario.URL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Set headers based on server type
	httpReq.Header.Set("Content-Type", "application/json")
	if err := setAuthHeaders(httpReq, scenario.Server); err != nil {
		return fmt.Errorf("set auth headers: %w", err)
	}

	// Execute request with VCR transport
	client := &http.Client{Transport: r}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Verify response is valid JSON
	var jsonCheck interface{}
	if err := json.Unmarshal(respBody, &jsonCheck); err != nil {
		return fmt.Errorf("invalid JSON response: %w", err)
	}

	return nil
}

func setAuthHeaders(req *http.Request, server string) error {
	switch strings.ToLower(server) {
	case "notion":
		// When using localhost bridge, auth is handled by the MCP server
		// No headers needed for localhost
		if !strings.Contains(req.URL.String(), "localhost") {
			token := os.Getenv("NOTION_TOKEN")
			if token == "" {
				return fmt.Errorf("NOTION_TOKEN not set")
			}
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Notion-Version", "2022-06-28")
		}

	case "github":
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return fmt.Errorf("GITHUB_TOKEN not set")
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

	default:
		return fmt.Errorf("unknown server: %s", server)
	}

	return nil
}

func cassetteName(scenario RecordRequest) string {
	return fmt.Sprintf("%s_%s", scenario.Server, scenario.Operation)
}

func loadScenarios(path string) ([]RecordRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var scenarios []RecordRequest
	if err := json.Unmarshal(data, &scenarios); err != nil {
		return nil, err
	}

	return scenarios, nil
}

func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("env file not found: %s (copy from .env.test.example)", path)
		}
		return err
	}

	// Parse simple KEY=VALUE format
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

		// Only set if not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return nil
}
