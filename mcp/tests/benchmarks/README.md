# SlimMCP Performance Benchmarks

This directory contains **non-functional requirement tests** for SlimMCP, focusing on performance, throughput, latency, and resource utilization.

## Purpose

These tests validate that SlimMCP meets the performance targets defined in `specs/001-trim/performance-targets.md`:

- **Latency**: p95 <10ms (minimal), p95 <15ms (smart)
- **Throughput**: 500+ req/sec per instance
- **Token Savings**: 70-90% reduction
- **Concurrency**: Handle 100+ concurrent requests
- **Horizontal Scaling**: Linear scaling to 1,500+ req/sec with 3 instances

## Test Files

### `performance_test.go`

Contains Phase 13 tests (T114-T117):

- **T114**: Latency benchmarks (minimal and explicit strategies)
- **T115**: Throughput test (500+ req/sec target)
- **T116**: Token savings test (70-90% target)
- **T117**: Concurrent request test (100 concurrent)

## Running Benchmarks

### Quick Benchmarks (Go benchmarks)

```bash
# Run all benchmarks
go test -bench=. ./tests/benchmarks/

# Run specific benchmark
go test -bench=BenchmarkProxyLatencyMinimal ./tests/benchmarks/

# With memory profiling
go test -bench=. -benchmem ./tests/benchmarks/

# Generate CPU profile
go test -bench=. -cpuprofile=cpu.prof ./tests/benchmarks/
```

### Performance Tests (T114-T117)

```bash
# Run all performance tests
go test -v ./tests/benchmarks/

# Run specific test
go test -v -run=T114 ./tests/benchmarks/

# Skip in short mode (performance tests are intensive)
go test -short ./tests/benchmarks/  # Will skip
```

### Load Testing (External Tools)

For production-grade load testing, use external tools:

```bash
# Using hey (install: go install github.com/rakyll/hey@latest)
hey -n 10000 -c 50 -m POST \
  -H "Content-Type: application/json" \
  -d @request.json \
  http://localhost:8080/notion

# Using k6 (install: brew install k6)
k6 run load-test.js
```

## Performance Targets

### Latency Targets

| Strategy | p95 Target | Test |
|----------|-----------|------|
| Minimal | <10ms | T114 |
| Explicit | <10ms | T114 |
| Smart | <15ms | T114 (Phase 8) |

### Throughput Targets

| Configuration | Target | Test |
|--------------|--------|------|
| Single instance | 500+ req/sec | T115 |
| 3 instances | 1,500+ req/sec | T115 |

### Token Savings Targets

| Response Type | Target | Test |
|--------------|--------|------|
| Notion | 70-90% | T116 |
| GitHub | 70-90% | T116 |
| Slack | 70-90% | T116 |

## Test Organization

**Benchmarks vs. Functional Tests:**

- **`tests/benchmarks/`**: Performance, throughput, latency, resource usage *(non-functional)*
- **`tests/integration/`**: Functional integration tests *(behavior)*
- **`tests/acceptance/`**: User story acceptance tests *(requirements)*
- **`tests/transport/`**: Transport layer tests *(protocol)*

## Continuous Performance Monitoring

Performance tests should be run:

1. **Pre-commit**: Quick benchmarks to catch regressions
2. **CI/CD**: Full performance suite on main branch
3. **Release**: Comprehensive load testing before tagging

## Interpreting Results

### T114 (Latency)

```
✅ PASS: p95 < 10ms (minimal), p95 < 15ms (smart)
⚠️  WARN: p95 between target and 2x target
❌ FAIL: p95 > 2x target
```

### T115 (Throughput)

```
✅ PASS: >=500 req/sec with 95%+ success rate
⚠️  WARN: 400-499 req/sec
❌ FAIL: <400 req/sec or <95% success rate
```

### T116 (Token Savings)

```
✅ PASS: 70-90% savings
⚠️  WARN: 60-69% or 91-95% savings
❌ FAIL: <60% or >95% (too aggressive)
```

### T117 (Concurrency)

```
✅ PASS: All 100 requests succeed
❌ FAIL: Any request fails or race condition detected
```

## Profiling

Generate profiling data:

```bash
# CPU profile
go test -bench=BenchmarkProxyLatency -cpuprofile=cpu.prof ./tests/benchmarks/
go tool pprof cpu.prof

# Memory profile
go test -bench=BenchmarkProxyLatency -memprofile=mem.prof ./tests/benchmarks/
go tool pprof mem.prof

# Block profile (goroutine contention)
go test -bench=BenchmarkProxyLatency -blockprofile=block.prof ./tests/benchmarks/
go tool pprof block.prof
```

## Next Steps

- [ ] Add load testing scripts (`hey`, `k6`)
- [ ] Add horizontal scaling tests (3+ instances)
- [ ] Add memory usage monitoring
- [ ] Add database/cache performance tests (Phase 9)
- [ ] Add smart analyzer benchmarks (Phase 8)
