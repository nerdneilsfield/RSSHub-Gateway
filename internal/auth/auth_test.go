package auth

import (
	"crypto/md5"
	"encoding/hex"
	"testing"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/config"
	"github.com/valyala/fasthttp"
)

func TestValidateGatewayKey(t *testing.T) {
	cfg := config.GatewayAuthConfig{Enabled: true, AccessKey: "KEY", AcceptKey: true}
	args := fasthttp.Args{}
	args.Set("key", "KEY")
	if !ValidateGateway(cfg, "/path", &args) {
		t.Fatalf("expected key auth to pass")
	}
}

func TestValidateGatewayCode(t *testing.T) {
	cfg := config.GatewayAuthConfig{Enabled: true, AccessKey: "KEY", AcceptCode: true}
	args := fasthttp.Args{}
	code := md5HexTest("/path" + "KEY")
	args.Set("code", code)
	if !ValidateGateway(cfg, "/path", &args) {
		t.Fatalf("expected code auth to pass")
	}
}

func TestValidateGatewayBypassPath(t *testing.T) {
	cfg := config.GatewayAuthConfig{
		Enabled:     true,
		AccessKey:   "KEY",
		AcceptKey:   true,
		BypassPaths: []string{"/favicon.ico"},
	}
	args := fasthttp.Args{}
	if !ValidateGateway(cfg, "/favicon.ico", &args) {
		t.Fatalf("expected bypass path to pass")
	}
	if ValidateGateway(cfg, "/robots.txt", &args) {
		t.Fatalf("expected non-bypass path to fail without key")
	}
}

func TestRewriteUpstreamQueryInjectCode(t *testing.T) {
	args := fasthttp.Args{}
	args.Set("key", "BAD")
	args.Set("code", "BAD")
	args.Set("foo", "bar")

	out := RewriteUpstreamQuery(&args, "/abc", "UPKEY", true)
	if string(out.Peek("key")) != "" {
		t.Fatalf("expected key removed")
	}
	if string(out.Peek("code")) == "" {
		t.Fatalf("expected code injected")
	}
	if string(out.Peek("foo")) != "bar" {
		t.Fatalf("expected foo preserved")
	}
}

func TestRewriteUpstreamQueryNoCode(t *testing.T) {
	args := fasthttp.Args{}
	args.Set("key", "BAD")
	args.Set("code", "BAD")
	args.Set("foo", "bar")

	out := RewriteUpstreamQuery(&args, "/abc", "", false)
	if string(out.Peek("key")) != "" {
		t.Fatalf("expected key removed")
	}
	if string(out.Peek("code")) != "" {
		t.Fatalf("expected code not injected")
	}
	if string(out.Peek("foo")) != "bar" {
		t.Fatalf("expected foo preserved")
	}
}

func md5HexTest(input string) string {
	sum := md5.Sum([]byte(input))
	return hex.EncodeToString(sum[:])
}
