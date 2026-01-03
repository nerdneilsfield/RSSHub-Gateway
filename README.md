# RSSHub-Gateway
Gateway for multi-instance RSSHub deployments.

## Quickstart
```bash
go build -o rsshub-gateway ./...
./rsshub-gateway serve -c config.example.yaml
```

## Docker
```bash
docker build -t rsshub-gateway:latest .
docker run --rm -p 8080:8080 rsshub-gateway:latest
```

To use a custom config, mount a file and update the command:
```bash
docker run --rm -p 8080:8080 \
  -v "$(pwd)/config.example.yaml:/app/config.yaml:ro" \
  rsshub-gateway:latest \
  /app/rsshub-gateway serve -c /app/config.yaml
```

## Configuration
See `config.example.yaml` for the full v0.1 schema.

Key behaviors:
- Gateway auth: `?key=<ACCESS_KEY>` or `?code=md5(path+ACCESS_KEY)`
- Upstream injection: client key/code is removed and upstream code is injected
- Routing: longest prefix wins, then priority, then config order
- Load balancing: `wrr` or `hash` per group

## Metrics
Prometheus endpoint: `GET /metrics?accesskey=<METRICS_ACCESS_KEY>`
- Invalid access key returns 403
- Metrics include request counts/latency, upstream health, retries, fallbacks, and reload totals

## Logging
JSON access logs include:
`method`, `path`, `group`, `upstream`, `status`, `duration_ms`, `retries`, `fallback_chain`, `err_type`, `err`

Event logs include:
`health_change`, `upstream_eject`, `config reload`

## Reload
Send SIGHUP to reload config without downtime:
```bash
kill -HUP <pid>
```
