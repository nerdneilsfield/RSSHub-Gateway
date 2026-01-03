## 1. Implementation
- [ ] 1.1 Define v0.1 config schema for server, routing, auth, health, failover, and metrics
- [ ] 1.2 Implement config load + validation (default group, unique names, URL scheme, key required, health and failover bounds)
- [ ] 1.3 Build runtime snapshot structure for handlers and background loops
- [ ] 1.4 Implement prefix router (allow/deny, longest prefix, priority, default)
- [ ] 1.5 Implement group-level LB (smooth WRR and hash(path))
- [ ] 1.6 Implement gateway auth (key/code) and upstream code injection (remove client key/code)
- [ ] 1.7 Implement proxy forwarding (header filtering, host rewrite, timeout handling)
- [ ] 1.8 Implement active health checks for upstreams
- [ ] 1.9 Implement passive eject, retry (GET/HEAD), and fallback chaining
- [ ] 1.10 Implement Prometheus metrics registry and /metrics access control
- [ ] 1.11 Implement JSON access logs and event logs
- [ ] 1.12 Implement SIGHUP reload with atomic swap and rollback on failure

## 2. Tests
- [ ] 2.1 Unit tests for router prefix selection rules
- [ ] 2.2 Unit tests for LB selection (WRR and hash stability)
- [ ] 2.3 Unit tests for gateway auth and upstream code injection
- [ ] 2.4 Unit tests for health transitions and passive eject backoff
- [ ] 2.5 Integration test for proxying with retry/fallback
- [ ] 2.6 Integration test for /metrics accesskey control
- [ ] 2.7 Integration test for SIGHUP reload success/failure

## 3. Docs
- [ ] 3.1 Add minimal config example to README
- [ ] 3.2 Document /metrics accesskey and JSON logging fields
- [ ] 3.3 Document SIGHUP reload usage
