package transport

import (
    "context"
    "encoding/json"
    "os/exec"
    "testing"
    "time"

    "github.com/slimmcp/slimmcp/pkg/config"
    "github.com/slimmcp/slimmcp/pkg/proxy"
    "github.com/slimmcp/slimmcp/tests/transport/harness"
    "github.com/slimmcp/slimmcp/tests/util"
)

// TestJITSpawnViaStdioBridge verifies that a stdio-spawned shim (slim-stdio-bridge)
// can forward JSON-RPC to SlimMCP over HTTP, enabling copy/paste JIT configs in clients.
func TestJITSpawnViaStdioBridge(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Locate slim-stdio-bridge; skip test if not on PATH (developers can `go build -o slim-stdio-bridge ./cmd/slim-stdio-bridge`).
    shimPath, err := exec.LookPath("slim-stdio-bridge")
    if err != nil {
        t.Skip("slim-stdio-bridge not found in PATH; build with `cd mcp && go build -o slim-stdio-bridge ./cmd/slim-stdio-bridge` to run this test")
        return
    }

    // Start a mock downstream MCP server
    mock := util.NewMockMCPServer()
    go mock.Start(3110)
    defer mock.Stop()
    time.Sleep(100 * time.Millisecond)

    // Seed mock server with a verbose Notion-like response to test filtering
    notionResponse := util.CreateNotionPageResponse()
    if err := mock.SetResponse("notion-fetch", notionResponse); err != nil {
        t.Fatalf("failed to set mock response: %v", err)
    }

    // Configure SlimMCP proxy to route /notion -> mock server with minimal filtering
    cfg := &config.Config{
        Proxy: config.ProxyConfig{
            ListenAddress: "127.0.0.1:8110",
            Strategy:      "minimal",
            Servers: map[string]*config.ServerConfig{
                "notion": {URL: "http://127.0.0.1:3110", Strategy: "minimal"},
            },
        },
        Optimization: config.OptimizationConfig{
            AlwaysKeep:   []string{"id", "type", "url", "title"},
            AlwaysRemove: []string{"icon", "cover", "created_by", "last_edited_by"},
            Safety:       config.SafetyConfig{FallbackOnError: true},
        },
        Monitoring:   config.MonitoringConfig{Enabled: false, MetricsPort: 9199},
    }

    proxyServer := proxy.NewServer(cfg)
    go proxyServer.Start()
    defer proxyServer.Stop(context.Background())
    time.Sleep(200 * time.Millisecond)

    // Spawn the stdio shim pointing at SlimMCP
    tr := harness.NewStdioTransport(
        "jit-stdio-bridge",
        shimPath,
        []string{"-url", "http://127.0.0.1:8110/notion"},
        nil,
    )

    if err := tr.Start(ctx); err != nil {
        t.Fatalf("failed to start stdio bridge: %v", err)
    }
    defer tr.Stop()

    // Wait for shim to become healthy (initialize)
    deadline := time.Now().Add(10 * time.Second)
    healthy := false
    for time.Now().Before(deadline) {
        if tr.IsHealthy(ctx) {
            healthy = true
            break
        }
        time.Sleep(200 * time.Millisecond)
    }
    if !healthy {
        t.Fatalf("stdio bridge not healthy: %s", tr.Name())
    }

    // Run a couple of basic scenarios through the shim
    for _, sc := range harness.CreateBasicScenarios() {
        execCtx, c := context.WithTimeout(ctx, sc.Timeout)
        if sc.Timeout == 0 {
            execCtx, c = context.WithTimeout(ctx, 5*time.Second)
        }
        resp, err := tr.Execute(execCtx, sc.Request)
        c()
        if err != nil {
            t.Fatalf("scenario %s failed: %v", sc.Name, err)
        }
        if sc.ValidateResponse != nil {
            if err := sc.ValidateResponse(t, resp); err != nil {
                t.Fatalf("scenario %s validation failed: %v", sc.Name, err)
            }
        }
    }

    // Extra scenario: tools/call exercising proxy filtering via the shim
    callReq := harness.MustMarshal(harness.MCPRequest{
        JSONRPC: "2.0",
        ID:      99,
        Method:  "tools/call",
        Params: harness.MustMarshal(map[string]interface{}{
            "name": "notion-fetch",
            "arguments": map[string]interface{}{
                "id": "page-xyz",
            },
        }),
    })

    execCtx, c := context.WithTimeout(ctx, 8*time.Second)
    respRaw, err := tr.Execute(execCtx, callReq)
    c()
    if err != nil {
        t.Fatalf("tools/call via shim failed: %v", err)
        return
    }

    // Parse response
    var rpcResp harness.MCPResponse
    if err := json.Unmarshal(respRaw, &rpcResp); err != nil {
        t.Fatalf("failed to unmarshal response: %v", err)
    }
    if rpcResp.Error != nil {
        t.Fatalf("unexpected error in response: %+v", rpcResp.Error)
    }

    // The result should be a filtered object
    var result map[string]interface{}
    if err := json.Unmarshal(rpcResp.Result, &result); err != nil {
        t.Fatalf("failed to unmarshal result: %v", err)
    }
    // Critical fields present
    mustHave := []string{"id", "type", "url", "title"}
    for _, k := range mustHave {
        if _, ok := result[k]; !ok {
            t.Fatalf("missing expected field %q in filtered result", k)
        }
    }
    // Obvious heavy fields should be removed by minimal strategy
    if _, ok := result["properties"]; ok {
        t.Fatalf("expected 'properties' to be filtered out")
    }
    if _, ok := result["icon"]; ok {
        t.Fatalf("expected 'icon' to be filtered out")
    }

    // Now exercise explicit field selection via _meta
    // Inject _meta.fieldSelection with strategy explicit and include title
    // Build raw JSON to merge meta (simple approach here):
    explicitRaw := []byte(`{"jsonrpc":"2.0","id":100,"method":"tools/call","params":{"name":"notion-fetch","arguments":{"id":"page-explicit"}},"_meta":{"fieldSelection":{"strategy":"explicit","include":["title"]}}}`)

    execCtx2, c2 := context.WithTimeout(ctx, 8*time.Second)
    respRaw2, err := tr.Execute(execCtx2, explicitRaw)
    c2()
    if err != nil {
        t.Fatalf("explicit tools/call via shim failed: %v", err)
    }

    var rpcResp2 harness.MCPResponse
    if err := json.Unmarshal(respRaw2, &rpcResp2); err != nil {
        t.Fatalf("failed to unmarshal explicit response: %v", err)
    }
    if rpcResp2.Error != nil {
        t.Fatalf("unexpected error in explicit response: %+v", rpcResp2.Error)
    }

    var result2 map[string]interface{}
    if err := json.Unmarshal(rpcResp2.Result, &result2); err != nil {
        t.Fatalf("failed to unmarshal explicit result: %v", err)
    }

    // Expect id/type/url/title present, heavy fields trimmed
    for _, k := range []string{"id", "type", "url", "title"} {
        if _, ok := result2[k]; !ok {
            t.Fatalf("explicit: missing expected field %q", k)
        }
    }
    if _, ok := result2["properties"]; ok {
        t.Fatalf("explicit: expected 'properties' to be filtered out")
    }

    t.Logf("✓ JIT stdio bridge forwarded tools/call and proxy applied minimal+explicit filtering")
}
