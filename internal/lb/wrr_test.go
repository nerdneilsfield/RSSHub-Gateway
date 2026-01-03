package lb

import (
	"net/url"
	"testing"
	"time"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/upstream"
)

func TestWRRDistribution(t *testing.T) {
	u1 := upstream.NewState(mustURL("http://up1"), 3, "k1")
	u2 := upstream.NewState(mustURL("http://up2"), 1, "k2")
	p := NewWRRPicker([]*upstream.State{u1, u2})

	counts := map[*upstream.State]int{}
	for i := 0; i < 40; i++ {
		picked := p.Pick("/path", time.Now(), nil)
		if picked == nil {
			t.Fatalf("expected pick")
		}
		counts[picked]++
	}
	if counts[u1] <= counts[u2] {
		t.Fatalf("expected u1 to be picked more often; got u1=%d u2=%d", counts[u1], counts[u2])
	}
	if counts[u1] < 2*counts[u2] {
		t.Fatalf("expected weight skew; got u1=%d u2=%d", counts[u1], counts[u2])
	}
}

func mustURL(raw string) *url.URL {
	parsed, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return parsed
}
