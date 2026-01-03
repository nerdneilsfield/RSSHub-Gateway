## ADDED Requirements
### Requirement: Short Subscription Redirects
The gateway SHALL support short subscription entries that redirect to configured targets using HTTP 301.

#### Scenario: Internal short to rsshub
- **WHEN** a request targets `/short/{name}` and the short target is `/rsshub/...`
- **THEN** the gateway responds with HTTP 301 and a `Location` header pointing to the target path.

#### Scenario: Internal short to upvote
- **WHEN** a request targets `/short/{name}` and the short target is `/upvote/...`
- **THEN** the gateway responds with HTTP 301 and a `Location` header pointing to the target path.

#### Scenario: External short to https URL
- **WHEN** a request targets `/short/{name}` and the short target is a full `https://...` URL
- **THEN** the gateway responds with HTTP 301 and a `Location` header pointing to the external URL.

### Requirement: Query Passthrough for Short Redirects
Short redirects SHALL preserve the original request query string without modification.

#### Scenario: Preserve key and code parameters
- **WHEN** a short request includes `key`, `code`, or other query parameters
- **THEN** the gateway appends the raw query string to the target URL without filtering.

### Requirement: Short Redirects Bypass Gateway Auth
Short redirect requests SHALL NOT be validated by gateway auth; the target endpoint handles auth.

#### Scenario: Short works without gateway key
- **WHEN** a request targets `/short/{name}` without gateway `key` or `code`
- **THEN** the gateway still issues the redirect.

### Requirement: Short Configuration and Validation
The gateway SHALL validate short configuration when enabled.

#### Scenario: Valid short configuration
- **WHEN** `short.enabled` is true and `short.path` begins with `/`
- **AND** each entry has a unique, non-empty name
- **AND** each target is either `/rsshub/...`, `/upvote/...`, or `https://...`
- **THEN** the config is accepted.

#### Scenario: Invalid short configuration
- **WHEN** `short.path` does not start with `/` or a target is not a supported format
- **THEN** configuration validation fails.
