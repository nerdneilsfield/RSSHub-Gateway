package lb

import (
	"hash/fnv"
	"math"
	"time"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/upstream"
)

type HashPicker struct {
	items []*upstream.State
}

func NewHashPicker(items []*upstream.State) *HashPicker {
	return &HashPicker{items: items}
}

func (p *HashPicker) Pick(path string, now time.Time, avoid map[*upstream.State]struct{}) *upstream.State {
	var best *upstream.State
	bestScore := math.Inf(1)
	for _, item := range p.items {
		if avoid != nil {
			if _, skip := avoid[item]; skip {
				continue
			}
		}
		if !item.Available(now) {
			continue
		}
		score := rendezvousScore(path, item)
		if score < bestScore {
			bestScore = score
			best = item
		}
	}
	return best
}

func rendezvousScore(path string, item *upstream.State) float64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(path))
	_, _ = h.Write([]byte("|"))
	_, _ = h.Write([]byte(item.HostLabel))
	sum := h.Sum64()
	u := (float64(sum) + 1) / (float64(math.MaxUint64) + 1)
	return -math.Log(u) / float64(item.Weight)
}
