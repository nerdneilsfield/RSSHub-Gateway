## 1. Implementation
- [x] 1.1 Extend config schema/validation for cache providers (memory/redis), TTL, key mode, and reload polling settings.
- [x] 1.2 Implement cache interface and in-memory cache (TTL + size cap) with thread-safe access.
- [x] 1.3 Implement Redis cache provider (connect/retry policy, optional auth/db).
- [x] 1.4 Wire cache into proxy for GET 2xx/3xx responses; ensure key/code are stripped and hop-by-hop headers are excluded.
- [x] 1.5 Add auto-reload polling by file hash with cached running hash (memory/redis) and reuse existing reload validation.
- [x] 1.6 Extend Prometheus metrics for cache hit/miss and per-upstream/per-route success/failure counts.
- [x] 1.7 Add tests for cache behavior, reload polling, and new metrics; update README/config examples.
