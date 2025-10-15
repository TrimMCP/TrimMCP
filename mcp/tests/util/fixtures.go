package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TestFixture represents a test fixture from POC recordings
type TestFixture struct {
	Operation  string          `json:"operation"`
	Server     string          `json:"server"`
	Request    json.RawMessage `json:"request"`
	Response   json.RawMessage `json:"response"`
	CapturedAt string          `json:"captured_at"`
}

// LoadFixture loads a test fixture from the fixtures directory
func LoadFixture(server, operation string) (*TestFixture, error) {
	// Try to find fixtures directory
	fixturesDir := findFixturesDir()
	if fixturesDir == "" {
		return nil, fmt.Errorf("fixtures directory not found")
	}

	filename := filepath.Join(fixturesDir, server, operation+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixture %s/%s: %w", server, operation, err)
	}

	var fixture TestFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		return nil, fmt.Errorf("failed to parse fixture: %w", err)
	}

	return &fixture, nil
}

// findFixturesDir finds the fixtures directory relative to the current location
func findFixturesDir() string {
	// Try common paths
	paths := []string{
		"tests/fixtures",
		"../tests/fixtures",
		"../../tests/fixtures",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// LoadAllFixtures loads all fixtures for a specific server
func LoadAllFixtures(server string) (map[string]*TestFixture, error) {
	fixturesDir := findFixturesDir()
	if fixturesDir == "" {
		return nil, fmt.Errorf("fixtures directory not found")
	}

	serverDir := filepath.Join(fixturesDir, server)
	entries, err := os.ReadDir(serverDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read server fixtures: %w", err)
	}

	fixtures := make(map[string]*TestFixture)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		operation := entry.Name()[:len(entry.Name())-5] // Remove .json
		fixture, err := LoadFixture(server, operation)
		if err != nil {
			return nil, err
		}

		fixtures[operation] = fixture
	}

	return fixtures, nil
}

// GetResponseContent extracts the response content from a fixture
func (f *TestFixture) GetResponseContent() (json.RawMessage, error) {
	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(f.Response, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// The text field contains the actual JSON response
	return json.RawMessage(response.Content[0].Text), nil
}

// SetupMockServerWithFixtures configures a mock server with all fixtures for a server
func SetupMockServerWithFixtures(mockServer *MockMCPServer, serverName string) error {
	fixtures, err := LoadAllFixtures(serverName)
	if err != nil {
		return fmt.Errorf("failed to load fixtures: %w", err)
	}

	for operation, fixture := range fixtures {
		content, err := fixture.GetResponseContent()
		if err != nil {
			return fmt.Errorf("failed to get response content for %s: %w", operation, err)
		}

		mockServer.SetRawResponse(operation, string(content))
	}

	return nil
}
