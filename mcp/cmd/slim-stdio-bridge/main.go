package main

import (
    "bufio"
    "bytes"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "os/exec"
    "strings"
)

// slim-stdio-bridge reads JSON-RPC messages from stdin (Claude/Cursor style),
// forwards them via HTTP POST to a SlimMCP endpoint, and writes responses to stdout.
//
// Usage examples:
//   slim-stdio-bridge -url http://localhost:8080/notion
//   SLIM_URL=http://localhost:8080/notion slim-stdio-bridge
//   SLIM_AUTOSTART="slimmcp start --config examples/dev.yaml" slim-stdio-bridge -url http://localhost:8080/notion
//
// Note: This is a convenience shim to enable JIT spawning via MCP clients that
// speak stdio only. SlimMCP itself remains an HTTP proxy.
func main() {
    var urlFlag string
    flag.StringVar(&urlFlag, "url", "", "SlimMCP HTTP endpoint (e.g., http://localhost:8080/notion)")
    flag.Parse()

    baseURL := strings.TrimSpace(urlFlag)
    if baseURL == "" {
        baseURL = strings.TrimSpace(os.Getenv("SLIM_URL"))
    }
    if baseURL == "" {
        log.Fatalf("missing -url (or SLIM_URL) for SlimMCP endpoint")
    }

    // Optional: autostart SlimMCP if requested
    if cmd := strings.TrimSpace(os.Getenv("SLIM_AUTOSTART")); cmd != "" {
        go func(command string) {
            // Run via shell for convenience so users can pass full commands
            c := exec.Command("bash", "-lc", command)
            c.Stdout = os.Stderr
            c.Stderr = os.Stderr
            if err := c.Start(); err != nil {
                log.Printf("failed to autostart SlimMCP: %v", err)
                return
            }
            log.Printf("autostarted SlimMCP: %s (pid=%d)", command, c.Process.Pid)
        }(cmd)
    }

    client := &http.Client{}
    in := bufio.NewScanner(os.Stdin)
    out := bufio.NewWriter(os.Stdout)
    defer out.Flush()

    for in.Scan() {
        line := strings.TrimSpace(in.Text())
        if line == "" {
            continue
        }

        reqBody := []byte(line)
        resp, err := client.Post(baseURL, "application/json", bytes.NewReader(reqBody))
        if err != nil {
            // Return a minimal JSON-RPC error response to client
            fmt.Fprintf(out, "{\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32000,\"message\":%q}}\n", "bridge POST error: "+err.Error())
            out.Flush()
            continue
        }

        body, err := io.ReadAll(resp.Body)
        resp.Body.Close()
        if err != nil {
            fmt.Fprintf(out, "{\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32001,\"message\":%q}}\n", "bridge read error: "+err.Error())
            out.Flush()
            continue
        }

        // Write raw response back to stdout (single-line)
        trimmed := strings.TrimSpace(string(body))
        if trimmed == "" {
            fmt.Fprintln(out, "{\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32002,\"message\":\"empty response from SlimMCP\"}}")
        } else {
            fmt.Fprintln(out, trimmed)
        }
        out.Flush()
    }

    if err := in.Err(); err != nil {
        // Best-effort final error log to stderr
        fmt.Fprintf(os.Stderr, "stdin scan error: %v\n", err)
    }
}

