package lb

import (
	"net/url"
	"testing"
	"time"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/upstream"
)

func TestHashStability(t *testing.T) {
	u1 := upstream.NewState(mustURLHash("http://up1"), 1, "k1")
	u2 := upstream.NewState(mustURLHash("http://up2"), 1, "k2")
	p := NewHashPicker([]*upstream.State{u1, u2})

	picked := p.Pick("/stable", time.Now(), nil)
	if picked == nil {
		t.Fatalf("expected pick")
	}
	for i := 0; i < 10; i++ {
		again := p.Pick("/stable", time.Now(), nil)
		if again != picked {
			t.Fatalf("expected stable pick")
		}
	}
}

func mustURLHash(raw string) *url.URL {
	parsed, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return parsed
}
