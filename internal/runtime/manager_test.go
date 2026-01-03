package runtime

import (
	"os"
	"testing"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/metrics"
	"go.uber.org/zap"
)

func TestReloadSuccessAndFailure(t *testing.T) {
	cfg1 := `server:
  listen: ":0"
  timeout_ms: 200

gateway_auth:
  enabled: false

metrics:
  enabled: false

routing:
  default_group: "primary"

failover:
  retry:
    enabled: false
    max_retries: 1
  passive_eject:
    enabled: false
    fail_threshold: 3
    base_eject_ms: 10000
    max_eject_ms: 60000

groups:
  - name: "primary"
    priority: 10
    allow: ["/"]
    deny: []
    lb:
      policy: "wrr"
    health:
      active:
        enabled: false
    upstreams:
      - url: "http://example.invalid"
        weight: 1
        access_key: "UP"
  - name: "backup"
    priority: 1
    allow: ["/"]
    deny: []
    lb:
      policy: "wrr"
    health:
      active:
        enabled: false
    upstreams:
      - url: "http://example.invalid"
        weight: 1
        access_key: "UP"
`
	path := writeTempConfig(t, cfg1)

	m := metrics.New()
	mgr, err := NewManager(path, m, zap.NewNop())
	if err != nil {
		t.Fatalf("manager init: %v", err)
	}

	first := mgr.Get()
	if first.DefaultGroup != "primary" {
		t.Fatalf("expected default group primary")
	}

	cfg2 := `server:
  listen: ":0"
  timeout_ms: 200

gateway_auth:
  enabled: false

metrics:
  enabled: false

routing:
  default_group: "backup"

failover:
  retry:
    enabled: false
    max_retries: 1
  passive_eject:
    enabled: false
    fail_threshold: 3
    base_eject_ms: 10000
    max_eject_ms: 60000

groups:
  - name: "primary"
    priority: 10
    allow: ["/"]
    deny: []
    lb:
      policy: "wrr"
    health:
      active:
        enabled: false
    upstreams:
      - url: "http://example.invalid"
        weight: 1
        access_key: "UP"
  - name: "backup"
    priority: 1
    allow: ["/"]
    deny: []
    lb:
      policy: "wrr"
    health:
      active:
        enabled: false
    upstreams:
      - url: "http://example.invalid"
        weight: 1
        access_key: "UP"
`
	if err := os.WriteFile(path, []byte(cfg2), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := mgr.Reload(); err != nil {
		t.Fatalf("reload success: %v", err)
	}
	second := mgr.Get()
	if second.DefaultGroup != "backup" {
		t.Fatalf("expected default group backup")
	}

	invalid := `server:
  listen: ":0"
  timeout_ms: 200

gateway_auth:
  enabled: false

metrics:
  enabled: false

routing:
  default_group: "missing"

groups:
  - name: "primary"
    priority: 10
    allow: ["/"]
    deny: []
    lb:
      policy: "wrr"
    health:
      active:
        enabled: false
    upstreams:
      - url: "http://example.invalid"
        weight: 1
        access_key: "UP"
`
	if err := os.WriteFile(path, []byte(invalid), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := mgr.Reload(); err == nil {
		t.Fatalf("expected reload failure")
	}
	third := mgr.Get()
	if third.DefaultGroup != "backup" {
		t.Fatalf("expected default group to remain backup")
	}
}

func TestReloadRejectsInvalidPprof(t *testing.T) {
	cfg := `server:
  listen: ":0"
  timeout_ms: 200

gateway_auth:
  enabled: false

metrics:
  enabled: false

routing:
  default_group: "primary"

failover:
  retry:
    enabled: false
    max_retries: 1
  passive_eject:
    enabled: false
    fail_threshold: 3
    base_eject_ms: 10000
    max_eject_ms: 60000

groups:
  - name: "primary"
    priority: 10
    allow: ["/"]
    deny: []
    lb:
      policy: "wrr"
    health:
      active:
        enabled: false
    upstreams:
      - url: "http://example.invalid"
        weight: 1
        access_key: "UP"
`
	path := writeTempConfig(t, cfg)

	m := metrics.New()
	mgr, err := NewManager(path, m, zap.NewNop())
	if err != nil {
		t.Fatalf("manager init: %v", err)
	}

	invalid := `server:
  listen: ":0"
  timeout_ms: 200

gateway_auth:
  enabled: false

metrics:
  enabled: false

pprof:
  enabled: true
  path: "debug/pprof"
  accesskey: ""

routing:
  default_group: "primary"

failover:
  retry:
    enabled: false
    max_retries: 1
  passive_eject:
    enabled: false
    fail_threshold: 3
    base_eject_ms: 10000
    max_eject_ms: 60000

groups:
  - name: "primary"
    priority: 10
    allow: ["/"]
    deny: []
    lb:
      policy: "wrr"
    health:
      active:
        enabled: false
    upstreams:
      - url: "http://example.invalid"
        weight: 1
        access_key: "UP"
`
	if err := os.WriteFile(path, []byte(invalid), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := mgr.Reload(); err == nil {
		t.Fatalf("expected reload failure")
	}
	if mgr.Get().DefaultGroup != "primary" {
		t.Fatalf("expected default group to remain primary")
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "cfg-*.yaml")
	if err != nil {
		t.Fatalf("tempfile: %v", err)
	}
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		t.Fatalf("write config: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close config: %v", err)
	}
	return file.Name()
}
