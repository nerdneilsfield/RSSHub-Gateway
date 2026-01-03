## ADDED Requirements

### Requirement: Config-driven routing groups
The system SHALL load a v0.1 config defining routing groups with allow/deny prefixes,
priority, load-balancing policy, and upstream lists, plus a default group.

#### Scenario: Default group validation
- **GIVEN** a config where routing.default_group does not exist in groups
- **WHEN** the config is loaded
- **THEN** the system rejects the config with a validation error

### Requirement: Prefix-based group selection
The system SHALL select a routing group by prefix matching where deny overrides allow,
longest allow prefix wins, then higher priority, then config order, and default group
is used when no group matches.

#### Scenario: Longest prefix wins
- **GIVEN** group A allows `/telegram/` and group B allows `/telegram/private/`
- **WHEN** a request is made to `/telegram/private/x`
- **THEN** the system selects group B

### Requirement: Group-level load balancing
The system SHALL support per-group load balancing with either smooth WRR or hash(path)
policy to select one upstream from the group.

#### Scenario: WRR selection
- **GIVEN** a group with two upstreams weighted 3 and 1 using policy `wrr`
- **WHEN** multiple requests are routed to the group
- **THEN** the selection distribution favors the higher-weight upstream

#### Scenario: Hash selection stability
- **GIVEN** a group using policy `hash`
- **WHEN** repeated requests are made to the same path
- **THEN** the same upstream is selected, barring membership changes

### Requirement: Basic proxy forwarding
The system SHALL proxy requests to the selected upstream and return the upstream
response; if no upstream is available, it SHALL return 502.

#### Scenario: Forwarding succeeds
- **GIVEN** a request to `/qdaily/column/59` and a matching group with at least one upstream
- **WHEN** the request is processed
- **THEN** the gateway forwards the request to the selected upstream and returns its response

#### Scenario: No upstream available
- **GIVEN** a request routed to a group with zero upstreams configured
- **WHEN** the request is processed
- **THEN** the gateway returns 502

### Requirement: Proxy header and timeout handling
The system SHALL filter hop-by-hop headers, set the Host header to the upstream
host, and return 504 on upstream timeout.

#### Scenario: Hop-by-hop headers filtered
- **GIVEN** a request containing hop-by-hop headers like `Connection` or `Transfer-Encoding`
- **WHEN** the request is proxied
- **THEN** those headers are not forwarded to the upstream

#### Scenario: Timeout returns 504
- **GIVEN** an upstream that does not respond within the configured timeout
- **WHEN** the proxy request times out
- **THEN** the gateway returns 504
