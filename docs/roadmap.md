# SlimMCP Walking Skeleton Roadmap

## Phase 0: Foundation
**Goal:** Project setup and basic infrastructure

**Criteria:**
- Project structure created
- Build system working
- Basic tests passing
- CI/CD pipeline running

**User Stories:**
- As a developer, I can build the project
- As a developer, I can run tests
- As a developer, I can see test coverage

---

## Phase 1: Passthrough Proxy
**Goal:** Basic proxy that forwards requests unchanged

**Criteria:**
- Accept MCP requests
- Forward to downstream server
- Return unmodified responses
- Health check endpoint works

**User Stories:**
- As a user, I can send requests through SlimMCP
- As a user, I receive the same response as direct connection
- As an operator, I can check if SlimMCP is running

---

## Phase 2: Minimal Filtering
**Goal:** Basic field filtering (hardcoded rules)

**Criteria:**
- Filter responses to critical fields only (id, type, url, title)
- Add _meta with filtering stats
- Measure token reduction
- No broken responses

**User Stories:**
- As a user, I receive smaller responses
- As a user, I see token savings in _meta
- As a developer, I can verify filtering works

---

## Phase 3: Configuration
**Goal:** Configurable filtering strategies

**Criteria:**
- YAML configuration file
- Multiple filtering strategies (minimal, smart, passthrough)
- Per-server configuration
- Configuration validation

**User Stories:**
- As an operator, I can configure filtering strategy
- As an operator, I can configure multiple downstream servers
- As an operator, I get errors for invalid config

---

## Phase 4: Smart Inference
**Goal:** Intelligent field selection

**Criteria:**
- Parse request parameters for field hints
- Infer needed fields from context
- Apply tool-specific defaults
- Maintain >90% accuracy

**User Stories:**
- As a user, I get relevant fields without specifying them
- As a user, I don't get missing field errors
- As a developer, I can see inference logic in logs

---

## Phase 5: Explicit Selection
**Goal:** Client-controlled field selection

**Criteria:**
- Support _meta.fieldSelection in requests
- Handle include/exclude lists
- Support nested field paths
- Validate field selections

**User Stories:**
- As a developer, I can specify exactly which fields I need
- As a developer, I can exclude specific fields
- As a developer, I can request nested fields (properties.status)

---

## Phase 6: Monitoring & Metrics
**Goal:** Observability and metrics

**Criteria:**
- Prometheus metrics endpoint
- Token savings metrics
- Request/error counters
- Performance histograms

**User Stories:**
- As an operator, I can see token savings
- As an operator, I can monitor error rates
- As an operator, I can track performance

---

## Phase 7: Safety Features
**Goal:** Production reliability

**Criteria:**
- Circuit breaker implementation
- Fallback to full response on error
- Request timeouts
- Connection pooling

**User Stories:**
- As a user, I never get broken responses
- As an operator, I see circuit breaker status
- As an operator, the service handles downstream failures gracefully

---

## Phase 8: Schema Management
**Goal:** Schema-aware filtering

**Criteria:**
- Load tool schemas from MCP servers
- Validate field selections against schema
- Preserve required fields
- Handle field relationships

**User Stories:**
- As a user, required fields are always included
- As a developer, invalid field selections are caught early
- As a developer, related fields are included automatically

---

## Phase 9: Native Support Detection
**Goal:** Detect and use server-native filtering

**Criteria:**
- Query server capabilities
- Detect native fieldSelection support
- Pass through _meta when supported
- Cache capability results

**User Stories:**
- As a user, I get optimal performance when servers support native filtering
- As an operator, I can see which servers use native vs proxy filtering
- As a developer, I can disable native detection for testing

---

## Phase 10: Learning Mode
**Goal:** Adaptive filtering based on usage

**Criteria:**
- Track field usage patterns
- Learn from historical requests
- Adapt filtering based on patterns
- Improve accuracy over time

**User Stories:**
- As a user, filtering improves the longer I use SlimMCP
- As an operator, I can see learning statistics
- As an operator, I can enable/disable learning mode

---

## Phase 11: Performance Optimization
**Goal:** Production-grade performance

**Criteria:**
- <10ms filtering latency (p95)
- 1000+ req/s throughput
- <100MB memory usage
- Zero memory leaks

**User Stories:**
- As a user, I experience minimal latency overhead
- As an operator, I can run high-traffic workloads
- As an operator, the service is stable over time

---

## Phase 12: Production Hardening
**Goal:** Enterprise-ready deployment

**Criteria:**
- TLS support
- Structured logging (JSON)
- Graceful shutdown
- Health checks with dependency status

**User Stories:**
- As an operator, I can deploy with TLS
- As an operator, I can integrate logs with my logging system
- As an operator, I can perform zero-downtime deployments

---

## Milestone Summary

**MVP (Phases 0-3):** Basic passthrough proxy with minimal filtering
- Timeline: 2-3 weeks
- Deliverable: Working prototype, proves concept

**Alpha (Phases 4-6):** Smart filtering with monitoring
- Timeline: 3-4 weeks
- Deliverable: Usable for testing, has key features

**Beta (Phases 7-9):** Production features and reliability
- Timeline: 3-4 weeks
- Deliverable: Production-ready for early adopters

**v1.0 (Phases 10-12):** Complete feature set
- Timeline: 4-6 weeks
- Deliverable: Full production release

**Total: ~12-17 weeks (3-4 months)**
