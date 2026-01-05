## ADDED Requirements
### Requirement: Response Cache Providers
The gateway SHALL support a response cache with exactly one provider enabled at a time: `memory` or `redis`.

#### Scenario: Memory cache enabled
- **GIVEN** cache provider `memory` is enabled with TTL and size limits
- **WHEN** a GET request returns HTTP 200 from an upstream
- **THEN** the gateway stores the response in memory and serves subsequent GET requests from cache until TTL expiry or eviction.

#### Scenario: Redis cache enabled
- **GIVEN** cache provider `redis` is enabled with a valid connection configuration
- **WHEN** a GET request returns HTTP 200 from an upstream
- **THEN** the gateway stores the response in Redis with TTL and serves subsequent GET requests from cache until TTL expiry.

#### Scenario: Invalid provider configuration
- **GIVEN** both memory and redis providers are enabled or the provider value is invalid
- **WHEN** the gateway loads or reloads configuration
- **THEN** configuration validation fails and no runtime swap occurs.

### Requirement: Cache Scope and Key Normalization
The gateway SHALL cache only GET responses with HTTP 2xx and 3xx status codes and SHALL build cache keys from the upstream path plus normalized query string after removing gateway auth parameters (`key`, `code`).

#### Scenario: Cache key excludes gateway auth params
- **GIVEN** a GET request to `/rsshub/route?key=K&foo=1&code=C`
- **WHEN** the gateway builds the cache key
- **THEN** the key includes `/rsshub/route?foo=1` and excludes `key` and `code`.

#### Scenario: Non-GET or non-2xx/3xx responses are not cached
- **GIVEN** a POST request or a GET request that returns a non-2xx/3xx status
- **WHEN** the gateway processes the upstream response
- **THEN** it does not store the response in cache.
