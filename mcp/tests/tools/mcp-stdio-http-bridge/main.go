package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// MCPStdioHTTPBridge wraps a stdio-based MCP server with an HTTP interface
//
// MCP servers can use three transport types:
// 1. stdio (stdin/stdout) - Most common for CLI tools (GitHub, Notion, etc.)
// 2. SSE (Server-Sent Events) - For streaming real-time updates
// 3. HTTP - Direct HTTP/REST endpoints
//
// This bridge converts stdio-based servers to HTTP for easier testing with
// tools like govcr, curl, and integration tests.
//
// Usage:
//   go run main.go -port 3100 -cmd npx -args "-y,@modelcontextprotocol/server-github"
//   go run main.go -port 3101 -cmd npx -args "-y,@modelcontextprotocol/server-notion"
func main() {
	var (
		port    = flag.Int("port", 3100, "HTTP server port")
		command = flag.String("cmd", "npx", "MCP server command")
		args    = flag.String("args", "-y,@modelcontextprotocol/server-github", "Comma-separated command arguments")
		envVars = flag.String("env", "", "Comma-separated KEY=VALUE environment variables")
	)
	flag.Parse()

	log.Printf("Starting MCP HTTP bridge on port %d", *port)
	log.Printf("Command: %s %s", *command, *args)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read request body
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Read error: %v", err), http.StatusBadRequest)
			return
		}

		// Validate JSON-RPC
		var jsonCheck map[string]interface{}
		if err := json.Unmarshal(reqBody, &jsonCheck); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		// Parse args
		var cmdArgs []string
		if *args != "" {
			for i := 0; i < len(*args); i++ {
				start := i
				for i < len(*args) && (*args)[i] != ',' {
					i++
				}
				cmdArgs = append(cmdArgs, (*args)[start:i])
			}
		}

		// Start MCP server process
		cmd := exec.Command(*command, cmdArgs...)

		// Copy environment variables
		cmd.Env = os.Environ()

		// Add custom environment variables from flag
		if *envVars != "" {
			for i := 0; i < len(*envVars); i++ {
				start := i
				for i < len(*envVars) && (*envVars)[i] != ',' {
					i++
				}
				envVar := (*envVars)[start:i]
				if envVar != "" {
					cmd.Env = append(cmd.Env, envVar)
				}
			}
		}

		// Set up pipes
		stdin, err := cmd.StdinPipe()
		if err != nil {
			http.Error(w, fmt.Sprintf("Stdin pipe error: %v", err), http.StatusInternalServerError)
			return
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			http.Error(w, fmt.Sprintf("Stdout pipe error: %v", err), http.StatusInternalServerError)
			return
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			http.Error(w, fmt.Sprintf("Stderr pipe error: %v", err), http.StatusInternalServerError)
			return
		}

		// Start the process
		if err := cmd.Start(); err != nil {
			http.Error(w, fmt.Sprintf("Start error: %v", err), http.StatusInternalServerError)
			return
		}

		// Write request to stdin
		_, err = stdin.Write(reqBody)
		if err != nil {
			http.Error(w, fmt.Sprintf("Write error: %v", err), http.StatusInternalServerError)
			cmd.Process.Kill()
			return
		}
		stdin.Write([]byte("\n"))
		stdin.Close()

		// Read stderr in background (for logging)
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				log.Printf("[MCP stderr] %s", scanner.Text())
			}
		}()

		// Read response from stdout
		scanner := bufio.NewScanner(stdout)
		var response []byte
		for scanner.Scan() {
			line := scanner.Bytes()
			// Skip empty lines
			if len(line) == 0 {
				continue
			}
			// Check if it's valid JSON
			var jsonTest interface{}
			if json.Unmarshal(line, &jsonTest) == nil {
				response = line
				break
			}
		}

		// Wait for process to finish
		cmd.Wait()

		if len(response) == 0 {
			http.Error(w, "No valid JSON response from MCP server", http.StatusInternalServerError)
			return
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("MCP HTTP bridge listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
