package util

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

// VCRMode determines how VCR operates
type VCRMode string

const (
	// VCRModeRecord records new HTTP interactions
	VCRModeRecord VCRMode = "record"
	// VCRModeReplay replays existing cassettes
	VCRModeReplay VCRMode = "replay"
	// VCRModeAuto uses replay if cassette exists, otherwise records
	VCRModeAuto VCRMode = "auto"
)

// VCRConfig configures VCR recorder behavior
type VCRConfig struct {
	// CassetteName is the name of the cassette file (without extension, govcr adds .yaml)
	CassetteName string
	// Mode determines record/replay behavior
	Mode VCRMode
	// CassetteDir is the directory where cassettes are stored
	CassetteDir string
	// SanitizeHeaders removes sensitive headers from recordings
	SanitizeHeaders bool
}

// NewVCRRecorder creates a new VCR recorder for testing
// Usage:
//
//	r, err := testutil.NewVCRRecorder(t, testutil.VCRConfig{
//	    CassetteName: "notion_query_database",
//	    Mode: testutil.VCRModeAuto,
//	})
//	defer r.Stop()
//	client := &http.Client{Transport: r}
func NewVCRRecorder(t *testing.T, cfg VCRConfig) (*recorder.Recorder, error) {
	// Set defaults
	if cfg.CassetteDir == "" {
		// Find project root
		root, err := findProjectRoot()
		if err != nil {
			return nil, fmt.Errorf("find project root: %w", err)
		}
		cfg.CassetteDir = filepath.Join(root, "tests", "cassettes")
	}

	if cfg.Mode == "" {
		cfg.Mode = getVCRModeFromEnv()
	}

	// Ensure cassette directory exists
	if err := os.MkdirAll(cfg.CassetteDir, 0755); err != nil {
		return nil, fmt.Errorf("create cassette directory: %w", err)
	}

	// Determine recorder mode (govcr v3 uses YAML cassettes)
	var recMode recorder.Mode
	cassettePath := filepath.Join(cfg.CassetteDir, cfg.CassetteName)

	switch cfg.Mode {
	case VCRModeRecord:
		recMode = recorder.ModeRecordOnly
		t.Logf("VCR: Recording to %s", cassettePath)
	case VCRModeReplay:
		recMode = recorder.ModeReplayOnly
		// govcr adds .yaml extension automatically
		if !fileExists(cassettePath) && !fileExists(cassettePath+".yaml") {
			return nil, fmt.Errorf("cassette not found: %s (set VCR_MODE=record to create it)", cassettePath)
		}
		t.Logf("VCR: Replaying from %s", cassettePath)
	case VCRModeAuto:
		// govcr adds .yaml extension automatically
		if fileExists(cassettePath) || fileExists(cassettePath+".yaml") {
			recMode = recorder.ModeReplayOnly
			t.Logf("VCR: Replaying from %s", cassettePath)
		} else {
			recMode = recorder.ModeRecordOnly
			t.Logf("VCR: Recording to %s (cassette not found)", cassettePath)
		}
	default:
		return nil, fmt.Errorf("invalid VCR mode: %s", cfg.Mode)
	}

	// Create recorder
	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       cassettePath,
		Mode:               recMode,
		SkipRequestLatency: true, // Don't simulate network latency in tests
	})
	if err != nil {
		return nil, fmt.Errorf("create recorder: %w", err)
	}

	// Sanitize sensitive headers
	if cfg.SanitizeHeaders {
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
	}

	return r, nil
}

// NewVCRHTTPClient creates an HTTP client with VCR transport
func NewVCRHTTPClient(t *testing.T, cfg VCRConfig) (*http.Client, func(), error) {
	r, err := NewVCRRecorder(t, cfg)
	if err != nil {
		return nil, nil, err
	}

	client := &http.Client{
		Transport: r,
	}

	cleanup := func() {
		if err := r.Stop(); err != nil {
			t.Errorf("Failed to stop VCR recorder: %v", err)
		}
	}

	return client, cleanup, nil
}

// getVCRModeFromEnv reads VCR mode from environment
func getVCRModeFromEnv() VCRMode {
	mode := os.Getenv("VCR_MODE")
	switch mode {
	case "record":
		return VCRModeRecord
	case "replay":
		return VCRModeReplay
	case "auto", "":
		return VCRModeAuto
	default:
		return VCRModeAuto
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// findProjectRoot finds the project root by looking for go.mod
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree
	for {
		if fileExists(filepath.Join(dir, "go.mod")) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}
		dir = parent
	}
}

// LoadEnvFile loads environment variables from a file
func LoadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("env file not found: %s (copy from .env.test.example)", path)
		}
		return fmt.Errorf("read env file: %w", err)
	}

	// Parse simple KEY=VALUE format
	lines := splitLines(string(data))
	for _, line := range lines {
		// Skip comments and empty lines
		if line == "" || line[0] == '#' {
			continue
		}

		// Parse KEY=VALUE
		parts := splitFirst(line, "=")
		if len(parts) != 2 {
			continue
		}

		key := trim(parts[0])
		value := trim(parts[1])

		// Only set if not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return nil
}

// Helper functions for simple string parsing

func splitLines(s string) []string {
	var lines []string
	var line string
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, line)
			line = ""
		} else if r != '\r' {
			line += string(r)
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func splitFirst(s, sep string) []string {
	idx := -1
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			idx = i
			break
		}
	}
	if idx == -1 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+len(sep):]}
}

func trim(s string) string {
	// Trim spaces and quotes
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '"' || s[start] == '\'') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '"' || s[end-1] == '\'') {
		end--
	}

	return s[start:end]
}
