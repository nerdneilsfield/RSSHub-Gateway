## ADDED Requirements
### Requirement: Route Prefix Aggregation for Metrics
The gateway SHALL expose Prometheus request metrics aggregated by matched route prefix to avoid per-path cardinality.

#### Scenario: Request metrics include route_prefix
- **WHEN** a request is routed via the longest matched allow prefix
- **THEN** request counters and duration histograms include a `route_prefix` label set to that prefix.
- **AND** if no allow prefix matches and `routing.default_group` is used, `route_prefix` SHALL be `default`.

### Requirement: Pprof Debug Endpoints
When pprof is enabled, the gateway SHALL expose Go pprof endpoints under the configured path (default `/debug/pprof`) and require `pprof.accesskey` for access.

#### Scenario: Pprof access requires accesskey
- **WHEN** a request targets the pprof path with a valid `accesskey`
- **THEN** the pprof handler responds with diagnostic data.
- **AND** when the accesskey is missing or invalid, the gateway returns 403.

#### Scenario: Pprof endpoints bypass proxy routing
- **WHEN** a request hits the pprof path
- **THEN** it is handled locally and MUST NOT be forwarded to upstreams.
