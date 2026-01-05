package cache

import (
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestBuildKeyStripsAuthParams(t *testing.T) {
	args := fasthttp.Args{}
	args.Add("foo", "2")
	args.Add("foo", "1")
	args.Set("bar", "a")
	args.Set("key", "K")
	args.Set("code", "C")

	key := BuildKey("/rsshub/route", &args)
	expected := "/rsshub/route?bar=a&foo=1&foo=2"
	if key != expected {
		t.Fatalf("expected %s, got %s", expected, key)
	}
}

func TestMemoryCacheSetGetExpire(t *testing.T) {
	store := NewMemoryStore(Options{Enabled: true, MaxItemBytes: 1024, MaxTotalBytes: 4096})
	entry := Entry{Status: 200, Body: []byte("ok")}
	if err := store.SetResponse("k1", entry, 20*time.Millisecond); err != nil {
		t.Fatalf("set response: %v", err)
	}
	if _, ok := store.GetResponse("k1"); !ok {
		t.Fatalf("expected cache hit")
	}
	time.Sleep(30 * time.Millisecond)
	if _, ok := store.GetResponse("k1"); ok {
		t.Fatalf("expected cache miss after expiry")
	}
}

func TestMemoryCacheEvictsBySize(t *testing.T) {
	store := NewMemoryStore(Options{Enabled: true, MaxItemBytes: 1024, MaxTotalBytes: 8})
	entryA := Entry{Status: 200, Body: []byte("aaaaa")}
	entryB := Entry{Status: 200, Body: []byte("bbbbb")}

	if err := store.SetResponse("a", entryA, time.Minute); err != nil {
		t.Fatalf("set a: %v", err)
	}
	if err := store.SetResponse("b", entryB, time.Minute); err != nil {
		t.Fatalf("set b: %v", err)
	}
	if _, ok := store.GetResponse("a"); ok {
		t.Fatalf("expected entry a to be evicted")
	}
	if _, ok := store.GetResponse("b"); !ok {
		t.Fatalf("expected entry b to remain")
	}
}
