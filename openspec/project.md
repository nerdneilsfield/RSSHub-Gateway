# Project Context

## Purpose
rsshub-gateway is a single-entry HTTP gateway for multi-instance RSSHub deployments.
It keeps RSSHub route compatibility, supports prefix-based group routing, per-group
load balancing (smooth WRR or hash(path)), dual-layer auth (gateway + upstream code
injection), health checks + passive eject + retry/fallback, Prometheus metrics,
JSON logs, and SIGHUP hot reload.

## Tech Stack
- Go 1.22+
- Fiber + fasthttp
- Prometheus client (prometheus/client_golang)
- YAML config (gopkg.in/yaml.v3)
- JSON logger (zap or zerolog)

## Project Conventions

### Code Style
- Standard Go style; run gofmt.
- Prefer small, testable packages under internal/ with clear responsibilities.
- Log in JSON with structured fields for access and events.

### Architecture Patterns
- Module layout: cmd/rsshub-gateway/main.go; internal/{config,runtime,router,lb,upstream,health,proxy,metrics,logging}.
- Immutable Runtime snapshot swapped via atomic.Value on reload; old health loops stopped via stop channel.
- Router uses allow/deny prefix matching with longest-prefix wins, then priority, then config order.
- Group-level LB: smooth WRR or hash(path) (HRW recommended); must skip unhealthy/ejected upstreams.
- Proxy layer rewrites query: remove client key/code, inject upstream code; filter hop-by-hop headers; set Host to upstream.

### Testing Strategy
- Unit: router (longest prefix, deny, priority), auth (key/code), lb (wrr distribution, hash stability), passive eject.
- Integration (httptest/mock): code injection, health transitions, retry/fallback, SIGHUP reload.
- Must pass `go test ./...` and `go test -race ./...`.

### Git Workflow
- TBD: not specified in design docs (confirm branching and commit conventions).

## Domain Context
- Path: URL path only (no query), leading slash; used for md5(code).
- Group: upstream group with allow/deny prefixes, LB policy, fallback list.
- Upstream: RSSHub instance URL with per-upstream access_key.
- Gateway auth: global access_key; accepts ?key or ?code=md5(path+key).
- Upstream auth injection: gateway injects code=md5(path+upstream_key) after stripping client key/code.
- Active healthcheck: /healthz with interval/timeout/retries; 3 consecutive fails => unhealthy.
- Passive eject: count failures (conn/timeout/5xx) and eject with exponential backoff.

## Important Constraints
- Remove client key and code before proxying; never forward gateway key.
- Metrics endpoint requires ?accesskey=; otherwise 403.
- Retry only GET/HEAD, max 1, must switch upstream.
- If target group unavailable, try fallback groups; otherwise 502. Timeout returns 504.
- routing.default_group must exist; upstream URLs must be http/https; access_key required for default code mode.
- v0.1 excludes multi-tenant keys/ACL, admin API, caching, rate limiting, UI.

## External Dependencies
- RSSHub upstream instances with /healthz.
- Prometheus for scraping /metrics.
- Optional: Docker/K8s for deployment; SIGHUP support for reload.
