## Context
We want to add response caching (memory or Redis), automatic config reload by file hash, and finer-grained Prometheus metrics. This introduces a new storage dependency (Redis) and new background polling behavior.

## Goals / Non-Goals
- Goals:
  - Reduce upstream load by caching GET 200 responses.
  - Enable hands-free config reloads via file hash polling with existing validation.
  - Improve observability with cache, route, and upstream success/failure counters.
- Non-Goals:
  - Caching non-GET requests.
  - Cross-instance cache consistency guarantees beyond Redis best-effort.
  - UI or admin API for cache management.

## Decisions
- Cache providers: mutually exclusive `memory` or `redis`; Redis optional and config-driven.
- Cache key: normalize by upstream path plus query string after removing gateway auth params (key/code). Query normalization order must be stable (required for Upvote feeds).
- Cache scope: cache only 2xx and 3xx responses by default; do not cache streaming or oversized bodies.
- Defaults: TTL 1h (3600000ms), max item 2MiB, max total size 50MiB, reload poll interval 30s.
- In-memory cache: bounded size with TTL; eviction policy is LRU.
- Redis cache: store response payload with TTL; key namespace `rsshub_gateway:cache:{hash}`.
- Auto-reload: poll file hash on interval; compare to cached running hash (memory/redis); reload only on change; update cached hash only after successful reload.

## Risks / Trade-offs
- Stale cache data may mask upstream changes; TTLs should be conservative.
- Redis outages should log an error and fall back to memory cache.
- Polling reload may trigger frequent validation; avoid too small intervals.

## Migration Plan
- Default cache disabled; auto-reload polling disabled by default.
- Existing SIGHUP reload remains supported.

## Open Questions
- None.
