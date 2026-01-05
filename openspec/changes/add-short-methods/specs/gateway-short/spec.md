## MODIFIED Requirements
### Requirement: Short Subscription Redirects
The gateway SHALL support short subscription entries with a configured `method` of `301`, `302`, or `proxy` (default `301`).

#### Scenario: Internal short redirect to rsshub
- **WHEN** a request targets `/short/{name}` with `method=301` and the short target is `/rsshub/...`
- **THEN** the gateway responds with HTTP 301 and a Location header to the target with the original query string appended.

#### Scenario: External short redirect to https URL
- **WHEN** a request targets `/short/{name}` with `method=302` and the short target is a full `https://...` URL
- **THEN** the gateway responds with HTTP 302 and a Location header to the target with the original query string appended.

#### Scenario: Internal short proxy
- **WHEN** a request targets `/short/{name}` with `method=proxy` and the short target is `/rsshub/...` or `/upvote/...`
- **THEN** the gateway proxies the request through the normal routing pipeline as if the client requested the target path.

#### Scenario: External short proxy
- **WHEN** a request targets `/short/{name}` with `method=proxy` and the short target is a full `https://...` URL
- **THEN** the gateway proxies the request to the external target directly.

### Requirement: Query Passthrough for Short Entries
Short entries SHALL preserve query parameters in a method-specific way.

#### Scenario: Redirect preserves full query string
- **WHEN** a short request uses method `301` or `302`
- **THEN** the Location target includes the original query string unmodified.

#### Scenario: Internal proxy preserves key/code
- **WHEN** a short request uses method `proxy` to an internal target
- **THEN** the gateway preserves the original query string (including `key` and `code`) when routing to the target path.

#### Scenario: External proxy strips key/code
- **WHEN** a short request uses method `proxy` to an external `https://...` target
- **THEN** the gateway removes `key` and `code` from the query string before proxying.

### Requirement: Short Entry Auth Behavior
Short entries SHALL apply gateway auth based on method and target.

#### Scenario: Redirects bypass gateway auth
- **WHEN** a short request uses method `301` or `302`
- **THEN** the gateway does not enforce gateway auth for the short entry request.

#### Scenario: Internal proxy enforces gateway auth
- **WHEN** a short request uses method `proxy` to an internal target
- **THEN** the gateway validates gateway auth using the target path and query.

#### Scenario: External proxy bypasses gateway auth
- **WHEN** a short request uses method `proxy` to an external target
- **THEN** the gateway does not enforce gateway auth for the short entry request.

### Requirement: Short Configuration and Validation
The gateway SHALL validate short configuration when enabled.

#### Scenario: Valid short configuration
- **WHEN** `short.enabled` is true and `short.path` begins with `/`, and each entry has a supported target and method
- **THEN** configuration validation succeeds.

#### Scenario: Invalid short configuration
- **WHEN** `short.path` does not start with `/`, a target is not a supported format, or method is not one of `301`, `302`, `proxy`
- **THEN** configuration validation fails and reload is rejected.
