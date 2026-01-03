## ADDED Requirements
### Requirement: Reload Preflight Validation
Before swapping runtime on SIGHUP, the gateway SHALL validate configuration comprehensively and reject invalid configs while keeping the current runtime active.

#### Scenario: Reject invalid routing and group references
- **WHEN** `routing.default_group` is missing, or any `fallback_groups` entry references an unknown group or itself
- **THEN** reload fails, logs the validation error, and increments `config_reload_total{result="fail"}`.

#### Scenario: Reject invalid upstream definitions
- **WHEN** any upstream has a non-http(s) URL, a non-positive weight, or a missing access_key while code injection is required
- **THEN** reload fails with error logging and no runtime swap.

#### Scenario: Reject invalid observability settings
- **WHEN** metrics or pprof are enabled without a non-empty accesskey, or their paths do not start with `/`
- **THEN** reload fails and reports validation errors.

#### Scenario: Reject empty groups
- **WHEN** any group has zero upstreams
- **THEN** reload fails and the prior runtime remains active.
