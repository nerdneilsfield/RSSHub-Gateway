## ADDED Requirements

### Requirement: SIGHUP config reload
The system SHALL reload configuration on SIGHUP by building a new runtime and
atomically swapping it in. On failure, it SHALL keep the previous runtime active.

#### Scenario: Reload failure rolls back
- **GIVEN** a running gateway with a valid configuration
- **WHEN** a SIGHUP reload is triggered with an invalid config
- **THEN** the gateway continues using the previous runtime

#### Scenario: Reload success swaps runtime
- **GIVEN** a running gateway with a valid configuration
- **WHEN** a SIGHUP reload is triggered with a valid config
- **THEN** new requests use the updated runtime
