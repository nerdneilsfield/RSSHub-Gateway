## ADDED Requirements

### Requirement: Metrics endpoint access control
The system SHALL expose a Prometheus metrics endpoint and require
`?accesskey=<METRICS_ACCESS_KEY>` for access; invalid accesskey SHALL return 403.

#### Scenario: Metrics access is denied
- **GIVEN** a request to `/metrics` with an invalid accesskey
- **WHEN** the request is processed
- **THEN** the gateway returns 403

### Requirement: Prometheus metrics coverage
The system SHALL emit core metrics including request counts/latency, upstream
request counts, upstream health status, eject counts, retry counts, fallback
counts, and config reload result counts.

#### Scenario: Metrics include reload counters
- **GIVEN** the metrics endpoint is accessed with a valid accesskey
- **WHEN** the response is returned
- **THEN** it includes `rsshub_gateway_config_reload_total`

### Requirement: JSON logging
The system SHALL emit JSON access logs per request and event logs for health
changes, ejections, and reload events, including method/path/status and timing.

#### Scenario: Access log is emitted
- **GIVEN** a successful proxied request
- **WHEN** the request completes
- **THEN** a JSON access log is emitted with method, path, status, and duration
