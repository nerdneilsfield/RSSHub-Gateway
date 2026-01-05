## Why
The gateway needs response caching, automatic config reload, and more granular metrics to reduce upstream load, simplify operations, and improve visibility without manual SIGHUPs.

## What Changes
- Add response caching for GET requests (200s) with mutually exclusive providers: in-memory or Redis.
- Add automatic config reload by polling file hash and comparing to cached running hash; support enable/disable and interval.
- Expand Prometheus metrics with cache hit/miss and per-upstream / per-route success/failure counters.

## Impact
- Affected specs: gateway-cache, gateway-reload, gateway-observability.
- Affected code: config schema/validation, runtime manager/reload loop, proxy handler, metrics, new cache provider package(s), docs/tests.
