package health

import (
	"net/url"
	"testing"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/upstream"
)

func TestBuildHealthURLAddsKey(t *testing.T) {
	base, err := url.Parse("http://example.com:1200")
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	state := upstream.NewState(base, 1, "ACCESS")
	got := buildHealthURL(state, "/healthz")
	if got != "http://example.com:1200/healthz?key=ACCESS" {
		t.Fatalf("unexpected url: %s", got)
	}
}
