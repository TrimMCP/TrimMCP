.PHONY: build build-bridge build-all test test-short test-coverage test-jit run lint clean help test-server

# Build configuration
BINARY_NAME=slim
BUILD_DIR=./bin
MAIN_PATH=./cmd/slimmcp
BRIDGE_NAME=slim-stdio-bridge
BRIDGE_PATH=./cmd/slim-stdio-bridge
CONFIG?=./examples/dev.yaml

# Local test environment (use workspace-scoped cache/paths to avoid system restrictions)
TEST_CACHE_DIR?=$(CURDIR)/.gocache
TEST_GOPATH_DIR?=$(CURDIR)/.gopath
TEST_ENV=GOCACHE=$(TEST_CACHE_DIR) GOPATH=$(TEST_GOPATH_DIR)

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build the stdio→HTTP JIT bridge shim
build-bridge:
	@echo "Building $(BRIDGE_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BRIDGE_NAME) $(BRIDGE_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BRIDGE_NAME)"

# Build both binaries (proxy + bridge)
build-all: build build-bridge
	@echo "Built $(BINARY_NAME) and $(BRIDGE_NAME) into $(BUILD_DIR)"

# Run all tests
test:
	@echo "Running tests..."
	@$(TEST_ENV) go test -v -cover ./...


# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@$(TEST_ENV) go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"


# Run the proxy
run: build
	@echo "Starting SlimMCP proxy..."
	@$(BUILD_DIR)/$(BINARY_NAME) start --config $(CONFIG)

# Run linter
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠ golangci-lint not found. Run 'make install-tools' to install"; \
		exit 1; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -f *.prof
	@echo "Clean complete"

# Development tools

# Start mock test server
test-server:
	@echo "Starting mock MCP server..."
	@go run ./tests/tools/mockserver --port 3000

# Start test servers (HTTP, HTTPS, SSE) for transport testing
test-servers:
	@echo "Starting test servers (HTTP:3200, HTTPS:3201, SSE:3202)..."
	@go run ./tests/tools/test-servers

# Record VCR cassettes
record-cassettes:
	@echo "Recording VCR cassettes..."
	@GITHUB_PERSONAL_ACCESS_TOKEN="$${GITHUB_TOKEN}" go run ./tests/tools/record-mcp -env .env.test -config tests/recording-scenarios.json -output tests/cassettes

# Generic stdio→HTTP bridge for any MCP server
# MCP Transport Types: stdio (stdin/stdout), SSE (Server-Sent Events), HTTP
# This bridge wraps stdio-based servers with an HTTP endpoint for testing
mcp-bridge:
	@echo "Starting MCP stdio→HTTP bridge on port $(PORT)..."
	@go run ./tests/tools/mcp-stdio-http-bridge \
		-port $(PORT) \
		-cmd $(CMD) \
		-args "$(ARGS)" \
		-env "$(ENV)"

# Start GitHub MCP server bridge (stdio→HTTP)
github-bridge:
	@echo "Starting GitHub MCP server bridge on port 3100..."
	@$(MAKE) mcp-bridge \
		PORT=3100 \
		CMD=npx \
		ARGS="-y,@modelcontextprotocol/server-github" \
		ENV="GITHUB_PERSONAL_ACCESS_TOKEN=$${GITHUB_TOKEN}"

# Start Notion MCP server bridge (stdio→HTTP)
notion-bridge:
	@echo "Starting Notion MCP server bridge on port 3101..."
	@$(MAKE) mcp-bridge \
		PORT=3101 \
		CMD=npx \
		ARGS="-y,@modelcontextprotocol/server-notion" \
		ENV="NOTION_API_KEY=$${NOTION_TOKEN}"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./tests/benchmarks/

# Run performance tests
test-perf:
	@echo "Running performance tests..."
	@$(TEST_ENV) go test -v ./tests/benchmarks/

# Run specific performance test
test-perf-latency:
	@echo "Running latency benchmark..."
	@$(TEST_ENV) go test -v -run=LatencyBenchmark ./tests/benchmarks/

test-perf-throughput:
	@echo "Running throughput test..."
	@$(TEST_ENV) go test -v -run=ThroughputTest ./tests/benchmarks/

test-perf-savings:
	@echo "Running token savings test..."
	@$(TEST_ENV) go test -v -run=TokenSavingsTest ./tests/benchmarks/

test-perf-concurrent:
	@echo "Running concurrent requests test..."
	@$(TEST_ENV) go test -v -run=ConcurrentRequestsTest ./tests/benchmarks/

# Run performance tests with profiling
test-perf-profile:
	@echo "Running performance tests with CPU profiling..."
	@$(TEST_ENV) go test -bench=. -cpuprofile=cpu.prof ./tests/benchmarks/
	@echo "Profile saved to cpu.prof - View with: go tool pprof cpu.prof"

# Run short tests (skip intensive performance tests)
test-short:
	@echo "Running short tests (skipping performance tests)..."
	@$(TEST_ENV) go test -short -v ./...


# Build the JIT bridge and run only the JIT stdio bridge test
test-jit: build-bridge
	@echo "Running JIT stdio bridge test..."
	@PATH="$(abspath $(BUILD_DIR)):$$PATH" $(TEST_ENV) go test -v -run JITSpawnViaStdioBridge ./tests/transport


# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"

# Format code
fmt:
	@echo "Formatting code..."
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "⚠ goimports not found, using go fmt only (run 'make install-tools' to install)"; \
	fi
	@go fmt ./...

# Show help
help:
	@echo "SlimMCP Makefile Commands:"
	@echo ""
	@echo "Production:"
	@echo "  make build          - Build the proxy binary (slim)"
	@echo "  make build-bridge   - Build the stdio→HTTP bridge (slim-stdio-bridge)"
	@echo "  make build-all      - Build both binaries into ./bin"
	@echo "  make run            - Build and run the proxy (CONFIG=path/to/config.yaml)"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make test                  - Run all tests"
	@echo "  make test-short            - Run tests (skip intensive performance tests)"
	@echo "  make test-coverage         - Run tests with coverage report"
	@echo "  make test-perf             - Run all performance tests"
	@echo "  make test-perf-latency     - Run latency benchmark"
	@echo "  make test-perf-throughput  - Run throughput test"
	@echo "  make test-perf-savings     - Run token savings test"
	@echo "  make test-perf-concurrent  - Run concurrent requests test"
	@echo "  make test-perf-profile     - Run benchmarks with CPU profiling"
	@echo "  make bench                 - Run Go benchmarks"
	@echo "  make test-jit              - Run only the JIT stdio bridge test"
	@echo ""
	@echo "Development Tools:"
	@echo "  make test-server       - Start mock MCP server for testing (port 3000)"
	@echo "  make test-servers      - Start HTTP/HTTPS/SSE test servers (ports 3200-3202)"
	@echo "  make record-cassettes  - Record VCR cassettes for integration tests"
	@echo "  make mcp-bridge        - Generic stdio→HTTP bridge (PORT=3100 CMD=npx ARGS=... ENV=...)"
	@echo "  make github-bridge     - Start GitHub MCP server bridge (port 3100)"
	@echo "  make notion-bridge     - Start Notion MCP server bridge (port 3101)"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint           - Run golangci-lint"
	@echo "  make fmt            - Format code with goimports and go fmt"
	@echo "  make install-tools  - Install development tools"
	@echo ""
	@echo "  make help           - Show this help message"
