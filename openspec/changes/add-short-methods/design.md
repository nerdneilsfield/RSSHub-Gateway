## Context
Short subscriptions currently resolve to a 301 Location. We need more flexible behaviors to support temporary redirects and direct proxying.

## Goals / Non-Goals
- Goals:
  - Support per-entry methods: 301, 302, and proxy.
  - Preserve query passthrough semantics, with special handling for key/code in proxy mode.
  - Keep existing redirect compatibility by defaulting to 301.
- Non-Goals:
  - Supporting 307/308 or per-entry custom response headers.

## Decisions
- Entry schema: add `method` with values `301`, `302`, or `proxy` (default `301`).
- Redirects: preserve full original query string for internal targets; strip key/code for external targets.
- Proxy internal: if target starts with `/rsshub` or `/upvote`, treat as a normal gateway request to that path; keep key/code in query for auth.
- Proxy external: for `https://` targets, forward the request directly and remove key/code from query before proxying.
- Short requests with proxy method do not re-enter short resolution (avoid loops).

## Risks / Trade-offs
- Proxying external targets can expose gateway to upstream latency; reuse existing timeout behavior.
- Internal proxy uses gateway auth; clients must include key/code as they would for direct access.

## Migration Plan
- Existing entries without `method` continue to behave as 301.

## Open Questions
- None.
