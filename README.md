# RSSHub-Gateway

[![GoReleaser](https://github.com/nerdneilsfield/RSSHub-Gateway/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/nerdneilsfield/RSSHub-Gateway/actions/workflows/goreleaser.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/nerdneilsfield/RSSHub-Gateway)](go.mod)
[![Release](https://img.shields.io/github/v/release/nerdneilsfield/RSSHub-Gateway?include_prereleases)](https://github.com/nerdneilsfield/RSSHub-Gateway/releases)
[![License](https://img.shields.io/github/license/nerdneilsfield/RSSHub-Gateway)](LICENSE)

One gateway, many feeds. Keep RSSHub and Upvote RSS stable under load with routing,
auth, health checks, and first-class observability, while staying RSSHub-compatible.

At a glance:
- Go + Fiber + fasthttp, optimized for low-latency proxying
- Gateway key/code auth with safe RSSHub code injection
- Active health checks, passive eject, retry, and fallback
- Prometheus metrics, pprof, JSON logs, and hot reload

[中文说明](README_zh.md)

## Highlights
- Multi-backend routing: `/rsshub/` for RSSHub, `/upvote/` for Upvote RSS
- Prefix-based grouping with longest-match selection and per-group LB
- Short subscriptions: `/short/{name}` 301 redirect with query passthrough
- Homepage: `/` renders README.md (use `/?lang=zh` or `/zh` for Chinese)
- Gateway auth: `?key=` or `?code=md5(path+key)` + RSSHub code injection
- Active health checks + passive eject + retry + fallback
- Prometheus metrics, pprof, JSON access/event logs
- SIGHUP config reload with rollback on failure
- Response cache (memory/redis) + auto reload polling

## Typical Use Cases
- One stable URL for multiple RSSHub clusters
- Fast failover across upstreams without changing feed URLs
- Short, memorable subscription links for feed readers

## Architecture

Request flow stays simple: authenticate, route by prefix, pick upstream, proxy.

```mermaid
flowchart LR
    Client -->|HTTP| Gateway
    Gateway -->|/short/*| Short
    Short -->|301| Target
    Gateway --> Router
    Router --> Group
    Group --> LB
    LB --> RSSHub
    LB --> Upvote
    Gateway -->|/metrics| Prometheus
    Gateway -->|logs| Logger
```

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway
    participant R as Router
    participant LB as Load Balancer
    participant U as Upstream

    C->>G: GET /short/latepost?key=...
    alt short hit
        G->>G: Resolve short + passthrough query
        G-->>C: 301 Location: /rsshub/latepost/4?key=...
    else proxy
        C->>G: GET /rsshub/path?key=...
        G->>G: Validate key/code
        G->>R: Select group by prefix
        R-->>G: Group name
        G->>LB: Pick upstream
        LB-->>G: Upstream
        G->>G: Remove key/code, inject upstream code (rsshub)
        G->>U: Proxy request
        U-->>G: Response
        G-->>C: Response
    end
```

## Try It in 60 Seconds

Build locally and run with the sample config:

```bash
# build
make build

# run
./rsshub-gateway serve -c config.example.yaml
```

## Docker

```bash
docker build -t rsshub-gateway:latest .
docker run --rm -p 8080:8080 rsshub-gateway:latest
```

Prebuilt images:
- `docker pull nerdneils/rsshub-gateway:latest`
- `docker pull ghcr.io/nerdneilsfield/rsshub-gateway:latest`

To use a custom config:

```bash
docker run --rm -p 8080:8080 \
  -v "$(pwd)/config.example.yaml:/app/config.yaml:ro" \
  rsshub-gateway:latest \
  /app/rsshub-gateway serve -c /app/config.yaml
```

## Configuration

The full schema lives in `config.example.yaml`.

<details>
<summary>How to write config (details)</summary>

- `routing.default_group` must match a group name.
- Prefix rules use `allow`/`deny` with leading `/`; deny overrides allow.
- `backend` must be `rsshub` or `upvote` (defaults to `rsshub`).
- `strip_prefix` removes service prefixes like `/rsshub` or `/upvote` before proxying.
- `gateway_auth` needs `access_key` and at least one of `accept_key`/`accept_code`.
- `gateway_auth.bypass_paths` skips gateway auth for exact paths (still proxied/code-injected).
- Upstream `access_key` is required for RSSHub code injection and healthcheck `?key=`.
- Health check uses `path`, `interval_ms`, `timeout_ms`, `retries`.
- `failover.passive_eject` requires `base_eject_ms <= max_eject_ms`.
- Metrics require `metrics.accesskey` when enabled.
- Pprof requires `pprof.accesskey` when enabled.
- Cache requires provider (`memory` or `redis`) with TTL and size limits; caches GET 2xx/3xx only.
- Auto reload polling uses config hash compare (`reload.auto.enabled` + `interval_ms`).
- Short requires `short.path` (starts with `/`) and unique entry names when enabled.
</details>

<details>
<summary>Full config example</summary>

```yaml
server:
  listen: ":8080"
  timeout_ms: 8000

gateway_auth:
  enabled: true
  access_key: "ILoveRSSHub"
  accept_key: true
  accept_code: true
  bypass_paths:
    - "/favicon.ico"
    - "/logo.png"
    - "/robots.txt"
    - "/manifest.json"

metrics:
  enabled: true
  path: "/metrics"
  accesskey: "PROM_KEY_123"

pprof:
  enabled: false
  path: "/debug/pprof"
  accesskey: "PPROF_KEY_123"

cache:
  enabled: false
  provider: "memory"
  ttl_ms: 3600000
  max_item_bytes: 2097152
  max_total_bytes: 52428800
  redis:
    addr: "127.0.0.1:6379"
    password: ""
    db: 0
    dial_timeout_ms: 1000
    read_timeout_ms: 1000
    write_timeout_ms: 1000
    key_prefix: "rsshub_gateway"

reload:
  auto:
    enabled: false
    interval_ms: 30000

short:
  enabled: true
  path: "/short"
  entries:
    - name: "latepost"
      target: "/rsshub/latepost/4"
    - name: "reddit-top"
      target: "https://example.com/rss?platform=reddit"

routing:
  default_group: "rsshub-public"

failover:
  retry:
    enabled: true
    max_retries: 1
  passive_eject:
    enabled: true
    fail_threshold: 3
    base_eject_ms: 10000
    max_eject_ms: 60000

groups:
  - name: "rsshub-public"
    backend: "rsshub"
    strip_prefix: "/rsshub"
    priority: 10
    allow: ["/rsshub/"]
    deny: []
    lb:
      policy: "wrr"
    fallback_groups: ["rsshub-backup"]
    health:
      active:
        enabled: true
        path: "/healthz"
        interval_ms: 30000
        timeout_ms: 10000
        retries: 3
    upstreams:
      - url: "http://rsshub-1:1200"
        weight: 3
        access_key: "UP1KEY"
      - url: "http://rsshub-2:1200"
        weight: 2
        access_key: "UP2KEY"

  - name: "rsshub-backup"
    backend: "rsshub"
    strip_prefix: "/rsshub"
    priority: 1
    allow: ["/rsshub/"]
    deny: []
    lb:
      policy: "hash"
    upstreams:
      - url: "http://rsshub-b1:1200"
        weight: 1
        access_key: "B1KEY"

  - name: "upvote"
    backend: "upvote"
    strip_prefix: "/upvote"
    priority: 5
    allow: ["/upvote/"]
    deny: []
    lb:
      policy: "wrr"
    upstreams:
      - url: "http://upvote-rss:80"
        weight: 1
```
</details>

## Authentication

Gateway access supports both key and code styles:

```text
http://127.0.0.1:8080/rsshub/latepost/4?key=ACCESS_KEY
http://127.0.0.1:8080/rsshub/latepost/4?code=md5(path+ACCESS_KEY)
http://127.0.0.1:8080/upvote/?platform=reddit&key=ACCESS_KEY
```

Migration note (code auth):
If you previously used `?code=` with `/latepost/...`, update the path to `/rsshub/latepost/...`
and compute `md5("/rsshub/latepost/4"+ACCESS_KEY)`. Key-based access is unchanged.

## Short Subscriptions

Short entries return a 301 redirect and append the original query string to the target.
If the target already contains query parameters, the short query is appended with `&`.

```text
GET /short/latepost?key=ACCESS_KEY
-> 301 Location: /rsshub/latepost/4?key=ACCESS_KEY

GET /short/reddit-top?code=ABC
-> 301 Location: https://example.com/rss?platform=reddit&code=ABC
```

Upstream injection rules:
- Remove client `key` and `code`
- Inject `code=md5(path+upstream_access_key)` for RSSHub only

## Routing and Load Balancing

- Match allow/deny by prefix, deny overrides allow
- Choose longest prefix, then higher priority, then config order
- Use `wrr` or `hash` per group
- Strip service prefix (like `/rsshub` or `/upvote`) before proxying when configured

## Health Checks

Active health checks call `/healthz` on each upstream. If the upstream requires
an `ACCESS_KEY`, the gateway automatically appends `?key=<upstream_access_key>`.

If you use Docker Compose, update healthcheck like this:

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:1200/healthz?key=${ACCESS_KEY}"]
```

## Metrics

Metrics endpoint (requires access key):

```text
GET /metrics?accesskey=<METRICS_ACCESS_KEY>
```

<details>
<summary>Metrics list</summary>

- rsshub_gateway_requests_total{method,group,route_prefix,status}
- rsshub_gateway_request_duration_seconds_bucket{group,route_prefix}
- rsshub_gateway_upstream_requests_total{group,upstream,status}
- rsshub_gateway_upstream_health{group,upstream}
- rsshub_gateway_upstream_eject_total{group,upstream}
- rsshub_gateway_retry_total{group}
- rsshub_gateway_fallback_total{from,to}
- rsshub_gateway_config_reload_total{result}
</details>

## Pprof

Pprof endpoint (requires access key):

```text
GET /debug/pprof/?accesskey=<PPROF_ACCESS_KEY>
```

## Logging

Access logs are JSON per request. Event logs include health changes, ejections,
and reload outcomes.

<details>
<summary>Suggested access log fields</summary>

- ts
- level
- req_id
- method
- path
- group
- upstream
- route_prefix
- status
- duration_ms
- retries
- fallback_chain
- err_type
- err
</details>

## Reload

Send SIGHUP to reload config without downtime:

```bash
kill -HUP <pid>
```

## Development

```bash
make test
make cover
```

## Release

```bash
goreleaser release --snapshot --clean --skip-publish
```

## License

MIT
