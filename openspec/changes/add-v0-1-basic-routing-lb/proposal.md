## Why
We need a usable v0.1 gateway that supports routing, load balancing, basic
availability, and operability. This establishes a production-viable baseline
for multi-instance RSSHub deployments.

## What Changes
- Add a config-driven routing model with groups, allow/deny prefixes, priority, and default group.
- Implement prefix-based group selection (longest prefix wins, deny overrides).
- Implement group-level load balancing (smooth WRR or hash(path)).
- Add gateway auth and upstream code injection (remove client key/code).
- Add active health checks, passive eject, retry (GET/HEAD), and fallback groups.
- Add Prometheus metrics endpoint with access key and JSON access/event logs.
- Add SIGHUP config reload with atomic runtime swap and rollback on failure.
- Tighten config validation for routing, upstreams, health, and failover.
- Proxy requests with header filtering, host rewrite, and timeout handling.

## Impact
- Affected specs: gateway-routing, gateway-auth, gateway-availability, gateway-observability, gateway-reload, gateway-config (new capabilities)
- Affected code: config loader/validation, router, load balancer, proxy handler, auth, health, metrics, logging, reload, main wiring
