## Why
MVP v0.2 needs better operational visibility and safer reload behavior to support production usage.

## What Changes
- Add route-prefix aggregated Prometheus metrics to reduce query noise and aid troubleshooting.
- Add authenticated pprof endpoints for runtime profiling.
- Strengthen SIGHUP reload validation to reject invalid configs before swapping runtime.

## Impact
- Affected specs: gateway-observability, gateway-reload
- Affected code: metrics, routing/proxy attribution, config validation, reload handler, docs
