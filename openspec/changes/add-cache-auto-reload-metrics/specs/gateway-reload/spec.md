## ADDED Requirements
### Requirement: Automatic config reload by file hash
The gateway SHALL support automatic config reload by polling the config file hash on a configurable interval and comparing it with the cached running hash.

#### Scenario: Polling disabled
- **GIVEN** auto-reload polling is disabled
- **WHEN** the config file changes
- **THEN** the gateway does not attempt an automatic reload.

#### Scenario: Hash change triggers reload
- **GIVEN** auto-reload polling is enabled with interval `T`
- **AND** the cached running hash differs from the current file hash
- **WHEN** the poller runs
- **THEN** the gateway performs a reload using existing validation and swaps runtime on success.

#### Scenario: Reload failure keeps prior hash
- **GIVEN** a config change that fails validation
- **WHEN** the poller attempts a reload
- **THEN** the gateway keeps the previous runtime and does not update the cached running hash.
