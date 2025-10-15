# Performance Targets: MCP Response Optimizer Proxy

**Feature**: MCP Response Optimizer Proxy (SlimMCP)
**Date**: 2025-10-13
**Status**: ✅ Aligned across all specs

## Overview

This document defines the official performance targets for SlimMCP. All specs (spec.md, plan.md, tasks.md) have been aligned to these targets.

---

## Single Instance Performance

### Latency Targets

| Strategy | p95 Latency Target | Success Criteria |
|----------|-------------------|------------------|
| Minimal | <10ms | SC-003 |
| Smart | <15ms | SC-003 |
| Total (with downstream) | <200ms | Constraint |

**Measurement**: p95 latency measured from proxy receiving request to sending filtered response (excludes downstream server time for filtering latency only; total includes downstream).

### Throughput Targets

| Metric | Target | Hardware Spec | Success Criteria |
|--------|--------|---------------|------------------|
| Requests/second | **500+** | 2 vCPU, 4GB RAM | **SC-004** |
| Strategy | Any (minimal, smart, explicit, passthrough) | Standard cloud compute | SC-004 |

**Key Points**:
- **500+ req/sec** is the baseline for a **single proxy instance**
- Standard cloud compute = 2 vCPU, 4GB RAM (e.g., AWS t3.medium, GCP e2-medium)
- Must handle 500+ req/sec with smart strategy (most expensive filtering)

---

## Horizontal Scaling Performance

### Multi-Instance Targets

| Configuration | Target Throughput | Success Criteria |
|---------------|------------------|------------------|
| 3 instances | **1,500+ req/sec** | SC-009 |
| N instances | **N × 500 req/sec** | SC-009 |

**Key Points**:
- Proxy operates **statelessly** (SC-009)
- Linear scaling expected: 3 instances = 3 × 500 = 1,500+ req/sec
- Requires external load balancer
- No shared state between instances

**When Required**:
- Production deployments exceeding 500 req/sec
- High-availability deployments (multiple instances for redundancy)
- Zero-downtime deployments

---

## Token Reduction Targets

| Metric | Target | Success Criteria |
|--------|--------|------------------|
| Token savings | **70-90%** | SC-001 |
| Field inference accuracy | **85%+** | SC-002 |
| Response correctness | **100%** | SC-005 |

**Measurement**:
- Token savings = (originalTokens - filteredTokens) / originalTokens × 100
- Measured on realistic Notion/Slack/Google Drive responses
- All required fields for successful downstream operations preserved

---

## Operational Targets

### Availability

| Target | Downtime Allowed | Success Criteria |
|--------|-----------------|------------------|
| **99.9%** (three nines) | ~43 min/month | **SC-013** |

### Other Operational Metrics

| Metric | Target | Success Criteria |
|--------|--------|------------------|
| Health check response | <100ms | SC-006 |
| Metrics update latency | <1 second | SC-007 |
| Config reload time | <5 seconds | SC-010 |
| First-attempt success rate | 95%+ | SC-011 |

---

## Testing Requirements

### Performance Test Suite (Phase 13)

All performance tests must validate these targets:

#### T114: Latency Benchmark
```
✓ Minimal strategy: p95 < 10ms
✓ Smart strategy: p95 < 15ms
```

#### T115: Throughput Test
```
✓ Single instance: 500+ req/sec sustained
✓ 3 instances with load balancer: 1,500+ req/sec sustained
```

#### T116: Token Savings Test
```
✓ Notion responses: 70-90% reduction
✓ Slack responses: 70-90% reduction
✓ Google Drive responses: 70-90% reduction
```

#### T117: Concurrent Requests
```
✓ 100 concurrent requests
✓ All succeed
✓ No race conditions
✓ Maintains throughput target
```

---

## Performance Constraints

### Hard Limits

| Constraint | Value | Reason |
|-----------|-------|--------|
| Max response size | 10 MB | Prevent OOM |
| Max nesting depth | 10 levels | Filtering performance |
| Request timeout | 30 seconds | Default (configurable) |
| Connection pool size | 100 connections | Per server |
| Idle connections per host | 10 | HTTP connection reuse |

### Memory Profile

| Instance Type | Expected Usage | Max Workload |
|---------------|----------------|--------------|
| 2 vCPU, 4GB RAM | ~2GB baseline | 500 req/sec |
| 4 vCPU, 8GB RAM | ~4GB baseline | 1,000 req/sec |

**Notes**:
- Memory usage scales with response size during filtering
- Must handle 10MB responses without OOM
- Go GC tuning may be needed for high-throughput scenarios

---

## Performance Optimization Strategy

### Priority 1: Meet Baseline (500 req/sec)

1. Efficient JSON parsing and filtering
2. Minimize allocations in hot paths
3. Use connection pooling
4. Optimize regex patterns (cache compiled patterns)
5. Profile with `pprof` under load

### Priority 2: Horizontal Scaling (1,500+ req/sec)

1. Ensure truly stateless operation
2. No shared state between instances
3. Schema/capability caching per instance
4. Load balancer configuration
5. Test with 3+ instances

### Priority 3: Optimize Further

1. Consider streaming for large responses
2. Optimize smart analyzer inference
3. Parallel field filtering for large objects
4. Tune Go GC settings (`GOMAXPROCS`, `GOGC`)

---

## Measurement Methodology

### Load Testing Setup

```bash
# Single instance test
hey -n 10000 -c 50 -m POST \
  -H "Content-Type: application/json" \
  -d @request.json \
  http://localhost:8080/notion

# Multi-instance test (with load balancer)
hey -n 30000 -c 150 -m POST \
  -H "Content-Type: application/json" \
  -d @request.json \
  http://loadbalancer:8080/notion
```

### Metrics to Collect

1. **Latency**:
   - p50, p95, p99 filtering time
   - Total request time (including downstream)

2. **Throughput**:
   - Requests per second sustained
   - Success rate percentage

3. **Resource Usage**:
   - CPU utilization
   - Memory usage (RSS, heap)
   - Goroutine count

4. **Token Savings**:
   - Average savings percentage
   - Original vs. filtered size distribution

---

## Success Criteria Mapping

| Success Criteria | Performance Target | Where Tested |
|-----------------|-------------------|--------------|
| SC-001 | 70-90% token reduction | T116 |
| SC-002 | 85%+ inference accuracy | T113 |
| SC-003 | <15ms p95 latency (smart), <10ms (minimal) | T114 |
| SC-004 | **500+ req/sec per instance** | **T115** |
| SC-005 | 100% correctness | T116, T111 |
| SC-006 | Health check <100ms | Manual test |
| SC-007 | Metrics <1s update | Manual test |
| SC-008 | Strategy switching | T112 |
| SC-009 | **Horizontal scaling to 1,500+ req/sec (3 instances)** | **T115** |
| SC-010 | Config reload <5s | Phase 10 tests |
| SC-011 | 95%+ success rate | T117, T111 |
| SC-012 | Measurable savings | T116, metrics |
| SC-013 | 99.9% availability | Production monitoring |

---

## Performance Target Evolution

### v0.1 (MVP)
- Target: 300+ req/sec single instance
- Focus: Correctness over performance

### v0.5 (Beta)
- Target: 500+ req/sec single instance (SC-004)
- Target: 1,500+ req/sec with 3 instances (SC-009)
- Latency: <15ms p95 (SC-003)

### v1.0 (Production)
- All targets met
- Sustained under load for 24+ hours
- 99.9% availability demonstrated

---

## Revision History

| Date | Change | Reason |
|------|--------|--------|
| 2025-10-13 | Initial alignment | Fixed mismatch: tasks.md showed 500 req/s, plan.md showed 1000 req/s |
| 2025-10-13 | Clarified multi-instance scaling | SC-009 specifies 3+ instances, therefore 3 × 500 = 1,500+ req/sec |
| 2025-10-13 | Updated Go version to 1.25+ | Latest stable release as of Oct 2025 |

---

**Status**: ✅ All specs aligned
**Next Action**: Begin implementation with these performance targets as acceptance criteria
