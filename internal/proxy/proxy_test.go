package proxy

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/metrics"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/runtime"
	"go.uber.org/zap"
)

func TestProxyFallbackAndInjection(t *testing.T) {
	upKey := "UPKEY"
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		expected := md5Hex(r.URL.Path + upKey)
		if r.URL.Query().Get("code") != expected {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer up.Close()

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
    enabled: true
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
    fallback_groups: ["backup"]
    health:
      active:
        enabled: false
    upstreams: []

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
      - url: "` + up.URL + `"
        weight: 1
        access_key: "` + upKey + `"
`
	path := writeTempConfig(t, cfg)

	m := metrics.New()
	mgr, err := runtime.NewManager(path, m, zap.NewNop())
	if err != nil {
		t.Fatalf("manager init: %v", err)
	}

	app := fiber.New()
	app.All("/*", New(mgr, m, zap.NewNop()).Serve)

	req := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
	}
}

func TestMetricsAccessKey(t *testing.T) {
	cfg := `server:
  listen: ":0"
  timeout_ms: 200

gateway_auth:
  enabled: false

metrics:
  enabled: true
  path: "/metrics"
  accesskey: "PROM"

routing:
  default_group: "public"

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
  - name: "public"
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
	mgr, err := runtime.NewManager(path, m, zap.NewNop())
	if err != nil {
		t.Fatalf("manager init: %v", err)
	}

	app := fiber.New()
	app.All("/*", New(mgr, m, zap.NewNop()).Serve)

	badReq := httptest.NewRequest(http.MethodGet, "http://localhost/metrics?accesskey=BAD", nil)
	badResp, err := app.Test(badReq)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	if badResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", badResp.StatusCode)
	}

	okReq := httptest.NewRequest(http.MethodGet, "http://localhost/metrics?accesskey=PROM", nil)
	okResp, err := app.Test(okReq)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	if okResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(okResp.Body)
		t.Fatalf("expected 200, got %d: %s", okResp.StatusCode, string(body))
	}
}

func TestGatewayKeyAuth(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer up.Close()

	cfg := `server:
  listen: ":0"
  timeout_ms: 200

gateway_auth:
  enabled: true
  access_key: "GATE"
  accept_key: true
  accept_code: false

metrics:
  enabled: false

routing:
  default_group: "public"

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
  - name: "public"
    priority: 10
    allow: ["/"]
    deny: []
    lb:
      policy: "wrr"
    health:
      active:
        enabled: false
    upstreams:
      - url: "` + up.URL + `"
        weight: 1
        access_key: "UP"
`
	path := writeTempConfig(t, cfg)

	m := metrics.New()
	mgr, err := runtime.NewManager(path, m, zap.NewNop())
	if err != nil {
		t.Fatalf("manager init: %v", err)
	}

	app := fiber.New()
	app.All("/*", New(mgr, m, zap.NewNop()).Serve)

	okReq := httptest.NewRequest(http.MethodGet, "http://localhost/path?key=GATE", nil)
	okResp, err := app.Test(okReq)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	if okResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(okResp.Body)
		t.Fatalf("expected 200, got %d: %s", okResp.StatusCode, string(body))
	}

	badReq := httptest.NewRequest(http.MethodGet, "http://localhost/path?key=BAD", nil)
	badResp, err := app.Test(badReq)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	if badResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", badResp.StatusCode)
	}
}

func TestProxyRetry(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("slow"))
	}))
	defer slow.Close()

	fast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fast"))
	}))
	defer fast.Close()

	cfg := `server:
  listen: ":0"
  timeout_ms: 10

gateway_auth:
  enabled: false

metrics:
  enabled: false

routing:
  default_group: "public"

failover:
  retry:
    enabled: true
    max_retries: 1
  passive_eject:
    enabled: false
    fail_threshold: 3
    base_eject_ms: 10000
    max_eject_ms: 60000

groups:
  - name: "public"
    priority: 10
    allow: ["/"]
    deny: []
    lb:
      policy: "wrr"
    health:
      active:
        enabled: false
    upstreams:
      - url: "` + slow.URL + `"
        weight: 1
        access_key: "SLOW"
      - url: "` + fast.URL + `"
        weight: 1
        access_key: "FAST"
`
	path := writeTempConfig(t, cfg)

	m := metrics.New()
	mgr, err := runtime.NewManager(path, m, zap.NewNop())
	if err != nil {
		t.Fatalf("manager init: %v", err)
	}

	app := fiber.New()
	app.All("/*", New(mgr, m, zap.NewNop()).Serve)

	req := httptest.NewRequest(http.MethodGet, "http://localhost/retry", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
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

func md5Hex(input string) string {
	sum := md5.Sum([]byte(input))
	return hex.EncodeToString(sum[:])
}
