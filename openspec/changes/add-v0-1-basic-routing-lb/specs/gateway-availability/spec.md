## ADDED Requirements

### Requirement: Active health checks
The system SHALL actively probe each upstream at `/healthz` using configured
interval, timeout, and retries. Three consecutive failures SHALL mark the
upstream unhealthy; a successful probe SHALL restore healthy status.

#### Scenario: Health failure marks unhealthy
- **GIVEN** an upstream that fails health checks
- **WHEN** three consecutive probes fail
- **THEN** the upstream is marked unhealthy and excluded from selection

#### Scenario: Health check includes access key
- **GIVEN** an upstream configured with an access_key
- **WHEN** the gateway probes `/healthz`
- **THEN** the request includes `?key=<access_key>`

### Requirement: Passive eject on proxy failures
The system SHALL track proxy failures (connection errors, timeouts, or 5xx) and
eject an upstream after `fail_threshold` consecutive failures with exponential
backoff bounded by `max_eject`. 4xx responses SHALL NOT count toward eject.

#### Scenario: 4xx does not eject
- **GIVEN** an upstream responding with 4xx
- **WHEN** multiple 4xx responses occur
- **THEN** the upstream is not ejected

#### Scenario: Consecutive failures trigger eject
- **GIVEN** `fail_threshold` is 3 and `base_eject_ms` is configured
- **WHEN** three consecutive eligible failures occur
- **THEN** the upstream is ejected for the base duration

### Requirement: Retry and fallback
The system SHALL retry GET/HEAD requests at most once using a different upstream.
If the target group has no available upstreams, it SHALL attempt fallback groups
in order; if all groups fail, it SHALL return 502.

#### Scenario: Retry uses a different upstream
- **GIVEN** a GET request and a group with multiple upstreams
- **WHEN** the first upstream attempt fails
- **THEN** the retry targets a different upstream

#### Scenario: Fallback groups are used
- **GIVEN** a target group with no available upstreams and configured fallback groups
- **WHEN** a request is processed
- **THEN** the gateway attempts fallback groups in order
