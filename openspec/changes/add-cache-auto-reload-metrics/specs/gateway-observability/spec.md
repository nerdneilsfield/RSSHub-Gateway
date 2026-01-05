## ADDED Requirements
### Requirement: Upstream success and failure counters
The gateway SHALL expose Prometheus counters for upstream successes and failures, labeled by group and upstream.

#### Scenario: Upstream result counters reflect HTTP status
- **GIVEN** a request is proxied to an upstream
- **WHEN** the upstream responds
- **THEN** the gateway increments a success counter for HTTP 2xx responses and a failure counter for all other status codes.

### Requirement: Route-level success and failure counters
The gateway SHALL expose Prometheus counters for route-level successes and failures, labeled by route prefix and group.

#### Scenario: Route result counters reflect gateway response
- **GIVEN** a client request is handled by the gateway
- **WHEN** the response is returned to the client
- **THEN** the gateway increments a success counter for HTTP 2xx responses and a failure counter for all other status codes.

### Requirement: Cache hit and miss counters
The gateway SHALL expose Prometheus counters for cache hits and misses, labeled by cache provider.

#### Scenario: Cache hit increments
- **GIVEN** caching is enabled
- **WHEN** a GET request is served from cache
- **THEN** the gateway increments the cache hit counter.

#### Scenario: Cache miss increments
- **GIVEN** caching is enabled
- **WHEN** a GET request is not found in cache and is fetched from upstream
- **THEN** the gateway increments the cache miss counter.
