package upstream

import (
	"net/url"
	"testing"
	"time"
)

func TestHealthTransitions(t *testing.T) {
	state := NewState(mustURLState("http://up"), 1, "k")
	if state.HealthFail(3) {
		t.Fatalf("expected not unhealthy yet")
	}
	if state.HealthFail(3) {
		t.Fatalf("expected not unhealthy yet")
	}
	if !state.HealthFail(3) {
		t.Fatalf("expected unhealthy on third failure")
	}
	if !state.HealthSuccess() {
		t.Fatalf("expected healthy transition")
	}
}

func TestPassiveEjectBackoff(t *testing.T) {
	state := NewState(mustURLState("http://up"), 1, "k")
	now := time.Now()
	base := 10 * time.Millisecond
	max := 40 * time.Millisecond

	if ejected, _ := state.RecordFailure(now, 2, base, max); ejected {
		t.Fatalf("expected no eject on first failure")
	}
	ejected, until := state.RecordFailure(now, 2, base, max)
	if !ejected {
		t.Fatalf("expected eject on second failure")
	}
	if until.Sub(now) < base {
		t.Fatalf("expected base eject duration")
	}

	now = now.Add(100 * time.Millisecond)
	state.RecordFailure(now, 2, base, max)
	ejected, until = state.RecordFailure(now, 2, base, max)
	if !ejected {
		t.Fatalf("expected eject on second failure")
	}
	if until.Sub(now) < 2*base {
		t.Fatalf("expected backoff to increase")
	}
}

func mustURLState(raw string) *url.URL {
	parsed, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return parsed
}
