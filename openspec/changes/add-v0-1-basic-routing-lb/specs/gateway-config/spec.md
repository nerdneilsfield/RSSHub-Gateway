## ADDED Requirements

### Requirement: Core config validation
The system SHALL validate configuration to ensure: default group exists, group
names are unique, upstream URLs are http/https, upstream access_key is present
for code injection, health intervals/timeouts are positive with retries >= 1,
and passive eject bounds satisfy base_eject <= max_eject.

#### Scenario: Invalid default group
- **GIVEN** a config where routing.default_group is not present in groups
- **WHEN** the config is loaded
- **THEN** validation fails

#### Scenario: Invalid eject bounds
- **GIVEN** a config with base_eject_ms greater than max_eject_ms
- **WHEN** the config is loaded
- **THEN** validation fails
