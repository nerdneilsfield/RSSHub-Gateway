## ADDED Requirements

### Requirement: Gateway inbound auth
The system SHALL enforce gateway authentication using a global access key, accepting
either `?key=<ACCESS_KEY>` or `?code=md5(path+ACCESS_KEY)` where path excludes query.
Requests that fail auth SHALL return 403.

#### Scenario: Missing credentials
- **GIVEN** gateway auth is enabled
- **WHEN** a request is made without key or code
- **THEN** the gateway returns 403

#### Scenario: Valid key auth
- **GIVEN** gateway auth is enabled
- **WHEN** a request is made with `?key=<ACCESS_KEY>`
- **THEN** the gateway authorizes the request

#### Scenario: Valid code auth
- **GIVEN** gateway auth is enabled
- **WHEN** a request is made with `?code=md5(path+ACCESS_KEY)`
- **THEN** the gateway authorizes the request

### Requirement: Upstream auth injection
The system SHALL remove client `key` and `code` query parameters and inject an
upstream `code=md5(path+upstream_access_key)` before proxying.

#### Scenario: Client key is not forwarded
- **GIVEN** a client request containing `key` and `code` parameters
- **WHEN** the request is proxied to an upstream
- **THEN** the upstream query contains only the injected `code` value
