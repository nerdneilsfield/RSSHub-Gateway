## ADDED Requirements
### Requirement: Service Prefix Routing for Multiple Backends
The gateway SHALL route requests to distinct backend types based on service prefixes `/rsshub/` and `/upvote/`, using the same gateway access key for authentication.

#### Scenario: RSSHub prefix uses rsshub backend
- **WHEN** a request path starts with `/rsshub/`
- **THEN** the gateway routes to a group configured with backend `rsshub`.
- **AND** gateway auth is validated with the global access key.

#### Scenario: Upvote prefix uses upvote backend
- **WHEN** a request path starts with `/upvote/`
- **THEN** the gateway routes to a group configured with backend `upvote`.
- **AND** gateway auth is validated with the same global access key.

### Requirement: Explicit Prefix Stripping Before Proxying
Groups MAY configure a `strip_prefix` value. When set, the gateway SHALL remove that literal prefix from the request path before proxying to upstreams. The resulting path SHALL begin with `/`, and an empty result SHALL be normalized to `/`.

#### Scenario: Strip `/rsshub` prefix
- **WHEN** a request path is `/rsshub/qdaily/column/59`
- **AND** the selected group has `strip_prefix: "/rsshub"`
- **THEN** the upstream request path is `/qdaily/column/59`.

#### Scenario: Strip `/upvote` prefix to root
- **WHEN** a request path is `/upvote/`
- **AND** the selected group has `strip_prefix: "/upvote"`
- **THEN** the upstream request path is `/`.

### Requirement: Backend-Specific Query Rewrite
The gateway SHALL remove client `key` and `code` query parameters before proxying.

#### Scenario: RSSHub backend injects upstream code
- **WHEN** the selected group backend is `rsshub`
- **AND** the upstream has an `access_key`
- **THEN** the gateway injects `code=md5(path + upstream_access_key)` into the upstream query.

#### Scenario: Upvote backend does not inject upstream code
- **WHEN** the selected group backend is `upvote`
- **THEN** the gateway forwards the remaining query parameters without adding a `code` parameter.

### Requirement: Upvote Upstream Auth Optional
Upvote backend upstreams SHALL NOT require an `access_key`.

#### Scenario: Upvote upstream without access_key
- **WHEN** a group uses backend `upvote` and an upstream omits `access_key`
- **THEN** config validation accepts it and the proxy forwards without injected code.

### Requirement: Backend Defaults
Groups SHALL default to backend `rsshub` to preserve existing behavior.

#### Scenario: Legacy config without backend
- **WHEN** a group omits the backend field
- **THEN** it is treated as `rsshub`.
